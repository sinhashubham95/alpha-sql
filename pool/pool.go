package pool

import (
	"context"
	"errors"
	alphasql "github.com/sinhashubham95/alpha-sql"
	"golang.org/x/sync/semaphore"
	"sync"
	"sync/atomic"
	"time"
)

type pool struct {
	// mu is the pool internal lock. Any modification of shared state of
	// the pool (but Acquires of acquireSem) must be performed only by
	// holder of the lock. Long-running operations are not allowed when mux
	// is held.
	mu sync.Mutex
	// acquireSem provides an allowance to acquire a resource.
	// Releases are allowed only when caller holds mux. Acquires have to
	// happen before mux is locked (doesn't apply to semaphore.TryAcquire in
	// AcquireAllIdle).
	acquireSem *semaphore.Weighted
	destructWG sync.WaitGroup

	allConnections  []*Connection
	idleConnections *mvStack

	maxSize int32

	constructor func(ctx context.Context) (*alphasql.Connection, error)
	destructor  func(ctx context.Context, c *alphasql.Connection) error

	acquireCount         int64
	acquireDuration      time.Duration
	emptyAcquireCount    int64
	canceledAcquireCount atomic.Int64

	resetCount int

	baseAcquireCtx       context.Context
	cancelBaseAcquireCtx context.CancelFunc
	closed               bool
}

// Pool is used to manage the set of connections, also being able to reuse it.
type Pool struct {
	// 64 bit fields accessed with atomics must be at beginning of struct to guarantee alignment for certain 32-bit
	// architectures. See BUGS section of https://pkg.go.dev/sync/atomic and https://github.com/jackc/pgx/issues/1288.
	newConnectionsCount  atomic.Int64
	lifetimeDestroyCount atomic.Int64
	idleDestroyCount     atomic.Int64

	p                           *pool
	db                          *alphasql.DB
	config                      *Config
	beforeConnect               func(context.Context, *alphasql.ConnectionConfig) error
	afterConnect                func(context.Context, *alphasql.Connection) error
	beforeAcquire               func(context.Context, *Connection) bool
	afterRelease                func(context.Context, *Connection) bool
	beforeClose                 func(context.Context, *alphasql.Connection)
	minConnections              int32
	maxConnections              int32
	maxConnectionLifetime       time.Duration
	maxConnectionLifetimeJitter time.Duration
	maxConnectionIdleTime       time.Duration
	healthCheckPeriod           time.Duration

	healthCheckChan chan struct{}

	closeOnce sync.Once
	closeChan chan struct{}
}

// New is used to create a new pool
func New(ctx context.Context, cfg *Config) (*Pool, error) {
	err := cfg.ValidateAndDefault()
	if err != nil {
		return nil, err
	}
	db, err := alphasql.Open(ctx, cfg.ConnectionConfig)
	if err != nil {
		return nil, err
	}
	pp := &Pool{
		config:                      cfg,
		db:                          db,
		beforeConnect:               cfg.BeforeConnect,
		afterConnect:                cfg.AfterConnect,
		beforeAcquire:               cfg.BeforeAcquire,
		afterRelease:                cfg.AfterRelease,
		beforeClose:                 cfg.BeforeClose,
		minConnections:              cfg.MinConnections,
		maxConnections:              cfg.MaxConnections,
		maxConnectionLifetime:       cfg.MaxConnectionLifetime,
		maxConnectionLifetimeJitter: cfg.MaxConnectionLifetimeJitter,
		maxConnectionIdleTime:       cfg.MaxConnectionIdleTime,
		healthCheckPeriod:           cfg.HealthCheckPeriod,
		healthCheckChan:             make(chan struct{}, 1),
		closeChan:                   make(chan struct{}),
	}
	p := newPool(ctx, pp)
	pp.p = p
	go pp.warmup(ctx)
	return pp, nil
}

// Close closes all connections in the pool and rejects future Acquire calls. Blocks until all connections are returned.
func (p *Pool) Close(ctx context.Context) {
	p.closeOnce.Do(func() {
		close(p.closeChan)
		p.p.close(ctx)
	})
}

func newPool(ctx context.Context, p *Pool) *pool {
	baseAcquireCtx, cancelBaseAcquireCtx := context.WithCancel(ctx)
	return &pool{
		acquireSem:           semaphore.NewWeighted(int64(p.maxConnections)),
		idleConnections:      newMVStack(),
		allConnections:       make([]*Connection, 0),
		maxSize:              p.maxConnections,
		constructor:          p.constructor,
		destructor:           p.destructor,
		baseAcquireCtx:       baseAcquireCtx,
		cancelBaseAcquireCtx: cancelBaseAcquireCtx,
	}
}

func (p *pool) createConnection(ctx context.Context, maxConnectionLifetime, maxConnectionLifetimeJitter time.Duration) error {
	if !p.acquireSem.TryAcquire(1) {
		return alphasql.ErrPoolSpaceNotAvailable
	}
	p.mu.Lock()
	if p.closed {
		p.acquireSem.Release(1)
		p.mu.Unlock()
		return alphasql.ErrPoolClosed
	}
	if len(p.allConnections) >= int(p.maxSize) {
		p.acquireSem.Release(1)
		p.mu.Unlock()
		return alphasql.ErrPoolSpaceNotAvailable
	}
	c := p.newConnection(maxConnectionLifetime, maxConnectionLifetimeJitter)
	p.mu.Unlock()
	value, err := p.constructor(ctx)
	p.mu.Lock()
	defer p.mu.Unlock()
	defer p.acquireSem.Release(1)
	if err != nil {
		removeFromConnections(&p.allConnections, c)
		p.destructWG.Done()
		return err
	}
	c.c = value
	c.status = connectionStatusIdle
	// If closed while constructing resource then destroy it and return an error
	if p.closed {
		_ = p.destructor(ctx, value)
		return alphasql.ErrPoolClosed
	}
	p.idleConnections.push(c)
	return nil
}

func (p *Pool) createIdleConnection(ctx context.Context, errs chan error) {
	err := p.p.createConnection(ctx, p.maxConnectionLifetime, p.maxConnectionLifetimeJitter)
	if errors.Is(err, alphasql.ErrPoolSpaceNotAvailable) {
		err = nil
	}
	errs <- err
}

func (p *Pool) createIdleConnections(ctx context.Context, count int) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errs := make(chan error, count)
	for i := 0; i < count; i++ {
		go p.createIdleConnection(ctx, errs)
	}
	var firstError error
	for i := 0; i < count; i++ {
		err := <-errs
		if err != nil && firstError == nil {
			cancel()
			firstError = err
		}
	}
	return firstError
}

func (p *pool) getTotalConnections() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.allConnections)
}

func (p *Pool) handleExpiryIdlenessForConnections(ctx context.Context) bool {
	destroyed := false
	total := p.p.getTotalConnections()
	idleConnections := p.p.acquireAllIdleConnections()
	for _, c := range idleConnections {
		if p.isExpiredConnection(c) && total >= int(p.minConnections) {
			p.lifetimeDestroyCount.Add(1)
			go p.p.destroyAcquiredConnection(ctx, c)
			total--
			destroyed = true
		} else if c.idleDuration() > p.maxConnectionIdleTime && total > int(p.minConnections) {
			p.idleDestroyCount.Add(1)
			go p.p.destroyAcquiredConnection(ctx, c)
			total--
			destroyed = true
		} else {
			p.p.releaseUnused(ctx, c)
		}
	}
	return destroyed
}

func (p *Pool) checkHealthForConnections(ctx context.Context) {
	for {
		if err := p.createIdleConnections(ctx, int(p.minConnections)-p.p.getTotalConnections()); err != nil {
			break
		}
		if destroyed := p.handleExpiryIdlenessForConnections(ctx); !destroyed {
			break
		}
		select {
		case <-p.closeChan:
			return
		case <-time.After(500 * time.Millisecond):
		}
	}
}

func (p *Pool) healthChecker(ctx context.Context) {
	ticker := time.NewTicker(p.healthCheckPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-p.closeChan:
			return
		case <-p.healthCheckChan:
			p.checkHealthForConnections(ctx)
		case <-ticker.C:
			p.checkHealthForConnections(ctx)
		}
	}
}

func (p *Pool) forceTriggerHealthCheck() {
	go func() {
		time.Sleep(500 * time.Millisecond)
		select {
		case p.healthCheckChan <- struct{}{}:
		default:
		}
	}()
}

func (p *Pool) warmup(ctx context.Context) {
	_ = p.createIdleConnections(ctx, int(p.minConnections))
	p.healthChecker(ctx)
}

func (p *pool) close(ctx context.Context) {
	defer p.destructWG.Wait()
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}
	p.closed = true
	p.cancelBaseAcquireCtx()

	for c, ok := p.idleConnections.pop(); ok; {
		removeFromConnections(&p.allConnections, c)
		go p.destroyConnection(ctx, c)
	}
}

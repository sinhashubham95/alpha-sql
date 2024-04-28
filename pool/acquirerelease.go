package pool

import (
	"context"
	alphasql "github.com/sinhashubham95/alpha-sql"
	"github.com/sinhashubham95/go-utils/maths"
	"time"
)

// Acquire is used to return a (*Connection) from the pool.
func (p *Pool) Acquire(ctx context.Context) (*Connection, error) {
	for {
		c, err := p.acquire(ctx)
		if err != nil {
			return nil, err
		}
		if c.idleDuration() > time.Second {
			err = c.Ping(ctx)
			if err != nil {
				go p.p.destroyAcquiredConnection(ctx, c)
				continue
			}
		}
		if p.beforeAcquire(ctx, c) {
			return c, nil
		}
		go p.p.destroyAcquiredConnection(ctx, c)
	}
}

func (p *Pool) Release(ctx context.Context, c *Connection) {
	if c.status != connectionStatusAcquired {
		return
	}
	if p.isExpiredConnection(c) {
		p.lifetimeDestroyCount.Add(1)
		go p.p.destroyAcquiredConnection(ctx, c)
		p.forceTriggerHealthCheck()
		return
	}
	go func() {
		if p.afterRelease(ctx, c) {
			p.p.release(ctx, c, time.Now().UnixNano())
		} else {
			p.lifetimeDestroyCount.Add(1)
			p.p.destroyAcquiredConnection(ctx, c)
			p.forceTriggerHealthCheck()
		}
	}()
}

// acquireSemAll tries to acquire entire count.
// if not available, it exponentially acquires count.
func (p *pool) acquireSemAll(count int) int {
	if p.acquireSem.TryAcquire(int64(count)) {
		return count
	}
	var acquired int
	for i := int(maths.Log2(float32(count))); i >= 0; i-- {
		v := 1 << i
		if p.acquireSem.TryAcquire(int64(v)) {
			acquired += v
		}
	}
	return acquired
}

func (p *pool) acquireAllIdleConnections() []*Connection {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	numberOfIdleConnections := p.idleConnections.length()
	if numberOfIdleConnections == 0 {
		return nil
	}

	acquiredSem := p.acquireSemAll(numberOfIdleConnections)

	idle := make([]*Connection, acquiredSem)
	for i := range idle {
		c, _ := p.idleConnections.pop()
		c.status = connectionStatusAcquired
		idle[i] = c
	}

	p.idleConnections.bump()

	return idle
}

func (p *pool) tryAcquireIdleConnection() *Connection {
	c, ok := p.idleConnections.pop()
	if !ok {
		return nil
	}
	c.status = connectionStatusAcquired
	return c
}

func (p *pool) initialiseAcquiredConnection(ctx context.Context, c *Connection) (*Connection, error) {
	errCh := make(chan error)
	go func() {
		cc, err := p.constructor(ctx)
		if err != nil {
			p.mu.Lock()
			removeFromConnections(&p.allConnections, c)
			p.destructWG.Done()
			p.acquireSem.Release(1)
			p.mu.Unlock()
			select {
			case <-ctx.Done():
			case errCh <- err:
			}
			return
		}
		p.mu.Lock()
		c.c = cc
		c.status = connectionStatusAcquired
		p.mu.Unlock()
		select {
		case errCh <- nil:
		case <-ctx.Done():
			p.releaseUnused(ctx, c)
		}
	}()

	select {
	case <-ctx.Done():
		p.canceledAcquireCount.Add(1)
		return nil, ctx.Err()
	case err := <-errCh:
		if err != nil {
			return nil, err
		}
		return c, nil
	}
}

func (p *pool) acquireConnection(ctx context.Context, maxConnectionLifetime,
	maxConnectionLifetimeJitter time.Duration) (*Connection, error) {
	st := time.Now().UnixNano()

	var waitedForLock bool
	if !p.acquireSem.TryAcquire(1) {
		waitedForLock = true
		err := p.acquireSem.Acquire(ctx, 1)
		if err != nil {
			p.canceledAcquireCount.Add(1)
			return nil, err
		}
	}

	p.mu.Lock()
	if p.closed {
		p.acquireSem.Release(1)
		p.mu.Unlock()
		return nil, alphasql.ErrPoolClosed
	}

	// try to get the connection from the pool itself.
	if c := p.tryAcquireIdleConnection(); c != nil {
		if waitedForLock {
			p.emptyAcquireCount += 1
		}
		p.acquireCount += 1
		p.acquireDuration += time.Duration(time.Now().UnixNano() - st)
		p.mu.Unlock()
		return c, nil
	}

	c := p.newConnection(maxConnectionLifetime, maxConnectionLifetimeJitter)
	p.mu.Unlock()

	c, err := p.initialiseAcquiredConnection(ctx, c)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.emptyAcquireCount += 1
	p.acquireCount += 1
	p.acquireDuration += time.Duration(time.Now().UnixNano() - st)

	return c, nil
}

func (p *Pool) acquire(ctx context.Context) (*Connection, error) {
	select {
	case <-ctx.Done():
		p.p.canceledAcquireCount.Add(1)
		return nil, ctx.Err()
	default:
	}
	return p.p.acquireConnection(ctx, p.maxConnectionLifetime, p.maxConnectionLifetimeJitter)
}

func (p *pool) releaseUnused(ctx context.Context, c *Connection) {
	p.release(ctx, c, c.lastUsedNano)
}

func (p *pool) release(ctx context.Context, c *Connection, ts int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	defer p.acquireSem.Release(1)
	if p.closed {
		removeFromConnections(&p.allConnections, c)
		go p.destroyConnection(ctx, c)
	} else {
		c.status = connectionStatusIdle
		c.lastUsedNano = ts
		p.idleConnections.push(c)
	}
}

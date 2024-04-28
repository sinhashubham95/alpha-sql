package pool

import (
	"context"
	alphasql "github.com/sinhashubham95/alpha-sql"
	"math/rand/v2"
	"time"
)

// connection statuses
const (
	connectionStatusInitialising = 0
	connectionStatusIdle         = iota
	connectionStatusAcquired     = iota
)

// Connection is used for managing the instance of a Connection in the pool.
type Connection struct {
	c            *alphasql.Connection
	creationTime time.Time
	maxAgeTime   time.Time
	lastUsedNano int64
	status       byte
}

// Ping verifies a Connection to the database is still alive,
// establishing a Connection if necessary.
func (c *Connection) Ping(ctx context.Context) error {
	return c.c.Ping(ctx)
}

// Query executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
func (c *Connection) Query(ctx context.Context, query string, args ...any) (alphasql.Rows, error) {
	return c.c.Query(ctx, query, args...)
}

// QueryRow executes a query that is expected to return at most one row.
// QueryRow always returns a non-nil value. Errors are deferred until
// [alphasql.Row]'s Scan method is called.
// If the query selects no rows, the [alphasql.Row.Scan] will return [alphasql.ErrNoRows].
// Otherwise, [alphasql.Row.Scan] scans the first selected row and discards
// the rest.
func (c *Connection) QueryRow(ctx context.Context, query string, args ...any) alphasql.Row {
	return c.c.QueryRow(ctx, query, args...)
}

// Exec executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func (c *Connection) Exec(ctx context.Context, query string, args ...any) (alphasql.Result, error) {
	return c.c.Exec(ctx, query, args...)
}

// Prepare creates a prepared statement for later queries or executions.
// Multiple queries or executions may be run concurrently from the
// returned statement.
// The caller must call the statement's [alphasql.Statement.Close] method
// when the statement is no longer needed.
func (c *Connection) Prepare(ctx context.Context, query string) (alphasql.Statement, error) {
	return c.c.Prepare(ctx, query)
}

// BeginTX starts a transaction.
//
// The provided context is used until the transaction is committed or rolled back.
// If the context is canceled, the `alphasql` package will roll back
// the transaction. [TX.Commit] will return an error if the context provided to
// BeginTX is canceled.
//
// The provided [TXOptions] is optional and may be nil if defaults should be used.
// If a non-default isolation level is used that the driver doesn't support,
// an error will be returned.
func (c *Connection) BeginTX(ctx context.Context, options *alphasql.TXOptions) (alphasql.TX, error) {
	return c.c.BeginTX(ctx, options)
}

func (p *pool) newConnection(maxConnectionLifetime, maxConnectionLifetimeJitter time.Duration) *Connection {
	jitterSeconds := rand.Float64() * maxConnectionLifetimeJitter.Seconds()
	c := &Connection{
		creationTime: time.Now(),
		maxAgeTime:   time.Now().Add(maxConnectionLifetime).Add(time.Duration(jitterSeconds) * time.Second),
		lastUsedNano: time.Now().UnixNano(),
		status:       connectionStatusInitialising,
	}
	p.allConnections = append(p.allConnections, c)
	p.destructWG.Add(1)
	return c
}

func (p *Pool) constructor(ctx context.Context) (*alphasql.Connection, error) {
	p.newConnectionsCount.Add(1)
	cfg := p.config.ConnectionConfig.Copy()
	err := p.beforeConnect(ctx, cfg)
	if err != nil {
		return nil, err
	}
	c, err := p.db.Connect(ctx)
	if err != nil {
		return nil, err
	}
	err = p.afterConnect(ctx, c)
	if err != nil {
		// here we have to close the Connection because of error from after connect
		_ = c.Close()
		return nil, err
	}
	return c, nil
}

func (p *Pool) destructor(ctx context.Context, c *alphasql.Connection) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	p.beforeClose(ctx, c)
	return c.Close()
}

func (p *Pool) isExpiredConnection(c *Connection) bool {
	return time.Now().After(c.maxAgeTime)
}

func (p *pool) destroyConnection(ctx context.Context, c *Connection) {
	defer p.destructWG.Done()
	_ = p.destructor(ctx, c.c)
}

func (p *pool) destroyAcquiredConnection(ctx context.Context, c *Connection) {
	p.destroyConnection(ctx, c)
	p.mu.Lock()
	defer p.mu.Unlock()
	defer p.acquireSem.Release(1)
	removeFromConnections(&p.allConnections, c)
}

func (c *Connection) idleDuration() time.Duration {
	return time.Duration(time.Now().UnixNano() - c.lastUsedNano)
}

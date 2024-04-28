package pool

import (
	"context"
	alphasql "github.com/sinhashubham95/alpha-sql"
)

// Ping verifies a Connection to the database is still alive,
// establishing a Connection if necessary.
func (p *Pool) Ping(ctx context.Context) error {
	c, err := p.Acquire(ctx)
	if err != nil {
		return err
	}
	defer p.Release(ctx, c)
	return c.Ping(ctx)
}

// Query executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
func (p *Pool) Query(ctx context.Context, query string, args ...any) (alphasql.Rows, error) {
	c, err := p.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	r, err := c.Query(ctx, query, args...)
	if err != nil {
		p.Release(ctx, c)
		return p.getPoolErrRows(err), err
	}
	return p.getPoolRows(c, r), nil
}

// QueryRow executes a query that is expected to return at most one row.
// QueryRow always returns a non-nil value. Errors are deferred until
// [alphasql.Row]'s Scan method is called.
// If the query selects no rows, the [alphasql.Row.Scan] will return [alphasql.ErrNoRows].
// Otherwise, [alphasql.Row.Scan] scans the first selected row and discards
// the rest.
func (p *Pool) QueryRow(ctx context.Context, query string, args ...any) alphasql.Row {
	c, err := p.Acquire(ctx)
	if err != nil {
		return p.getPoolErrRow(err)
	}
	r := c.QueryRow(ctx, query, args...)
	return p.getPoolRow(c, r)
}

// Exec executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func (p *Pool) Exec(ctx context.Context, query string, args ...any) (alphasql.Result, error) {
	c, err := p.Acquire(ctx)
	if err != nil {
		return p.getPoolErrResult(err), err
	}
	defer p.Release(ctx, c)
	return c.Exec(ctx, query, args...)
}

// Prepare creates a prepared statement for later queries or executions.
// Multiple queries or executions may be run concurrently from the
// returned statement.
// The caller must call the statement's [alphasql.Statement.Close] method
// when the statement is no longer needed.
func (p *Pool) Prepare(ctx context.Context, query string) (alphasql.Statement, error) {
	c, err := p.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer p.Release(ctx, c)
	return c.Prepare(ctx, query)
}

// BeginTX starts a transaction.
//
// The provided context is used until the transaction is committed or rolled back.
// If the context is canceled, the `alphasql` package will roll back
// the transaction. [alphasql.TX.Commit] will return an error if the context provided to
// BeginTX is canceled.
//
// The provided [alphasql.TXOptions] is optional and may be nil if defaults should be used.
// If a non-default isolation level is used that the driver doesn't support,
// an error will be returned.
func (p *Pool) BeginTX(ctx context.Context, options *alphasql.TXOptions) (alphasql.TX, error) {
	c, err := p.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	t, err := c.BeginTX(ctx, options)
	if err != nil {
		p.Release(ctx, c)
		return nil, err
	}
	return p.getPoolTX(c, t), nil
}

func (p *Pool) getPoolRows(c *Connection, r alphasql.Rows) *poolRows {
	return &poolRows{c: c, p: p, rows: r}
}

func (p *Pool) getPoolErrRows(err error) *poolErrRows {
	return &poolErrRows{err: err}
}

func (p *Pool) getPoolRow(c *Connection, r alphasql.Row) *poolRow {
	return &poolRow{p: p, c: c, r: r}
}

func (p *Pool) getPoolErrRow(err error) *poolErrRow {
	return &poolErrRow{err: err}
}

func (p *Pool) getPoolErrResult(err error) *poolErrResult {
	return &poolErrResult{err: err}
}

func (p *Pool) getPoolTX(c *Connection, t alphasql.TX) *poolTX {
	return &poolTX{p: p, c: c, t: t}
}

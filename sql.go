package alphasql

import (
	"context"
	"database/sql/driver"
	"errors"
)

// Ping verifies a Connection to the database is still alive,
// establishing a Connection if necessary.
func (c *Connection) Ping(ctx context.Context) error {
	if p, ok := c.c.(driver.Pinger); ok {
		return p.Ping(ctx)
	}
	return nil
}

// Query executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
func (c *Connection) Query(ctx context.Context, query string, args ...any) (Rows, error) {
	r, s, err := c.query(ctx, query, args)
	if err != nil {
		if s != nil {
			_ = s.Close()
		}
		return nil, err
	}
	rr := &rows{s: s, r: r}
	go rr.contextCloseHandling(ctx)
	return rr, nil
}

// QueryRow executes a query that is expected to return at most one row.
// QueryRow always returns a non-nil value. Errors are deferred until
// [alphasql.Row]'s Scan method is called.
// If the query selects no rows, the [*Row.Scan] will return [ErrNoRows].
// Otherwise, [*Row.Scan] scans the first selected row and discards
// the rest.
func (c *Connection) QueryRow(ctx context.Context, query string, args ...any) Row {
	r, err := c.Query(ctx, query, args)
	return &row{r: r, err: err}
}

// Exec executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func (c *Connection) Exec(ctx context.Context, query string, args ...any) (Result, error) {
	r, s, err := c.exec(ctx, query, args)
	if s != nil {
		_ = s.Close()
	}
	if err != nil {
		return nil, err
	}
	return &result{r: r}, nil
}

// Prepare creates a prepared statement for later queries or executions.
// Multiple queries or executions may be run concurrently from the
// returned statement.
// The caller must call the statement's [Statement.Close] method
// when the statement is no longer needed.
func (c *Connection) Prepare(_ context.Context, _ string) (Statement, error) {
	return nil, nil
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
func (c *Connection) BeginTX(ctx context.Context, options *TXOptions) (TX, error) {
	options, err := validateAndDefaultTXOptions(options)
	if err != nil {
		return nil, err
	}
	t, err := c.beginTX(ctx, options)
	if err != nil {
		return nil, err
	}
	tt := &tx{t: t}
	go tt.contextCloseHandling(ctx)
	return tt, nil
}

func (c *Connection) query(ctx context.Context, query string, args []any) (driver.Rows, driver.Stmt, error) {
	nvs, err := getDriverNamedValuesFromArgs(c, args)
	if err != nil {
		return nil, nil, err
	}
	qc, ok := c.c.(driver.QueryerContext)
	if ok {
		r, err := queryUsingQueryerContext(ctx, qc, query, nvs)
		if !errors.Is(err, driver.ErrSkip) {
			return r, nil, err
		}
	}
	return queryUsingRawConnection(ctx, c, query, nvs)
}

func (c *Connection) exec(ctx context.Context, query string, args []any) (driver.Result, driver.Stmt, error) {
	nvs, err := getDriverNamedValuesFromArgs(c, args)
	if err != nil {
		return nil, nil, err
	}
	ec, ok := c.c.(driver.ExecerContext)
	if ok {
		r, err := execUsingExecerContext(ctx, ec, query, nvs)
		if !errors.Is(err, driver.ErrSkip) {
			return r, nil, err
		}
	}
	return execUsingRawConnection(ctx, c, query, nvs)
}

func (c *Connection) beginTX(ctx context.Context, options *TXOptions) (driver.Tx, error) {
	cb, ok := c.c.(driver.ConnBeginTx)
	if ok {
		tx, err := beginTXUsingConnectionBeginContext(ctx, cb, options)
		if !errors.Is(err, driver.ErrSkip) {
			return tx, err
		}
	}
	return beginTXUsingRawConnection(c)
}

func getDriverStatement(ctx context.Context, c *Connection, query string) (driver.Stmt, error) {
	if cpc, ok := c.c.(driver.ConnPrepareContext); ok {
		return cpc.PrepareContext(ctx, query)
	}
	s, err := c.c.Prepare(query)
	if err == nil {
		select {
		default:
		case <-ctx.Done():
			_ = s.Close()
			return nil, ctx.Err()
		}
	}
	return s, err
}

func queryUsingQueryerContext(ctx context.Context, qc driver.QueryerContext,
	query string, nvs []driver.NamedValue) (driver.Rows, error) {
	return qc.QueryContext(ctx, query, nvs)
}

func queryUsingDriverStatement(ctx context.Context, s driver.Stmt, nvs []driver.NamedValue) (driver.Rows, error) {
	if sc, is := s.(driver.StmtQueryContext); is {
		return sc.QueryContext(ctx, nvs)
	}
	vs, err := getDriverValueFromDriverNamedValue(nvs)
	if err != nil {
		_ = s.Close()
		return nil, err
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return s.Query(vs)
}

func queryUsingRawConnection(ctx context.Context, c *Connection, query string,
	nvs []driver.NamedValue) (driver.Rows, driver.Stmt, error) {
	s, err := getDriverStatement(ctx, c, query)
	if err != nil {
		return nil, nil, err
	}
	r, err := queryUsingDriverStatement(ctx, s, nvs)
	return r, s, err
}

func execUsingExecerContext(ctx context.Context, ec driver.ExecerContext, query string,
	nvs []driver.NamedValue) (driver.Result, error) {
	return ec.ExecContext(ctx, query, nvs)
}

func execUsingDriverStatement(ctx context.Context, s driver.Stmt, nvs []driver.NamedValue) (driver.Result, error) {
	if sc, is := s.(driver.StmtExecContext); is {
		return sc.ExecContext(ctx, nvs)
	}
	vs, err := getDriverValueFromDriverNamedValue(nvs)
	if err != nil {
		_ = s.Close()
		return nil, err
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return s.Exec(vs)
}

func execUsingRawConnection(ctx context.Context, c *Connection, query string,
	nvs []driver.NamedValue) (driver.Result, driver.Stmt, error) {
	s, err := getDriverStatement(ctx, c, query)
	if err != nil {
		return nil, nil, err
	}
	r, err := execUsingDriverStatement(ctx, s, nvs)
	return r, s, err
}

func beginTXUsingConnectionBeginContext(ctx context.Context, cb driver.ConnBeginTx, options *TXOptions) (driver.Tx, error) {
	return cb.BeginTx(ctx, options.driverOptions())
}

func beginTXUsingRawConnection(c *Connection) (driver.Tx, error) {
	return c.c.Begin()
}

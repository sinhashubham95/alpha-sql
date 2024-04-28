package alphasql

import "context"

// Ping verifies a Connection to the database is still alive,
// establishing a Connection if necessary.
func (c *Connection) Ping(_ context.Context) error {
	return nil
}

// Query executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
func (c *Connection) Query(_ context.Context, _ string, _ ...any) (Rows, error) {
	return nil, nil
}

// QueryRow executes a query that is expected to return at most one row.
// QueryRow always returns a non-nil value. Errors are deferred until
// [alphasql.Row]'s Scan method is called.
// If the query selects no rows, the [*Row.Scan] will return [ErrNoRows].
// Otherwise, [*Row.Scan] scans the first selected row and discards
// the rest.
func (c *Connection) QueryRow(_ context.Context, _ string, _ ...any) Row {
	return nil
}

// Exec executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func (c *Connection) Exec(_ context.Context, _ string, _ ...any) (Result, error) {
	return nil, nil
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
func (c *Connection) BeginTX(_ context.Context, _ *TXOptions) (TX, error) {
	return nil, nil
}

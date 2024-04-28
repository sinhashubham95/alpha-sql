package alphasql

import "context"

// Statement is a prepared statement.
// A Statement is safe for concurrent use by multiple goroutines.
//
// If a Statement is prepared on a [TX] or [Connection], it will be bound to a single
// underlying connection forever. If the [TX] or [Connection] closes, the Statement will
// become unusable and all operations will return an error.
type Statement interface {
	Close(ctx context.Context) error
	NumberOfInputs() int
	Exec(ctx context.Context, args ...any) (Result, error)
	Query(ctx context.Context, args ...any) (Rows, error)
	QueryRow(ctx context.Context, args ...any) (Row, error)
}

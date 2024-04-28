package alphasql

import "context"

// Rows is the result of a query. Its cursor starts before the first row
// of the result set. Use [Rows.Next] to advance from row to row.
type Rows interface {
	Next() bool
	NextResultSet() bool
	Error() error
	Close(ctx context.Context) error
	Scan(values ...any) error
	Values() ([]any, error)
	RawValues() [][]byte
}

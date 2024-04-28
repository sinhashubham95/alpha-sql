package alphasql

import (
	"context"
	"database/sql/driver"
)

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

type rows struct {
	c *Connection
	r driver.Rows
}

func (r *rows) Next() bool {
	return false
}

func (r *rows) NextResultSet() bool {
	return false
}

func (r *rows) Error() error {
	return nil
}

func (r *rows) Close(_ context.Context) error {
	return nil
}

func (r *rows) Scan(_ ...any) error {
	return nil
}

func (r *rows) Values() ([]any, error) {
	return nil, nil
}

func (r *rows) RawValues() [][]byte {
	return nil
}

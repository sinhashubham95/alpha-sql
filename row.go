package alphasql

import (
	"context"
	"database/sql/driver"
)

type Row interface {
	// Scan copies the columns from the matched row into the values
	// pointed at by dest. See the documentation on [Rows.Scan] for details.
	// If more than one row matches the query,
	// Scan uses the first row and discards the rest. If no row matches
	// the query, Scan returns [ErrNoRows].
	Scan(ctx context.Context, values ...any) error

	Columns() []string

	// Error provides a way for wrapping packages to check for
	// query errors without calling [Row.Scan].
	// Error returns the error, if any, that was encountered while running the query.
	// If this error is not nil, this error will also be returned from [Row.Scan].
	Error() error
}

type row struct {
	s       driver.Stmt
	r       Rows
	columns []string
	err     error
}

func (r *row) Scan(ctx context.Context, values ...any) error {
	if r.err != nil {
		return r.err
	}
	if !r.r.Next(ctx) {
		_ = r.close(ctx)
		if err := r.r.Error(); err != nil {
			return err
		}
		return ErrNoRows
	}
	err := r.r.Scan(values...)
	if err != nil {
		return err
	}
	return r.close(ctx)
}

func (r *row) Columns() []string {
	return r.r.Columns()
}

func (r *row) Error() error {
	return r.err
}

func (r *row) close(ctx context.Context) error {
	err := r.r.Close(ctx)
	if err != nil {
		_ = r.s.Close()
		return err
	}
	return r.s.Close()
}

package alphasql

import (
	"context"
	"database/sql/driver"
)

type Row interface {
	Scan(ctx context.Context, values ...any) error
	Error() error
}

type row struct {
	c   *Connection
	s   driver.Stmt
	r   Rows
	err error
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

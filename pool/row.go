package pool

import (
	"context"
	alphasql "github.com/sinhashubham95/alpha-sql"
)

type poolRow struct {
	c *Connection
	p *Pool
	r alphasql.Row
}

type poolErrRow struct {
	err error
}

func (p *poolRow) Scan(ctx context.Context, values ...any) error {
	var err error
	panicked := true
	defer func() {
		if panicked && p.c != nil {
			p.p.closeOrRelease(ctx, p.c, err)
		}
	}()
	err = p.r.Scan(ctx, values...)
	panicked = false
	if p.c != nil {
		p.p.closeOrRelease(ctx, p.c, err)
	}
	return err
}

func (p *poolRow) Error() error {
	return p.r.Error()
}

func (p *poolRow) Columns() []alphasql.Column {
	return p.r.Columns()
}

func (p *poolErrRow) Scan(_ context.Context, _ ...any) error {
	return nil
}

func (p *poolErrRow) Error() error {
	return p.err
}

func (p *poolErrRow) Columns() []alphasql.Column {
	return nil
}

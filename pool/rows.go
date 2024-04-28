package pool

import (
	"context"
	alphasql "github.com/sinhashubham95/alpha-sql"
)

type poolRows struct {
	c    *Connection
	p    *Pool
	rows alphasql.Rows
}

type poolErrRows struct {
	err error
}

func (p *poolRows) Next(ctx context.Context) bool {
	return p.rows.Next(ctx)
}

func (p *poolRows) NextResultSet(ctx context.Context) bool {
	return p.rows.Next(ctx)
}

func (p *poolRows) Error() error {
	return p.rows.Error()
}

func (p *poolRows) Close(ctx context.Context) error {
	err := p.rows.Close(ctx)
	if p.c != nil {
		p.p.Release(ctx, p.c)
		p.c = nil
	}
	return err
}

func (p *poolRows) Scan(values ...any) error {
	return p.rows.Scan(values...)
}

func (p *poolErrRows) Next(_ context.Context) bool {
	return false
}

func (p *poolErrRows) NextResultSet(_ context.Context) bool {
	return false
}

func (p *poolErrRows) Error() error {
	return p.err
}

func (p *poolErrRows) Close(_ context.Context) error {
	return nil
}

func (p *poolErrRows) Scan(_ ...any) error {
	return nil
}

func (p *poolErrRows) Values() ([]any, error) {
	return nil, nil
}

func (p *poolErrRows) RawValues() [][]byte {
	return nil
}

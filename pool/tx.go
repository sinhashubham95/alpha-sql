package pool

import (
	"context"
	alphasql "github.com/sinhashubham95/alpha-sql"
)

type poolTX struct {
	p *Pool
	c *Connection
	t alphasql.TX
}

func (p *poolTX) Commit(ctx context.Context) error {
	err := p.t.Commit(ctx)
	if p.c != nil {
		p.p.closeOrRelease(ctx, p.c, err)
		p.c = nil
	}
	return err
}

func (p *poolTX) Rollback(ctx context.Context) error {
	err := p.t.Rollback(ctx)
	if p.c != nil {
		if p.KeepConnectionOnRollback() {
			p.p.closeOrRelease(ctx, p.c, err)
		} else {
			go p.p.p.destroyAcquiredConnection(ctx, p.c)
		}
		p.c = nil
	}
	return err
}

func (p *poolTX) Query(ctx context.Context, query string, args ...any) (alphasql.Rows, error) {
	return p.t.Query(ctx, query, args...)
}

func (p *poolTX) QueryRow(ctx context.Context, query string, args ...any) alphasql.Row {
	return p.t.QueryRow(ctx, query, args...)
}

func (p *poolTX) Exec(ctx context.Context, query string, args ...any) (alphasql.Result, error) {
	return p.t.Exec(ctx, query, args...)
}

func (p *poolTX) Prepare(ctx context.Context, query string) (alphasql.Statement, error) {
	return p.t.Prepare(ctx, query)
}

func (p *poolTX) Statement(ctx context.Context, s alphasql.Statement) (alphasql.Statement, error) {
	return p.t.Statement(ctx, s)
}

func (p *poolTX) KeepConnectionOnRollback() bool {
	return p.t.KeepConnectionOnRollback()
}

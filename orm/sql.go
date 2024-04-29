package orm

import (
	"context"
	alphasql "github.com/sinhashubham95/alpha-sql"
	"github.com/sinhashubham95/alpha-sql/orm/entity"
)

func (o *orm) GetByID(ctx context.Context, e entity.Entity) error {
	r := o.p.QueryRow(ctx, e.GetIDQuery(), e.GetIDArgs()...)
	if r.Error() != nil {
		return r.Error()
	}
	return e.BindRow(&scannerRow{r: r, isScanToStructureEnabled: o.isScanToStructureEnabled})
}

func (o *orm) GetAll(ctx context.Context, e entity.Entity) ([]entity.Entity, error) {
	r, err := o.p.Query(ctx, e.GetAllQuery(), e.GetAllQueryArgs()...)
	if err != nil {
		return nil, err
	}
	defer closeRows(ctx, r)
	result := make([]entity.Entity, 0)
	for r.Next(ctx) {
		err = e.BindRow(&scannerRows{r: r, isScanToStructureEnabled: o.isScanToStructureEnabled})
		if err != nil {
			return nil, err
		}
		result = append(result, e)
		e = e.GetNext()
	}
	if len(result) == 0 {
		return nil, alphasql.ErrNoRows
	}
	return result, nil
}

func (o *orm) FreshSave(ctx context.Context, es ...entity.Entity) error {
	tx, err := o.p.BeginTX(ctx, nil)
	if err != nil {
		return err
	}
	defer rollbackTX(ctx, tx)
	for _, e := range es {
		r, err := tx.Exec(ctx, e.GetFreshSaveQuery(), e.GetFreshSaveArgs()...)
		if err != nil {
			return err
		}
		if o.failOnNoRowsAffected {
			rows, err := r.RowsAffected()
			if err != nil {
				return err
			}
			if rows == 0 {
				return alphasql.ErrNoRowsAffected
			}
		}
	}
	return tx.Commit(ctx)
}

func (o *orm) Save(ctx context.Context, es ...entity.Entity) error {
	tx, err := o.p.BeginTX(ctx, nil)
	if err != nil {
		return err
	}
	defer rollbackTX(ctx, tx)
	for _, e := range es {
		r, err := tx.Exec(ctx, e.GetSaveQuery(), e.GetSaveArgs()...)
		if err != nil {
			return err
		}
		if o.failOnNoRowsAffected {
			rows, err := r.RowsAffected()
			if err != nil {
				return err
			}
			if rows == 0 {
				return alphasql.ErrNoRowsAffected
			}
		}
	}
	return tx.Commit(ctx)
}

func (o *orm) Delete(ctx context.Context, es ...entity.Entity) error {
	tx, err := o.p.BeginTX(ctx, nil)
	if err != nil {
		return err
	}
	defer rollbackTX(ctx, tx)
	for _, e := range es {
		r, err := tx.Exec(ctx, e.GetDeleteQuery(), e.GetDeleteArgs()...)
		if err != nil {
			return err
		}
		if o.failOnNoRowsAffected {
			rows, err := r.RowsAffected()
			if err != nil {
				return err
			}
			if rows == 0 {
				return alphasql.ErrNoRowsAffected
			}
		}
	}
	return tx.Commit(ctx)
}

func (o *orm) QueryRow(ctx context.Context, e entity.RawEntity, code int) error {
	r := o.p.QueryRow(ctx, e.GetQueryRow(code), e.GetQueryRowArgs(code)...)
	if r.Error() != nil {
		return r.Error()
	}
	return e.BindRow(code, &scannerRow{r: r, isScanToStructureEnabled: o.isScanToStructureEnabled})
}

func (o *orm) Query(ctx context.Context, e entity.RawEntity, code int) ([]entity.RawEntity, error) {
	r, err := o.p.Query(ctx, e.GetQuery(code), e.GetQueryArgs(code)...)
	if err != nil {
		return nil, err
	}
	defer closeRows(ctx, r)
	result := make([]entity.RawEntity, 0)
	for r.Next(ctx) {
		err = e.BindRow(code, &scannerRows{r: r, isScanToStructureEnabled: o.isScanToStructureEnabled})
		if err != nil {
			return nil, err
		}
		result = append(result, e)
		e = e.GetNext()
	}
	if len(result) == 0 {
		return nil, alphasql.ErrNoRows
	}
	return result, nil
}

func (o *orm) Exec(ctx context.Context, es ...entity.RawExec) error {
	tx, err := o.p.BeginTX(ctx, nil)
	if err != nil {
		return err
	}
	defer rollbackTX(ctx, tx)
	for _, e := range es {
		r, err := tx.Exec(ctx, e.Entity.GetExec(e.Code), e.Entity.GetExecArgs(e.Code)...)
		if err != nil {
			return err
		}
		if o.failOnNoRowsAffected {
			rows, err := r.RowsAffected()
			if err != nil {
				return err
			}
			if rows == 0 {
				return alphasql.ErrNoRowsAffected
			}
		}
	}
	return tx.Commit(ctx)
}

func (t *transactionalORM) GetByID(ctx context.Context, e entity.Entity) error {
	r := t.tx.QueryRow(ctx, e.GetIDQuery(), e.GetIDArgs()...)
	if r.Error() != nil {
		return r.Error()
	}
	return e.BindRow(&scannerRow{r: r, isScanToStructureEnabled: t.o.isScanToStructureEnabled})
}

func (t *transactionalORM) GetAll(ctx context.Context, e entity.Entity) ([]entity.Entity, error) {
	r, err := t.tx.Query(ctx, e.GetAllQuery(), e.GetAllQueryArgs()...)
	if err != nil {
		return nil, err
	}
	defer closeRows(ctx, r)
	result := make([]entity.Entity, 0)
	for r.Next(ctx) {
		err = e.BindRow(&scannerRows{r: r, isScanToStructureEnabled: t.o.isScanToStructureEnabled})
		if err != nil {
			return nil, err
		}
		result = append(result, e)
		e = e.GetNext()
	}
	if len(result) == 0 {
		return nil, alphasql.ErrNoRows
	}
	return result, nil
}

func (t *transactionalORM) FreshSave(ctx context.Context, es ...entity.Entity) error {
	for _, e := range es {
		r, err := t.tx.Exec(ctx, e.GetFreshSaveQuery(), e.GetFreshSaveArgs()...)
		if err != nil {
			return err
		}
		if t.o.failOnNoRowsAffected {
			rows, err := r.RowsAffected()
			if err != nil {
				return err
			}
			if rows == 0 {
				return alphasql.ErrNoRowsAffected
			}
		}
	}
	return nil
}

func (t *transactionalORM) Save(ctx context.Context, es ...entity.Entity) error {
	for _, e := range es {
		r, err := t.tx.Exec(ctx, e.GetSaveQuery(), e.GetSaveArgs()...)
		if err != nil {
			return err
		}
		if t.o.failOnNoRowsAffected {
			rows, err := r.RowsAffected()
			if err != nil {
				return err
			}
			if rows == 0 {
				return alphasql.ErrNoRowsAffected
			}
		}
	}
	return nil
}

func (t *transactionalORM) Delete(ctx context.Context, es ...entity.Entity) error {
	for _, e := range es {
		r, err := t.tx.Exec(ctx, e.GetDeleteQuery(), e.GetDeleteArgs()...)
		if err != nil {
			return err
		}
		if t.o.failOnNoRowsAffected {
			rows, err := r.RowsAffected()
			if err != nil {
				return err
			}
			if rows == 0 {
				return alphasql.ErrNoRowsAffected
			}
		}
	}
	return nil
}

func (t *transactionalORM) QueryRow(ctx context.Context, e entity.RawEntity, code int) error {
	r := t.tx.QueryRow(ctx, e.GetQueryRow(code), e.GetQueryRowArgs(code)...)
	if r.Error() != nil {
		return r.Error()
	}
	return e.BindRow(code, &scannerRow{r: r, isScanToStructureEnabled: t.o.isScanToStructureEnabled})
}

func (t *transactionalORM) Query(ctx context.Context, e entity.RawEntity, code int) ([]entity.RawEntity, error) {
	r, err := t.tx.Query(ctx, e.GetQuery(code), e.GetQueryArgs(code)...)
	if err != nil {
		return nil, err
	}
	defer closeRows(ctx, r)
	result := make([]entity.RawEntity, 0)
	for r.Next(ctx) {
		err = e.BindRow(code, &scannerRows{r: r, isScanToStructureEnabled: t.o.isScanToStructureEnabled})
		if err != nil {
			return nil, err
		}
		result = append(result, e)
		e = e.GetNext()
	}
	if len(result) == 0 {
		return nil, alphasql.ErrNoRows
	}
	return result, nil
}

func (t *transactionalORM) Exec(ctx context.Context, es ...entity.RawExec) error {
	for _, e := range es {
		r, err := t.tx.Exec(ctx, e.Entity.GetExec(e.Code), e.Entity.GetExecArgs(e.Code)...)
		if err != nil {
			return err
		}
		if t.o.failOnNoRowsAffected {
			rows, err := r.RowsAffected()
			if err != nil {
				return err
			}
			if rows == 0 {
				return alphasql.ErrNoRowsAffected
			}
		}
	}
	return nil
}

func (t *transactionalORM) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *transactionalORM) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

func closeRows(ctx context.Context, r alphasql.Rows) {
	_ = r.Close(ctx)
}

func rollbackTX(ctx context.Context, tx alphasql.TX) {
	_ = tx.Rollback(ctx)
}

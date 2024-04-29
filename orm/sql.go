package orm

import (
	"context"
	"github.com/sinhashubham95/alpha-sql/orm/entity"
)

func (o *orm) GetByID(ctx context.Context, entity entity.Entity) error {
	//TODO implement me
	panic("implement me")
}

func (o *orm) GetAll(ctx context.Context, entity entity.Entity) ([]entity.Entity, error) {
	//TODO implement me
	panic("implement me")
}

func (o *orm) FreshSave(ctx context.Context, entities ...entity.Entity) error {
	//TODO implement me
	panic("implement me")
}

func (o *orm) Save(ctx context.Context, source string, entities ...entity.Entity) error {
	//TODO implement me
	panic("implement me")
}

func (o *orm) Delete(ctx context.Context, entities ...entity.Entity) error {
	//TODO implement me
	panic("implement me")
}

func (o *orm) QueryRow(ctx context.Context, entity entity.RawEntity, code int) error {
	//TODO implement me
	panic("implement me")
}

func (o *orm) Query(ctx context.Context, entity entity.RawEntity, code int) ([]entity.RawEntity, error) {
	//TODO implement me
	panic("implement me")
}

func (o *orm) Exec(ctx context.Context, source string, execs ...entity.RawExec) error {
	//TODO implement me
	panic("implement me")
}

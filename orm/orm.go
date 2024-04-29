package orm

import (
	"context"
	alphasql "github.com/sinhashubham95/alpha-sql"
	"github.com/sinhashubham95/alpha-sql/orm/entity"
	"github.com/sinhashubham95/alpha-sql/pool"
	"sync/atomic"
)

// ORM is used to provide the ORM functionalities.
type ORM interface {
	// GetByID is used to handle scenarios where the data of an entity has to be fetched by a primary key.
	GetByID(ctx context.Context, entity entity.Entity) error

	// GetAll is used to handle scenarios where all the data of an entity has to be fetched.
	GetAll(ctx context.Context, entity entity.Entity) ([]entity.Entity, error)

	// FreshSave is used to freshly save(insert) the provided set of entities.
	FreshSave(ctx context.Context, entities ...entity.Entity) error

	// Save is used ot save(upsert) the provided set of entities.
	Save(ctx context.Context, source string, entities ...entity.Entity) error

	// Delete is used to delete the provided set of entities.
	Delete(ctx context.Context, entities ...entity.Entity) error

	// QueryRow is used to perform the query for the code specified.
	QueryRow(ctx context.Context, entity entity.RawEntity, code int) error

	// Query is used to perform the query fetching all the rows as per the code specified.
	Query(ctx context.Context, entity entity.RawEntity, code int) ([]entity.RawEntity, error)

	// Exec is used to execute all the executions as per the entity and the code specified.
	Exec(ctx context.Context, source string, execs ...entity.RawExec) error

	// Close is used to close the ORM.
	Close(ctx context.Context) error
}

// Configuration is the set of parameters for ORM.
type Configuration struct {
	PoolConfig               *pool.Config
	IsScanToStructureEnabled bool
}

// orm is used to provide a wrapper around the orm functionalities.
type orm struct {
	p *pool.Pool

	cfg                      *Configuration
	isScanToStructureEnabled bool

	closed atomic.Bool
}

// New is used to create a new instance of the ORM.
func New(ctx context.Context, cfg *Configuration) (ORM, error) {
	err := cfg.validate()
	if err != nil {
		return nil, err
	}
	p, err := pool.New(ctx, cfg.PoolConfig)
	if err != nil {
		return nil, err
	}
	return &orm{
		p:                        p,
		cfg:                      cfg,
		isScanToStructureEnabled: cfg.IsScanToStructureEnabled,
	}, nil
}

// Close is used to close the orm.
func (o *orm) Close(ctx context.Context) error {
	if o.closed.CompareAndSwap(false, true) {
		o.p.Close(ctx)
	}
	return alphasql.ErrORMClosed
}

func (c *Configuration) validate() error {
	if c.PoolConfig == nil {
		return alphasql.ErrMissingPoolConfig
	}
	return nil
}

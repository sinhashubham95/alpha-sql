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
	GetByID(ctx context.Context, e entity.Entity) error

	// GetAll is used to handle scenarios where all the data of an entity has to be fetched.
	GetAll(ctx context.Context, e entity.Entity) ([]entity.Entity, error)

	// FreshSave is used to freshly save(insert) the provided set of entities.
	FreshSave(ctx context.Context, es ...entity.Entity) error

	// Save is used ot save(upsert) the provided set of entities.
	Save(ctx context.Context, es ...entity.Entity) error

	// Delete is used to delete the provided set of entities.
	Delete(ctx context.Context, es ...entity.Entity) error

	// QueryRow is used to perform the query for the code specified.
	QueryRow(ctx context.Context, e entity.RawEntity, code int) error

	// Query is used to perform the query fetching all the rows as per the code specified.
	Query(ctx context.Context, e entity.RawEntity, code int) ([]entity.RawEntity, error)

	// Exec is used to execute all the executions as per the entity and the code specified.
	Exec(ctx context.Context, es ...entity.RawExec) error

	// BeginTX is used to start a new transaction on ORM.
	BeginTX(ctx context.Context, options *alphasql.TXOptions) (TransactionalORM, error)

	// Close is used to close the ORM.
	Close(ctx context.Context) error
}

// TransactionalORM is used to provide the ORM functionalities around an alphasql.TX.
type TransactionalORM interface {
	// GetByID is used to handle scenarios where the data of an entity has to be fetched by a primary key.
	GetByID(ctx context.Context, e entity.Entity) error

	// GetAll is used to handle scenarios where all the data of an entity has to be fetched.
	GetAll(ctx context.Context, e entity.Entity) ([]entity.Entity, error)

	// FreshSave is used to freshly save(insert) the provided set of entities.
	FreshSave(ctx context.Context, es ...entity.Entity) error

	// Save is used ot save(upsert) the provided set of entities.
	Save(ctx context.Context, es ...entity.Entity) error

	// Delete is used to delete the provided set of entities.
	Delete(ctx context.Context, es ...entity.Entity) error

	// QueryRow is used to perform the query for the code specified.
	QueryRow(ctx context.Context, e entity.RawEntity, code int) error

	// Query is used to perform the query fetching all the rows as per the code specified.
	Query(ctx context.Context, e entity.RawEntity, code int) ([]entity.RawEntity, error)

	// Exec is used to execute all the executions as per the entity and the code specified.
	Exec(ctx context.Context, es ...entity.RawExec) error

	// Commit is used to commit the transaction.
	Commit(ctx context.Context) error

	// Rollback is used to roll back the transaction.
	Rollback(ctx context.Context) error
}

// Configuration is the set of parameters for ORM.
type Configuration struct {
	PoolConfig               *pool.Config
	IsScanToStructureEnabled bool
	FailOnNoRowsAffected     bool
}

// orm is used to provide a wrapper around the orm functionalities.
type orm struct {
	p *pool.Pool

	cfg                      *Configuration
	isScanToStructureEnabled bool
	failOnNoRowsAffected     bool

	closed atomic.Bool
}

// transactionalORM is used to provide a wrapper around using the transactional orm functionalities.
type transactionalORM struct {
	o  *orm
	tx alphasql.TX
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
		failOnNoRowsAffected:     cfg.FailOnNoRowsAffected,
	}, nil
}

func (o *orm) BeginTX(ctx context.Context, options *alphasql.TXOptions) (TransactionalORM, error) {
	tx, err := o.p.BeginTX(ctx, options)
	if err != nil {
		return nil, err
	}
	return &transactionalORM{
		o:  o,
		tx: tx,
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

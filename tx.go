package alphasql

import (
	"context"
	"database/sql/driver"
)

// TXIsolationLevel is the transaction isolation level used in [TXOptions].
type TXIsolationLevel int

// Various isolation levels that drivers may support in [Connection.BeginTX].
// If a driver does not support a given isolation level an error may be returned.
//
// See https://en.wikipedia.org/wiki/Isolation_(database_systems)#Isolation_levels.
const (
	TXIsolationLevelDefault TXIsolationLevel = iota
	TXIsolationLevelReadUncommitted
	TXIsolationLevelReadCommitted
	TXIsolationLevelWriteCommitted
	TXIsolationLevelRepeatableRead
	TXIsolationLevelSnapshot
	TXIsolationLevelSerializable
	TXIsolationLevelLinearizable
)

// TXAccessMode is the transaction access mode (read write or read only)
type TXAccessMode string

// Transaction access modes
const (
	TXAccessModeReadWrite TXAccessMode = "read write"
	TXAccessModeReadOnly  TXAccessMode = "read only"
)

// TXOptions holds the transaction options to be used in [Connection.BeginTX].
type TXOptions struct {
	IsolationLevel TXIsolationLevel
	AccessMode     TXAccessMode
}

// TX is an in-progress database transaction.
//
// A transaction must end with a call to [TX.Commit] or [TX.Rollback].
//
// After a call to [TX.Commit] or [TX.Rollback], all operations on the
// transaction fail with [ErrTXClosed].
//
// The statements prepared for a transaction by calling
// the transaction's [TX.Prepare] are closed
// by the call to [TX.Commit] or [TX.Rollback].
type TX interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	Query(ctx context.Context, query string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) Row
	Exec(ctx context.Context, query string, args ...any) (Result, error)
	Prepare(ctx context.Context, query string) (Statement, error)
	Statement(ctx context.Context, s Statement) (Statement, error)
	KeepConnectionOnRollback() bool
}

type tx struct {
	c                        *Connection
	t                        driver.Tx
	closed                   bool
	keepConnectionOnRollback bool
}

func (t *tx) Commit(_ context.Context) error {
	if t.closed {
		return ErrTXClosed
	}
	t.closed = true
	return t.t.Commit()
}

func (t *tx) Rollback(_ context.Context) error {
	if t.closed {
		return ErrTXClosed
	}
	t.closed = true
	return t.t.Rollback()
}

func (t *tx) Query(ctx context.Context, query string, args ...any) (Rows, error) {
	return t.c.Query(ctx, query, args...)
}

func (t *tx) QueryRow(ctx context.Context, query string, args ...any) Row {
	return t.c.QueryRow(ctx, query, args...)
}

func (t *tx) Exec(ctx context.Context, query string, args ...any) (Result, error) {
	return t.c.Exec(ctx, query, args...)
}

func (t *tx) Prepare(_ context.Context, _ string) (Statement, error) {
	return nil, nil
}

func (t *tx) Statement(_ context.Context, _ Statement) (Statement, error) {
	return nil, nil
}

func (t *tx) KeepConnectionOnRollback() bool {
	return t.keepConnectionOnRollback
}

func validateAndDefaultTXOptions(options *TXOptions) (*TXOptions, error) {
	if options == nil {
		options = &TXOptions{
			IsolationLevel: TXIsolationLevelDefault,
			AccessMode:     TXAccessModeReadWrite,
		}
	}
	if options.IsolationLevel < TXIsolationLevelDefault || options.IsolationLevel > TXIsolationLevelLinearizable {
		return nil, ErrTXOptionsInvalidIsolationLevel
	}
	if options.AccessMode != TXAccessModeReadWrite && options.AccessMode != TXAccessModeReadOnly {
		return nil, ErrTXOptionsInvalidAccessMode
	}
	return options, nil
}

func (o *TXOptions) driverOptions() driver.TxOptions {
	return driver.TxOptions{
		Isolation: driver.IsolationLevel(o.IsolationLevel),
		ReadOnly:  o.AccessMode == TXAccessModeReadOnly,
	}
}

package alphasql

import "context"

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

// TXDeferrableMode is the transaction deferrable mode (deferrable or not deferrable)
type TXDeferrableMode string

// Transaction deferrable modes
const (
	TXDeferrableModeDeferrable    TXDeferrableMode = "deferrable"
	TXDeferrableModeNotDeferrable TXDeferrableMode = "not deferrable"
)

// TXOptions holds the transaction options to be used in [Connection.BeginTX].
type TXOptions struct {
	IsolationLevel TXIsolationLevel
	AccessMode     TXAccessMode
	DeferrableMode TXDeferrableMode

	// BeginQuery is the SQL query that will be executed to begin the transaction. This allows using non-standard syntax
	// such as BEGIN PRIORITY HIGH with CockroachDB. If set this will override the other settings.
	BeginQuery string
}

// TX is an in-progress database transaction.
//
// A transaction must end with a call to [TX.Commit] or [TX.Rollback].
//
// After a call to [TX.Commit] or [TX.Rollback], all operations on the
// transaction fail with [ErrTXDone].
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
}

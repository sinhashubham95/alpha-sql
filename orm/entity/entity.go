package entity

import "context"

// Scanner is used to scan the data to the respective types.
type Scanner interface {
	Scan(values ...interface{}) error
	ScanStructure(value interface{}) error
}

// Entity is used to provide the set of functionalities common around database operations on a table.
type Entity interface {
	GetIDQuery() string
	GetIDValues() []interface{}
	GetAllQuery(ctx context.Context) string
	GetNext() Entity
	BindRow(row Scanner) error
	GetFreshSaveQuery() string
	GetFreshFieldValues(source string) []interface{}
	GetSaveQuery() string
	GetFieldValues(source string) []interface{}
	GetDeleteQuery() string
	GetDeleteValues() []interface{}
}

// RawEntity is used to provide the set of raw functionalities around the database operations on a table.
type RawEntity interface {
	GetQuery(ctx context.Context, code int) string
	GetQueryValues(code int) []interface{}
	GetMultiQuery(ctx context.Context, code int) string
	GetMultiQueryValues(code int) []interface{}
	GetNextRaw() RawEntity
	BindRawRow(code int, row Scanner) error
	GetExec(code int) string
	GetExecValues(code int, source string) []interface{}
}

// RawExec is the structure for the entity and the code
type RawExec struct {
	Entity RawEntity
	Code   int
}

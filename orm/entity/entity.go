package entity

import "context"

// Scanner is used to scan the data to the respective types.
type Scanner interface {
	Scan(ctx context.Context, values ...interface{}) error
	ScanStructure(ctx context.Context, value interface{}) error
}

// Entity is used to provide the set of functionalities common around database operations on a table.
type Entity interface {
	GetIDQuery() string
	GetIDArgs() []interface{}
	GetAllQuery() string
	GetAllQueryArgs() []interface{}
	GetNext() Entity
	BindRow(row Scanner) error
	GetFreshSaveQuery() string
	GetFreshSaveArgs() []interface{}
	GetSaveQuery() string
	GetSaveArgs() []interface{}
	GetDeleteQuery() string
	GetDeleteArgs() []interface{}
}

// RawEntity is used to provide the set of raw functionalities around the database operations on a table.
type RawEntity interface {
	GetQueryRow(code int) string
	GetQueryRowArgs(code int) []interface{}
	GetQuery(code int) string
	GetQueryArgs(code int) []interface{}
	GetNext() RawEntity
	BindRow(code int, row Scanner) error
	GetExec(code int) string
	GetExecArgs(code int) []interface{}
}

// RawExec is the structure for the entity and the code
type RawExec struct {
	Entity RawEntity
	Code   int
}

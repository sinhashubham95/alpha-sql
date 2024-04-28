package alphasql

import "errors"

// errors
var (
	ErrMissingConnectionConfig = errors.New("no connection config provided")
	ErrMissingDriverName       = errors.New("driver name is a mandatory config")
	ErrMissingURL              = errors.New("url is a mandatory config")
	ErrPoolSpaceNotAvailable   = errors.New("no space available to create new connections")
	ErrPoolClosed              = errors.New("closed pool")
	ErrNoRows                  = errors.New("no rows in result set")
	ErrTXDone                  = errors.New("transaction has already been committed or rolled back")
)

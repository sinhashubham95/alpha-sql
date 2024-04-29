package alphasql

import "errors"

// errors
var (
	ErrMissingConnectionConfig        = errors.New("no connection config provided")
	ErrMissingDriverName              = errors.New("driver name is a mandatory config")
	ErrMissingURL                     = errors.New("url is a mandatory config")
	ErrPoolSpaceNotAvailable          = errors.New("no space available to create new connections")
	ErrPoolClosed                     = errors.New("closed pool")
	ErrRowsClosed                     = errors.New("rows are closed")
	ErrNoRows                         = errors.New("no rows in result set")
	ErrRowsScanWithoutNext            = errors.New("scan called without calling next")
	ErrRowsUnexpectedScanValues       = errors.New("unexpected scan values")
	ErrRowsUnexpectedScan             = errors.New("unexpected scan")
	ErrRowsUnsupportedScan            = errors.New("unsupported scan")
	ErrMissingParentRows              = errors.New("invalid context to convert cursor rows, missing parent rows")
	ErrTXClosed                       = errors.New("transaction has already been committed or rolled back")
	ErrTXOptionsInvalidIsolationLevel = errors.New("invalid transaction isolation level")
	ErrTXOptionsInvalidAccessMode     = errors.New("invalid transaction access mode")
	ErrNamedArgNoLetterBegin          = errors.New("name does not begin with a letter")
	ErrConvertingArgumentToNamedArg   = errors.New("unable to convert argument to named arg")
	ErrNamedArgNotSupported           = errors.New("driver does not support the use of Named Parameters")
	ErrNilPointer                     = errors.New("destination pointer is nil")
	ErrNotAPointer                    = errors.New("destination is not a pointer")
)

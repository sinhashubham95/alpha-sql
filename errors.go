package alphasql

import "errors"

// errors
var (
	ErrDBClosed                       = errors.New("db is closed")
	ErrMissingPoolConfig              = errors.New("no pool config provided")
	ErrMissingConnectionConfig        = errors.New("no connection config provided")
	ErrMissingDriverName              = errors.New("driver name is a mandatory config")
	ErrMissingURL                     = errors.New("url is a mandatory config")
	ErrPoolSpaceNotAvailable          = errors.New("no space available to create new connections")
	ErrORMClosed                      = errors.New("orm is closed")
	ErrPoolClosed                     = errors.New("closed pool")
	ErrRowsClosed                     = errors.New("rows are closed")
	ErrNoRows                         = errors.New("no rows in result set")
	ErrNoRowsAffected                 = errors.New("no rows affected")
	ErrRowsScanWithoutNext            = errors.New("scan called without calling next")
	ErrRowsUnexpectedScanValues       = errors.New("unexpected scan values")
	ErrRowsUnexpectedScan             = errors.New("unexpected scan")
	ErrRowsUnsupportedScan            = errors.New("unsupported scan")
	ErrTXClosed                       = errors.New("transaction has already been committed or rolled back")
	ErrTXOptionsInvalidIsolationLevel = errors.New("invalid transaction isolation level")
	ErrTXOptionsInvalidAccessMode     = errors.New("invalid transaction access mode")
	ErrNamedArgNoLetterBegin          = errors.New("name does not begin with a letter")
	ErrConvertingArgumentToNamedArg   = errors.New("unable to convert argument to named arg")
	ErrNilPointer                     = errors.New("destination pointer is nil")
	ErrNotAPointer                    = errors.New("destination is not a pointer")
	ErrBadConnection                  = errors.New("bad connection")
	ErrScanToStructureNotEnabled      = errors.New("scanning to a structure not enabled")
	ErrBatchProcessing                = errors.New("batch is processing")
	ErrBatchClosed                    = errors.New("batch is closed")
)

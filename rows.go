package alphasql

import (
	"context"
	"database/sql/driver"
	"io"
)

// Rows is the result of a query. Its cursor starts before the first row
// of the result set. Use [Rows.Next] to advance from row to row.
type Rows interface {
	// Next prepares the next result row for reading with the [Rows.Scan] method. It
	// returns true on success, or false if there is no next result row or an error
	// happened while preparing it. [Rows.Err] should be consulted to distinguish between
	// the two cases.
	//
	// Every call to [Rows.Scan], even the first one, must be preceded by a call to [Rows.Next].
	Next(ctx context.Context) bool

	// NextResultSet prepares the next result set for reading. It reports whether
	// there is further result sets, or false if there is no further result set
	// or if there is an error advancing to it. The [Rows.Err] method should be consulted
	// to distinguish between the two cases.
	//
	// After calling NextResultSet, the [Rows.Next] method should always be called before
	// scanning. If there are further result sets they may not have rows in the result
	// set.
	NextResultSet(ctx context.Context) bool

	// Error returns the error, if any, that was encountered during iteration.
	// Error may be called after an explicit or implicit [Rows.Close].
	Error() error

	// Close closes the [Rows], preventing further enumeration. If [Rows.Next] is called
	// and returns false and there are no further result sets,
	// the [Rows] are closed automatically, and it will suffice to check the
	// result of [Rows.Err]. Close is idempotent and does not affect the result of [Rows.Error].
	Close(ctx context.Context) error

	// Scan copies the columns in the current row into the values pointed
	// at by dest. The number of values in dest must be the same as the
	// number of columns in [Rows].
	//
	// Scan converts columns read from the database into the following
	// common Go types and special types provided by the sql package:
	//
	//	*string
	//	*[]byte
	//	*int, *int8, *int16, *int32, *int64
	//	*uint, *uint8, *uint16, *uint32, *uint64
	//	*bool
	//	*float32, *float64
	//	*interface{}
	//	*RawBytes
	//	*Rows (cursor value)
	//	any type implementing Scanner (see Scanner docs)
	//
	// In the most simple case, if the type of the value from the source
	// column is an integer, bool or string type T and dest is of type *T,
	// Scan simply assigns the value through the pointer.
	//
	// Scan also converts between string and numeric types, as long as no
	// information would be lost. While Scan stringifies all numbers
	// scanned from numeric database columns into *string, scans into
	// numeric types are checked for overflow. For example, a float64 with
	// value 300 or a string with value "300" can scan into an uint16, but
	// not into an uint8, though float64(255) or "255" can scan into an
	// uint8. One exception is that scans of some float64 numbers to
	// strings may lose information when stringing. In general, scan
	// floating point columns into *float64.
	//
	// If a dest argument has type *[]byte, Scan saves in that argument a
	// copy of the corresponding data. The copy is owned by the caller and
	// can be modified and held indefinitely. The copy can be avoided by
	// using an argument of type [*RawBytes] instead; see the documentation
	// for [RawBytes] for restrictions on its use.
	//
	// If an argument has type *interface{}, Scan copies the value
	// provided by the underlying driver without conversion. When scanning
	// from a source value of type []byte to *interface{}, a copy of the
	// slice is made and the caller owns the result.
	//
	// Source values of type [time.Time] may be scanned into values of type
	// *time.Time, *interface{}, *string, or *[]byte. When converting to
	// the latter two, [time.RFC3339Nano] is used.
	//
	// Source values of type bool may be scanned into types *bool,
	// *interface{}, *string, *[]byte, or [*RawBytes].
	//
	// For scanning into *bool, the source may be true, false, 1, 0, or
	// string inputs parseable by [strconv.ParseBool].
	//
	// Scan can also convert a cursor returned from a query, such as
	// "select cursor(select * from my_table) from dual", into a
	// [Rows] value that can itself be scanned from. The parent
	// select query will close any cursor [Rows] if the parent [Rows] is closed.
	//
	// If any of the first arguments implementing [driver.Scanner] returns an error,
	// that error will be wrapped in the returned error.
	Scan(values ...any) error

	// Columns are used to provide the current set of columns in the result set.
	// Similar to how until [Rows.Next] is not called, [Rows.Scan] won't work, [Rows.Columns]
	// will also return stale or nil data until [Rows.Next] is called.
	// Even though the result set has changed, until [Rows.Next] is called, the column list won't be updated.
	Columns() []string
}

type rows struct {
	s      driver.Stmt
	r      driver.Rows
	end    bool
	err    error
	closed bool

	current []driver.Value
	columns []string
}

func (r *rows) Next(ctx context.Context) bool {
	if r.closed {
		return false
	}
	doClose, ok := r.next()
	if doClose {
		_ = r.Close(ctx)
	}
	if doClose && !ok {
		r.end = true
	}
	return ok
}

func (r *rows) NextResultSet(ctx context.Context) bool {
	if r.closed {
		return false
	}
	r.current = nil
	nextResultSet, ok := r.r.(driver.RowsNextResultSet)
	if !ok {
		_ = r.Close(ctx)
		return false
	}
	r.err = nextResultSet.NextResultSet()
	if r.err != nil {
		_ = r.Close(ctx)
		return false
	}
	return true
}

func (r *rows) Error() error {
	if r.err != nil && r.err != io.EOF {
		return r.err
	}
	return nil
}

func (r *rows) Close(_ context.Context) error {
	return r.close(nil)
}

func (r *rows) Scan(vs ...any) error {
	if r.err != nil && r.err != io.EOF {
		return r.err
	}
	if r.closed {
		return ErrRowsClosed
	}
	if r.current == nil {
		return ErrRowsScanWithoutNext
	}
	if len(vs) != len(r.current) {
		return ErrRowsUnexpectedScanValues
	}
	for i, v := range r.current {
		err := convertAssignRows(v, vs[i])
		if err != nil {
			return ErrRowsUnexpectedScan
		}
	}
	return nil
}

func (r *rows) Columns() []string {
	return r.columns
}

func (r *rows) close(err error) error {
	if r.closed {
		return nil
	}
	r.closed = true
	if r.err == nil {
		r.err = err
	}
	err = r.r.Close()
	if r.s != nil {
		_ = r.s.Close()
	}
	return err
}

func (r *rows) next() (doClose bool, ok bool) {
	if r.closed {
		return false, false
	}

	r.columns = r.r.Columns()
	if r.current == nil {
		r.current = make([]driver.Value, len(r.columns))
	}

	r.err = r.r.Next(r.current)
	if r.err != nil {
		// Close the connection if there is a driver error.
		if r.err != io.EOF {
			return true, false
		}
		nextResultSet, ok := r.r.(driver.RowsNextResultSet)
		if !ok {
			return true, false
		}
		// The driver is at the end of the current result set.
		// Test to see if there is another result set after the current one.
		// Only close Rows if there is no further result sets to read.
		return !nextResultSet.HasNextResultSet(), false
	}
	return false, true
}

package alphasql

import (
	"database/sql/driver"
	"reflect"
)

// Column is used to provide the details around the columns of the values part of the result set.
type Column struct {
	name              string
	scanType          reflect.Type
	databaseType      string
	length            int64
	hasLength         bool
	nullable          bool
	hasNullable       bool
	precision         int64
	scale             int64
	hasPrecisionScale bool
}

// Name returns the name or alias of the column.
func (c *Column) Name() string {
	return c.name
}

// Length returns the column type length for variable length column types such
// as text and binary field types. If the type length is unbounded the value will
// be [math.MaxInt64] (any database limits will still apply).
// If the column type is not variable length, such as an int, or if not supported
// by the driver ok is false.
func (c *Column) Length() (length int64, ok bool) {
	return c.length, c.hasLength
}

// PrecisionScale returns the scale and precision of a decimal type.
// If not applicable or if not supported ok is false.
func (c *Column) PrecisionScale() (precision, scale int64, ok bool) {
	return c.precision, c.scale, c.hasPrecisionScale
}

// ScanType returns a Go type suitable for scanning into using [Rows.Scan].
// If a driver does not support this property ScanType will return
// the type of empty interface.
func (c *Column) ScanType() reflect.Type {
	return c.scanType
}

// Nullable reports whether the column may be null.
// If a driver does not support this property ok will be false.
func (c *Column) Nullable() (nullable, ok bool) {
	return c.nullable, c.hasNullable
}

// DatabaseTypeName returns the database system name of the column type. If an empty
// string is returned, then the driver type name is not supported.
// Consult your driver documentation for a list of driver data types. [Column.Length] specifiers
// are not included.
// Common type names include "VARCHAR", "TEXT", "NVARCHAR", "DECIMAL", "BOOL",
// "INT", and "BIGINT".
func (c *Column) DatabaseTypeName() string {
	return c.databaseType
}

func getColumnsFromDriverColumns(r driver.Rows) []Column {
	names := r.Columns()
	columns := make([]Column, len(names))
	for i, n := range names {
		c := Column{name: n}
		if st, ok := r.(driver.RowsColumnTypeScanType); ok {
			c.scanType = st.ColumnTypeScanType(i)
		} else {
			c.scanType = reflect.TypeFor[any]()
		}
		if dt, ok := r.(driver.RowsColumnTypeDatabaseTypeName); ok {
			c.databaseType = dt.ColumnTypeDatabaseTypeName(i)
		}
		if cl, ok := r.(driver.RowsColumnTypeLength); ok {
			c.length, c.hasLength = cl.ColumnTypeLength(i)
		}
		if cn, ok := r.(driver.RowsColumnTypeNullable); ok {
			c.nullable, c.hasNullable = cn.ColumnTypeNullable(i)
		}
		if ps, ok := r.(driver.RowsColumnTypePrecisionScale); ok {
			c.precision, c.scale, c.hasPrecisionScale = ps.ColumnTypePrecisionScale(i)
		}
		columns[i] = c
	}
	return columns
}

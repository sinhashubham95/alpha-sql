package alphasql

import (
	"database/sql/driver"
	"errors"
	"unicode"
	"unicode/utf8"
)

// NamedArg is a named argument. NamedArg values may be used as
// arguments to [Connection.Query] or [Connection.QueryRow] or [Connection.Exec]
// and bind to the corresponding named parameter in the SQL statement.
//
// For a more concise way to create NamedArg values, see
// the [Named] function.
type NamedArg struct {
	Name  string
	Value any
}

// Named provides a more concise way to create [NamedArg] values.
//
// Example usage:
//
//	c.Exec(ctx, `
//	    delete from Invoice
//	    where
//	        TimeCreated < @end
//	        and TimeCreated >= @start;`,
//	    sql.Named("start", startTime),
//	    sql.Named("end", endTime),
//	)
func Named(name string, value any) *NamedArg {
	return &NamedArg{Name: name, Value: value}
}

func validateNamedValueName(name string) error {
	if len(name) == 0 {
		return nil
	}
	r, _ := utf8.DecodeRuneInString(name)
	if unicode.IsLetter(r) {
		return nil
	}
	return ErrNamedArgNoLetterBegin
}

// defaultCheckNamedValue wraps the default ColumnConverter to have the same
// function signature as the CheckNamedValue in the driver.NamedValueChecker
// interface.
func defaultCheckNamedValue(nv *driver.NamedValue) (err error) {
	nv.Value, err = driver.DefaultParameterConverter.ConvertValue(nv.Value)
	return err
}

func getDriverNamedValuesFromArgs(c *Connection, args []any) ([]driver.NamedValue, error) {
	nvs := make([]driver.NamedValue, len(args))

	nvc, _ := c.c.(driver.NamedValueChecker)

	// Loop through all the arguments, checking each one.
	// If no error is returned simply increment the index
	// and continue. However, if driver.ErrRemoveArgument
	// is returned the argument is not included in the query
	// argument list.
	n := 0
	for _, a := range args {
		nv := &nvs[n]
		if np, ok := a.(NamedArg); ok {
			if err := validateNamedValueName(np.Name); err != nil {
				return nil, err
			}
			a = np.Value
			nv.Name = np.Name
		}
		nv.Ordinal = n + 1
		nv.Value = a

		// Checking sequence has four routes:
		// A: 1. Default
		// C: 1. NamedValueChecker 2. Default
		checker := defaultCheckNamedValue
		if nvc != nil {
			checker = nvc.CheckNamedValue
		}

		// perform the check
		err := checker(nv)
		switch {
		case err == nil:
			n++
			continue
		case errors.Is(err, driver.ErrRemoveArgument):
			nvs = nvs[:len(nvs)-1]
			continue
		default:
			return nil, ErrConvertingArgumentToNamedArg
		}
	}

	return nvs, nil
}

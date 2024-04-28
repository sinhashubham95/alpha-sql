package alphasql

import "database/sql/driver"

// Result summarizes an executed SQL command.
type Result interface {
	LastInsertID() (int64, error)
	RowsAffected() (int64, error)
}

type result struct {
	r driver.Result
}

func (r *result) LastInsertID() (int64, error) {
	return r.r.LastInsertId()
}

func (r *result) RowsAffected() (int64, error) {
	return r.r.RowsAffected()
}

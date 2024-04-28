package alphasql

// Result summarizes an executed SQL command.
type Result interface {
	LastInsertID() (int64, error)
	RowsAffected() (int64, error)
	Error() error
}

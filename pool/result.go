package pool

type poolErrResult struct {
	err error
}

func (p *poolErrResult) LastInsertID() (int64, error) {
	return 0, p.err
}

func (p *poolErrResult) RowsAffected() (int64, error) {
	return 0, p.err
}

func (p *poolErrResult) Error() error {
	return p.err
}

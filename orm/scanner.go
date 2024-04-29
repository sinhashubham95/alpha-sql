package orm

import (
	"context"
	alphasql "github.com/sinhashubham95/alpha-sql"
)

type scannerRow struct {
	r                        alphasql.Row
	isScanToStructureEnabled bool
}

type scannerRows struct {
	r                        alphasql.Rows
	isScanToStructureEnabled bool
}

func (s *scannerRow) Scan(ctx context.Context, values ...any) error {
	return s.r.Scan(ctx, values...)
}

func (s *scannerRow) ScanStructure(_ context.Context, _ interface{}) error {
	if s.isScanToStructureEnabled {
		return alphasql.ErrScanToStructureNotEnabled
	}
	return nil
}

func (s *scannerRows) Scan(_ context.Context, values ...any) error {
	return s.r.Scan(values...)
}

func (s *scannerRows) ScanStructure(_ context.Context, _ interface{}) error {
	if s.isScanToStructureEnabled {
		return alphasql.ErrScanToStructureNotEnabled
	}
	return nil
}

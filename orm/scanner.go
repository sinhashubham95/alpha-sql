package orm

import (
	"context"
	alphasql "github.com/sinhashubham95/alpha-sql"
)

type scan func(ctx context.Context, values ...any) error

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

func (s *scannerRow) ScanStructure(ctx context.Context, value interface{}) error {
	if s.isScanToStructureEnabled {
		return alphasql.ErrScanToStructureNotEnabled
	}
	return scanToStructure(ctx, s.r.Scan, s.r.Columns(), value)
}

func (s *scannerRows) Scan(_ context.Context, values ...any) error {
	return s.r.Scan(values...)
}

func (s *scannerRows) ScanStructure(ctx context.Context, value interface{}) error {
	if s.isScanToStructureEnabled {
		return alphasql.ErrScanToStructureNotEnabled
	}
	return scanToStructure(ctx, getScan(s.r), s.r.Columns(), value)
}

func getScan(r alphasql.Rows) scan {
	return func(_ context.Context, values ...any) error {
		return r.Scan(values...)
	}
}

func scanToStructure(_ context.Context, _ scan, _ []alphasql.Column, _ interface{}) error {
	return nil
}

package alphasql

import "context"

type Row interface {
	Scan(ctx context.Context, values ...any) error
	Error() error
}

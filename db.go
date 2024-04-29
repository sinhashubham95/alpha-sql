package alphasql

import (
	"context"
	"database/sql/driver"
	"fmt"
	"io"
	"sync/atomic"
)

// DB is the instance that will be used to start new connections.
type DB struct {
	c driver.Connector

	closed               atomic.Bool
	baseAcquireCtx       context.Context
	cancelBaseAcquireCtx context.CancelFunc
}

// Open is used to open a new DB instance to manage the connections.
func Open(ctx context.Context, cfg *ConnectionConfig) (*DB, error) {
	driversMu.RLock()
	d, ok := drivers[cfg.DriverName]
	driversMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("driver %s not registered", cfg.DriverName)
	}
	c, err := d.OpenConnector(cfg.URL)
	if err != nil {
		return nil, err
	}
	baseAcquireCtx, cancelBaseAcquireCtx := context.WithCancel(ctx)
	return &DB{c: c, baseAcquireCtx: baseAcquireCtx, cancelBaseAcquireCtx: cancelBaseAcquireCtx}, nil
}

// Close closes the database and prevents new queries from starting.
// Close then waits for all queries that have started processing on the server
// to finish.
//
// It is rare to Close a [DB], as the [DB] handle is meant to be
// long-lived and shared between many goroutines.
func (db *DB) Close() error {
	if db.closed.CompareAndSwap(false, true) {
		defer db.cancelBaseAcquireCtx()
		if c, ok := db.c.(io.Closer); ok {
			return c.Close()
		}
		return nil
	}
	return ErrDBClosed
}

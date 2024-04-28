package alphasql

import (
	"context"
	"database/sql/driver"
	"fmt"
)

// DB is the instance that will be used to start new connections.
type DB struct {
	c driver.Connector

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

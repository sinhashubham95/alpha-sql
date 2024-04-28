package pool

import (
	"context"
	alphasql "github.com/sinhashubham95/alpha-sql"
	"time"
)

// Config is the configuration required for creating a pool.
type Config struct {
	ConnectionConfig *alphasql.ConnectionConfig

	// BeforeConnect is called before a new Connection is made. It is passed a copy of the underlying
	// alphasql.ConnectionConfig and will not impact any existing open connections.
	BeforeConnect func(context.Context, *alphasql.ConnectionConfig) error

	// AfterConnect is called after a Connection is established, but before it is added to the pool.
	AfterConnect func(context.Context, *alphasql.Connection) error

	// BeforeAcquire is called before a Connection is acquired from the pool. It must return true to allow the
	// acquisition or false to indicate that the Connection should be destroyed and a different Connection should be
	// acquired.
	BeforeAcquire func(context.Context, *Connection) bool

	// AfterRelease is called after a Connection is released, but before it is returned to the pool. It must return true to
	// return the Connection to the pool or false to destroy the Connection.
	AfterRelease func(context.Context, *Connection) bool

	// BeforeClose is called right before a Connection is closed and removed from the pool.
	BeforeClose func(context.Context, *alphasql.Connection)

	// MaxConnectionLifetime is the duration since creation after which a Connection will be automatically closed.
	MaxConnectionLifetime time.Duration

	// MaxConnectionLifetimeJitter is the duration after MaxConnectionLifetime to randomly decide to close a Connection.
	// This helps prevent all connections from being closed at the exact same time, starving the pool.
	MaxConnectionLifetimeJitter time.Duration

	// MaxConnectionIdleTime is the duration after which an idle Connection will be automatically closed by the health check.
	MaxConnectionIdleTime time.Duration

	// MaxConnections is the maximum size of the pool. The default is the greatest of 4 or runtime.NumCPU().
	MaxConnections int32

	// MinConnections is the minimum size of the pool. After Connection closes, the pool might dip below MinConnections.
	// A low number of MinConnections might mean the pool is empty after MaxConnectionLifetime until the health check
	// has a chance to create new connections.
	MinConnections int32

	// HealthCheckPeriod is the duration between checks of the health of idle connections.
	HealthCheckPeriod time.Duration
}

// default functions for pool configs.
var (
	defaultBeforeConnect         = func(_ context.Context, _ *alphasql.ConnectionConfig) error { return nil }
	defaultAfterConnect          = func(_ context.Context, _ *alphasql.Connection) error { return nil }
	defaultBeforeAcquire         = func(_ context.Context, _ *Connection) bool { return true }
	defaultAfterRelease          = func(_ context.Context, _ *Connection) bool { return true }
	defaultBeforeClose           = func(_ context.Context, _ *alphasql.Connection) {}
	defaultMaxConnectionLifetime = time.Hour
	defaultMaxConnectionIdleTime = time.Minute * 30
	defaultMaxConnections        = int32(4)
	defaultMinConnections        = int32(0)
	defaultHealthCheckPeriod     = time.Minute
)

// ValidateAndDefault is used to validate and set the defaults for the mandatory parameters not passed.
func (c *Config) ValidateAndDefault() error {
	if c.ConnectionConfig == nil {
		return alphasql.ErrMissingConnectionConfig
	}
	err := c.ConnectionConfig.ValidateAndDefault()
	if err != nil {
		return err
	}
	if c.BeforeConnect == nil {
		c.BeforeConnect = defaultBeforeConnect
	}
	if c.AfterConnect == nil {
		c.AfterConnect = defaultAfterConnect
	}
	if c.BeforeAcquire == nil {
		c.BeforeAcquire = defaultBeforeAcquire
	}
	if c.AfterRelease == nil {
		c.AfterRelease = defaultAfterRelease
	}
	if c.BeforeClose == nil {
		c.BeforeClose = defaultBeforeClose
	}
	if c.MaxConnectionLifetime == 0 {
		c.MaxConnectionLifetime = defaultMaxConnectionLifetime
	}
	if c.MaxConnectionIdleTime == 0 {
		c.MaxConnectionIdleTime = defaultMaxConnectionIdleTime
	}
	if c.MaxConnections <= 0 {
		c.MaxConnections = defaultMaxConnections
	}
	if c.MinConnections <= 0 {
		c.MinConnections = defaultMinConnections
	}
	if c.HealthCheckPeriod == 0 {
		c.HealthCheckPeriod = defaultHealthCheckPeriod
	}
	return nil
}

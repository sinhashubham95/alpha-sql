package alphasql

import (
	"context"
	"database/sql/driver"
)

// ConnectionConfig is the set of parameters needed to initialise the connection.
type ConnectionConfig struct {
	DriverName string
	URL        string
}

// Connection is used as the connection created.
type Connection struct {
	c driver.Conn
}

// ValidateAndDefault is used to validate and set the defaults for the mandatory parameters not passed.
func (c *ConnectionConfig) ValidateAndDefault() error {
	if c.DriverName == "" {
		return ErrMissingDriverName
	}
	if c.URL == "" {
		return ErrMissingURL
	}
	return nil
}

// Copy is used to copy the connection config.
func (c *ConnectionConfig) Copy() *ConnectionConfig {
	return &ConnectionConfig{DriverName: c.DriverName, URL: c.URL}
}

// Connect is used to create a new connection.
func (db *DB) Connect(ctx context.Context) (*Connection, error) {
	if db.closed.Load() {
		return nil, ErrDBClosed
	}
	c, err := db.c.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return &Connection{c: c}, nil
}

// Connection is used to get the underlying driver connection.
func (c *Connection) Connection() driver.Conn {
	return c.c
}

// Close invalidates and potentially stops any current
// prepared statements and transactions, marking this
// connection as no longer in use.
//
// Because the sql package maintains a free pool of
// connections and only calls Close when there's a surplus of
// idle connections, it shouldn't be necessary for drivers to
// do their own connection caching.
//
// Drivers must ensure all network calls made by Close
// do not block indefinitely (e.g. apply a timeout).
func (c *Connection) Close() error {
	return c.c.Close()
}

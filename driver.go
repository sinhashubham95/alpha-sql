package alphasql

import (
	"database/sql/driver"
	"sync"
)

// registered drivers
var (
	driversMu sync.RWMutex
	drivers   = make(map[string]driver.DriverContext)
)

// RegisterDriver is used to register a driver.
func RegisterDriver(name string, d driver.DriverContext) {
	driversMu.Lock()
	defer driversMu.Unlock()
	drivers[name] = d
}

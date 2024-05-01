package alphasql

// BatchConfig is used as the set of configurations for batch.
type BatchConfig struct {
}

// Batch is used as the set of functionalities for a batch operation on the database.
type Batch interface {
}

type batch struct {
	cfg *BatchConfig
}

// NewBatch is used to provide a way to batch queries to save round-trip.
func (c *Connection) NewBatch(cfg *BatchConfig) (Batch, error) {
	return &batch{cfg: cfg}, nil
}

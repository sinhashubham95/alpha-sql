package alphasql

import "context"

// BatchRemove can be called to remove the corresponding queued operation from the batch.
// It is a noop if the batch is already processing([Batch.Do] has already been called) or
// batch is already closed([Batch.Close] has already been called).
type BatchRemove func()

// BatchConfig is used as the set of configurations for batch.
type BatchConfig struct{}

// Batch is used as the set of functionalities for a batch operation on the database.
type Batch interface {
	// QueueQuery queues a query that returns rows once executed, typically a SELECT.
	// The args are for any placeholder parameters in the query.
	QueueQuery(ctx context.Context, query string, args ...any) BatchRemove

	// QueueQueryRow queues a query that is expected to return at most one row.
	// When executed, it always returns a non-nil value. Errors are deferred until
	// [Row]'s Scan method is called.
	// If the query selects no rows, the [*Row.Scan] will return [ErrNoRows].
	// Otherwise, [*Row.Scan] scans the first selected row and discards
	// the rest.
	QueueQueryRow(ctx context.Context, query string, args ...any) BatchRemove

	// QueueExec queues a query that just executes without returning any rows.
	// The args are for any placeholder parameters in the query.
	QueueExec(ctx context.Context, query string, args ...any) BatchRemove

	// Do is used to execute all the queued queries.
	// It returns the result as an iterator, providing specific methods for all the above operations.
	Do(ctx context.Context) (BatchResults, error)

	// Close is used to close the batch, releasing the connection(making it available for use somewhere else).
	// If Close is called before Do, then all the queued queries will be removed.
	Close(ctx context.Context) error
}

// BatchResults is used as the results for the batch.
type BatchResults interface{}

// BatchOperationMode is used to specify the type of the operation.
type BatchOperationMode string

// batch operation modes
const (
	BatchOperationModeQueryRow BatchOperationMode = "QUERY_ROW"
	BatchOperationModeQuery    BatchOperationMode = "QUERY"
	BatchOperationModeExec     BatchOperationMode = "EXEC"
)

type batchOperation struct {
	id    int
	mode  BatchOperationMode
	query string
	args  []any
}

type batch struct {
	cfg           *BatchConfig
	c             *Connection
	baseCtx       context.Context
	cancelBaseCtx context.CancelFunc
	operations    map[int]batchOperation
	processing    bool
	closed        bool
}

// NewBatch is used to provide a way to batch queries to save round-trip.
func (c *Connection) NewBatch(ctx context.Context, cfg *BatchConfig) (Batch, error) {
	ctx, cancel := context.WithCancel(ctx)
	return &batch{
		cfg:           cfg,
		c:             c,
		baseCtx:       ctx,
		cancelBaseCtx: cancel,
		operations:    make(map[int]batchOperation),
	}, nil
}

func (b *batch) QueueQuery(_ context.Context, query string, args ...any) BatchRemove {
	id := len(b.operations)
	b.operations[id] = batchOperation{
		id:    len(b.operations),
		mode:  BatchOperationModeQuery,
		query: query,
		args:  args,
	}
	return b.getBatchRemoveForID(id)
}

func (b *batch) QueueQueryRow(_ context.Context, query string, args ...any) BatchRemove {
	id := len(b.operations)
	b.operations[id] = batchOperation{
		id:    len(b.operations),
		mode:  BatchOperationModeQueryRow,
		query: query,
		args:  args,
	}
	return b.getBatchRemoveForID(id)
}

func (b *batch) QueueExec(_ context.Context, query string, args ...any) BatchRemove {
	id := len(b.operations)
	b.operations[id] = batchOperation{
		id:    len(b.operations),
		mode:  BatchOperationModeExec,
		query: query,
		args:  args,
	}
	return b.getBatchRemoveForID(id)
}

func (b *batch) Do(_ context.Context) (BatchResults, error) {
	if b.processing {
		return nil, ErrBatchProcessing
	}
	if b.closed {
		return nil, ErrBatchClosed
	}
	return nil, nil
}

func (b *batch) Close(_ context.Context) error {
	if b.processing {
		return ErrBatchProcessing
	}
	if b.closed {
		return ErrBatchClosed
	}
	b.closed = true
	return nil
}

func (b *batch) getBatchRemoveForID(id int) BatchRemove {
	return func() {
		if b.processing || b.closed {
			return
		}
		delete(b.operations, id)
	}
}

package memory

import (
	"context"
	"errors"
	"maps"
)

type txKey struct{}

// WithinTx implements port.Transactor by serializing fn and restoring the rows when fn fails.
func (r *Repo) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if inTx(ctx) {
		return fn(ctx)
	}
	r.txMu.Lock()
	defer r.txMu.Unlock()

	r.mu.RLock()
	before := maps.Clone(r.rows)
	r.mu.RUnlock()

	if err := fn(context.WithValue(ctx, txKey{}, true)); err != nil {
		r.mu.Lock()
		r.rows = before
		r.mu.Unlock()
		return err
	}
	return nil
}

func (r *Repo) write(ctx context.Context, change func() error) error {
	if !inTx(ctx) {
		r.txMu.Lock()
		defer r.txMu.Unlock()
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return change()
}

func inTx(ctx context.Context) bool {
	v, _ := ctx.Value(txKey{}).(bool)
	return v
}

var errNoTx = errors.New("memory: GetWidgetForUpdate outside WithinTx")

package xerrors

import "context"

// callerConfigContextKey is the key type for the caller configuration context.
type callerConfigContextKey struct{}

// ContextWithErrorOptions returns a derived context that carries options to apply to errors.
// Options from any previous [ContextWithErrorOptions] on this context chain are merged before the new opts, so
// repeated calls accumulate rather than replace the full list (later option functions still override the same
// [ErrorOptions] fields when applied in order).
func ContextWithErrorOptions(ctx context.Context, opts ...ErrorOptionFn) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	var merged []ErrorOptionFn
	if prev, ok := ctx.Value(callerConfigContextKey{}).([]ErrorOptionFn); ok && len(prev) > 0 {
		merged = append(merged, prev...)
	}

	merged = append(merged, opts...)

	return context.WithValue(ctx, callerConfigContextKey{}, merged)
}

// ErrorOptionsFromContext returns the list of options attached with [ContextWithErrorOptions] or nil if none
// were set.
func ErrorOptionsFromContext(ctx context.Context) []ErrorOptionFn {
	if ctx == nil {
		return nil
	}

	if v, ok := ctx.Value(callerConfigContextKey{}).([]ErrorOptionFn); ok {
		return v
	}

	return nil
}

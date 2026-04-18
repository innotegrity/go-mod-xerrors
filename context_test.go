package xerrors_test

import (
	"context"
	"testing"

	"go.innotegrity.dev/mod/xerrors"
)

// foreignContextKey is used to test ErrorOptionsFromContext with a mismatched value type.
type foreignContextKey struct{}

func TestContextWithErrorOptions_NilContextUsesBackground(t *testing.T) {
	t.Parallel()

	// Package API explicitly treats nil like Background(); callers should still prefer a real Context.
	ctx := xerrors.ContextWithErrorOptions(
		nil, //nolint:staticcheck // intentional: nil is documented as Background()
		xerrors.WithCaptureCaller(),
	)
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}

	opts := xerrors.ErrorOptionsFromContext(ctx)
	if len(opts) != 1 || opts[0] == nil {
		t.Fatalf("expected one non-nil option, got %#v", opts)
	}
}

func TestErrorOptionsFromContext_NilReturnsNil(t *testing.T) {
	t.Parallel()

	if v := xerrors.ErrorOptionsFromContext(
		nil, //nolint:staticcheck // intentional: test nil handling
	); v != nil {
		t.Fatalf("expected nil, got %#v", v)
	}
}

func TestErrorOptionsFromContext_NoOptionsReturnsNil(t *testing.T) {
	t.Parallel()

	if v := xerrors.ErrorOptionsFromContext(context.Background()); v != nil {
		t.Fatalf("expected nil, got %#v", v)
	}
}

func TestErrorOptionsFromContext_WrongTypeReturnsNil(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), foreignContextKey{}, "not-options")
	if v := xerrors.ErrorOptionsFromContext(ctx); v != nil {
		t.Fatalf("expected nil, got %#v", v)
	}
}

func TestErrorOptionsFromContext_ReturnsAttachedOptions(t *testing.T) {
	t.Parallel()

	fn := xerrors.WithSkipBias(2)
	ctx := xerrors.ContextWithErrorOptions(context.Background(), fn)

	opts := xerrors.ErrorOptionsFromContext(ctx)
	if len(opts) != 1 || opts[0] == nil {
		t.Fatalf("unexpected opts: %#v", opts)
	}
}

func TestContextWithErrorOptions_MergesWithExisting(t *testing.T) {
	t.Parallel()

	ctx := xerrors.ContextWithErrorOptions(context.Background(), xerrors.WithCaptureCaller())
	ctx = xerrors.ContextWithErrorOptions(ctx, xerrors.WithSkipBias(3))

	opts := xerrors.ErrorOptionsFromContext(ctx)
	if len(opts) != 2 {
		t.Fatalf("expected merged option list length 2, got %d (%#v)", len(opts), opts)
	}
}

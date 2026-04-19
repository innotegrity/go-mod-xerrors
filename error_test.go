package xerrors_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"go.innotegrity.dev/mod/xerrors"
)

// Sentinel errors for tests (err113).
var (
	errTestBase  = errors.New("base")
	errTestOther = errors.New("other")
	errTestStd   = errors.New("std")
)

func mustErrPtr(t *testing.T, xe xerrors.Error) *xerrors.XError {
	t.Helper()

	out, ok := xe.(*xerrors.XError)
	if !ok {
		t.Fatalf("expected *xerrors.XError, got %T", xe)
	}

	return out
}

// notFoundTestError embeds [*xerrors.XError] under the promoted field name XError (see package doc).
type notFoundTestError struct{ *xerrors.XError }

func TestNewAs_NewfAs_WrapAs_WrapfAs(t *testing.T) {
	t.Parallel()

	const code = 404

	n1 := xerrors.NewAs(func(e *xerrors.XError) *notFoundTestError {
		return &notFoundTestError{XError: e}
	}, code, "plain")
	if n1.Code() != code || n1.Error() != "plain" {
		t.Fatalf("unexpected NewAs: code=%d msg=%q", n1.Code(), n1.Error())
	}

	n2 := xerrors.NewfAs(func(e *xerrors.XError) *notFoundTestError {
		return &notFoundTestError{XError: e}
	}, code, "n=%d", 42)
	if n2.Code() != code || n2.Error() != "n=42" {
		t.Fatalf("unexpected NewfAs: code=%d msg=%q", n2.Code(), n2.Error())
	}

	n3 := xerrors.WrapAs(func(e *xerrors.XError) *notFoundTestError {
		return &notFoundTestError{XError: e}
	}, errTestBase, code, "wrapped")
	if !n3.Is(errTestBase) || n3.Error() != "wrapped" {
		t.Fatalf("unexpected WrapAs: msg=%q is=%v", n3.Error(), n3.Is(errTestBase))
	}

	n4 := xerrors.WrapfAs(func(e *xerrors.XError) *notFoundTestError {
		return &notFoundTestError{XError: e}
	}, errTestBase, code, "fmt %s", "ok")
	if n4.Error() != "fmt ok" || !n4.Is(errTestBase) {
		t.Fatalf("unexpected WrapfAs: msg=%q", n4.Error())
	}
}

func TestXErrorAs(t *testing.T) {
	t.Parallel()

	base := mustErrPtr(t, xerrors.New(7, "base"))

	out := xerrors.XErrorAs(base, func(e *xerrors.XError) *notFoundTestError {
		return &notFoundTestError{XError: e}
	})
	if out.Code() != 7 || out.Error() != "base" {
		t.Fatalf("unexpected XErrorAs: code=%d msg=%q", out.Code(), out.Error())
	}
}

func TestNewfAs_NoArgsUsesLiteralFormat(t *testing.T) {
	t.Parallel()

	const want = "literal plaintext without fmt verbs"

	got := xerrors.NewfAs(func(e *xerrors.XError) *notFoundTestError {
		return &notFoundTestError{XError: e}
	}, 201, want)
	if got.Error() != want {
		t.Fatalf("got %q, want %q", got.Error(), want)
	}
}

func TestWrapfAs_NoArgsUsesLiteralMessage(t *testing.T) {
	t.Parallel()

	const want = "literal plaintext without fmt verbs"

	got := xerrors.WrapfAs(func(e *xerrors.XError) *notFoundTestError {
		return &notFoundTestError{XError: e}
	}, errTestBase, 202, want)
	if got.Error() != want {
		t.Fatalf("got %q, want %q", got.Error(), want)
	}
}

func TestConstruct_ErrNil_NewfPath(t *testing.T) {
	t.Parallel()

	got := xerrors.Construct(context.Background(), nil, 101, "plain")
	if got.Code() != 101 || got.Error() != "plain" {
		t.Fatalf("unexpected Construct: code=%d msg=%q", got.Code(), got.Error())
	}

	if got.Unwrap() != nil {
		t.Fatalf("Unwrap: got %v, want nil", got.Unwrap())
	}
}

func TestConstruct_ErrNil_FormattedMessage(t *testing.T) {
	t.Parallel()

	e := xerrors.Construct(context.Background(), nil, 102, "k=%d", 9)
	if e.Code() != 102 || e.Error() != "k=9" {
		t.Fatalf("unexpected Construct: code=%d msg=%q", e.Code(), e.Error())
	}
}

func TestConstruct_ErrNil_NoArgsUsesLiteralFormat(t *testing.T) {
	t.Parallel()

	const want = "literal plaintext without fmt verbs"

	e := xerrors.Construct(context.Background(), nil, 103, want)
	if e.Error() != want {
		t.Fatalf("got %q, want %q", e.Error(), want)
	}
}

func TestConstruct_ErrNotNil_WrapfPath(t *testing.T) {
	t.Parallel()

	got := xerrors.Construct(context.Background(), errTestBase, 201, "outer")
	if got.Code() != 201 || got.Error() != "outer" {
		t.Fatalf("unexpected Construct: code=%d msg=%q", got.Code(), got.Error())
	}

	if !got.Is(errTestBase) {
		t.Fatal("expected Is(errTestBase)")
	}

	if !errors.Is(got.Unwrap(), errTestBase) {
		t.Fatalf("Unwrap: got %v", got.Unwrap())
	}
}

func TestConstruct_ErrNotNil_FormattedOuterMessage(t *testing.T) {
	t.Parallel()

	e := xerrors.Construct(context.Background(), errTestBase, 202, "tag=%s", "v")
	if e.Error() != "tag=v" {
		t.Fatalf("unexpected message: %q", e.Error())
	}
}

func TestConstruct_ErrNotNil_NoArgsUsesLiteralMessage(t *testing.T) {
	t.Parallel()

	const want = "literal plaintext without fmt verbs"

	e := xerrors.Construct(context.Background(), errTestBase, 203, want)
	if e.Error() != want {
		t.Fatalf("got %q, want %q", e.Error(), want)
	}
}

func TestConstruct_AppliesOptionsFromContext_NilErr(t *testing.T) {
	t.Parallel()

	workDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	prefix := workDir + string(os.PathSeparator)

	ctx := xerrors.ContextWithErrorOptions(context.Background(),
		xerrors.WithCaptureCaller(),
		xerrors.WithStripFilePrefixes(prefix),
	)

	e := xerrors.Construct(ctx, nil, 301, "with-ctx")

	caller := e.Caller()
	if caller.File == xerrors.UnknownCallerFile || caller.Line <= 0 || caller.Func == xerrors.UnknownCallerFunc {
		t.Fatalf("expected captured caller, got %+v", caller)
	}

	if strings.HasPrefix(caller.File, workDir) {
		t.Fatalf("expected stripped file path, got %q", caller.File)
	}
}

func TestConstruct_AppliesOptionsFromContext_WrappedErr(t *testing.T) {
	t.Parallel()

	workDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	prefix := workDir + string(os.PathSeparator)

	ctx := xerrors.ContextWithErrorOptions(context.Background(),
		xerrors.WithCaptureCaller(),
		xerrors.WithStripFilePrefixes(prefix),
	)

	e := xerrors.Construct(ctx, errTestBase, 302, "wrapped-via-construct")

	caller := e.Caller()
	if caller.File == xerrors.UnknownCallerFile || caller.Line <= 0 || caller.Func == xerrors.UnknownCallerFunc {
		t.Fatalf("expected captured caller, got %+v", caller)
	}

	if strings.HasPrefix(caller.File, workDir) {
		t.Fatalf("expected stripped file path, got %q", caller.File)
	}
}

func TestNew_Newf_Code_Error(t *testing.T) {
	t.Parallel()

	e1 := xerrors.New(100, "plain message")
	if e1.Code() != 100 || e1.Error() != "plain message" {
		t.Fatalf("unexpected New result: code=%d msg=%q", e1.Code(), e1.Error())
	}

	e2 := xerrors.Newf(200, "n=%d", 42)
	if e2.Code() != 200 || e2.Error() != "n=42" {
		t.Fatalf("unexpected Newf result: code=%d msg=%q", e2.Code(), e2.Error())
	}
}

func TestWrap_Wrapf_Is(t *testing.T) {
	t.Parallel()

	e1 := xerrors.New(1, "top-level")
	if e1.Is(errTestBase) {
		t.Fatal("expected Is=false without wrapped error")
	}

	e2 := xerrors.Wrap(errTestBase, 2, "wrapped")
	if !e2.Is(errTestBase) {
		t.Fatal("expected Is=true for wrapped base error")
	}

	if e2.Is(errTestOther) {
		t.Fatal("expected Is=false for unrelated error")
	}

	e3 := xerrors.Wrapf(errTestBase, 3, "fmt %s", "ok")
	if e3.Error() != "fmt ok" {
		t.Fatalf("unexpected Wrapf message: %q", e3.Error())
	}
}

func TestUnwrap(t *testing.T) {
	t.Parallel()

	plain := xerrors.New(1, "plain")
	if plain.Unwrap() != nil {
		t.Fatalf("Unwrap without wrap: got %v, want nil", plain.Unwrap())
	}

	wrapped := xerrors.Wrap(errTestBase, 2, "outer")
	if !errors.Is(wrapped.Unwrap(), errTestBase) {
		t.Fatalf("Unwrap after Wrap: got %v, want errTestBase", wrapped.Unwrap())
	}

	wrappedf := xerrors.Wrapf(errTestBase, 3, "fmt %d", 1)
	if !errors.Is(wrappedf.Unwrap(), errTestBase) {
		t.Fatalf("Unwrap after Wrapf: got %v, want errTestBase", wrappedf.Unwrap())
	}

	inner := xerrors.New(10, "inner")

	chain := xerrors.Wrap(inner, 11, "mid")
	if !errors.Is(chain.Unwrap(), inner) {
		t.Fatalf("Unwrap nested *XError: got %v, want inner", chain.Unwrap())
	}
}

func TestAttrs_WithAttr_WithAttrs(t *testing.T) {
	t.Parallel()

	errAttrs := xerrors.New(1, "attrs")
	if errAttrs.Attrs() != nil {
		t.Fatalf("expected nil attrs initially, got %#v", errAttrs.Attrs())
	}

	errAttrs = errAttrs.WithAttr("a", 1)
	errAttrs = errAttrs.WithAttrs(map[string]any{"b": "two"})

	attrs := errAttrs.Attrs()
	if attrs["a"] != 1 || attrs["b"] != "two" {
		t.Fatalf("unexpected attrs: %#v", attrs)
	}

	errAttrs = errAttrs.WithAttrs(map[string]any{"c": 3})
	if errAttrs.Attrs()["c"] != 3 {
		t.Fatalf("expected merged attrs, got %#v", errAttrs.Attrs())
	}
}

func TestAttrs_ReturnedMapIsCopy(t *testing.T) {
	t.Parallel()

	errAttrs := mustErrPtr(t, xerrors.New(1, "x").WithAttr("k", 1))

	snapshot := errAttrs.Attrs()
	snapshot["k"] = 999

	if errAttrs.Attrs()["k"] != 1 {
		t.Fatalf("mutating Attrs() return should not change error: got %#v", errAttrs.Attrs())
	}
}

func TestMarshalJSON_WrappedStdError(t *testing.T) {
	t.Parallel()

	e := xerrors.Wrap(errTestStd, 10, "outer")

	raw, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out map[string]any

	err = json.Unmarshal(raw, &out)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	wrapped, ok := out["wrappedError"].(map[string]any)
	if !ok {
		t.Fatalf("wrappedError missing or wrong type: %#v", out["wrappedError"])
	}

	const jsonKindStd = "std"

	if wrapped["kind"] != jsonKindStd {
		t.Fatalf("expected kind %s, got %#v", jsonKindStd, wrapped["kind"])
	}

	if wrapped["message"] != "std" {
		t.Fatalf("unexpected wrapped message: %#v", wrapped["message"])
	}
}

func TestMarshalJSON_WrappedXError(t *testing.T) {
	t.Parallel()

	inner := xerrors.New(5, "inner")
	outer := xerrors.Wrap(inner, 6, "outer")

	raw, err := json.Marshal(outer)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out map[string]any

	err = json.Unmarshal(raw, &out)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	wrapped, hasWrapped := out["wrappedError"].(map[string]any)
	if !hasWrapped {
		t.Fatalf("wrappedError missing: %#v", out["wrappedError"])
	}

	const jsonKindNested = "xerrors.Error"

	if wrapped["kind"] != jsonKindNested {
		t.Fatalf("expected kind %s, got %#v", jsonKindNested, wrapped["kind"])
	}

	nested, ok := wrapped["nested"].(map[string]any)
	if !ok {
		t.Fatalf("nested missing or wrong type: %#v", wrapped["nested"])
	}

	if nested["message"] != "inner" {
		t.Fatalf("unexpected inner: %#v", nested["message"])
	}
}

func TestMarshalJSON_AttrsCopyIsolation(t *testing.T) {
	t.Parallel()

	attrs := map[string]any{"k1": "v1"}
	outer := xerrors.Wrap(xerrors.New(1, "in"), 2, "out").WithAttrs(attrs).WithAttr("k2", "v2")
	attrs["k1"] = "mutated"

	raw, err := json.Marshal(outer)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out map[string]any

	err = json.Unmarshal(raw, &out)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	m, ok := out["attrs"].(map[string]any)
	if !ok || m["k1"] != "v1" || m["k2"] != "v2" {
		t.Fatalf("unexpected attrs in JSON: %#v", out["attrs"])
	}
}

func TestMarshalJSON_MarshalFailure(t *testing.T) {
	t.Parallel()

	e := xerrors.New(1, "msg").WithAttr("bad", make(chan int))

	_, err := e.MarshalJSON()
	if err == nil || !strings.Contains(err.Error(), "marshal extended error to JSON") {
		t.Fatalf("expected marshal error, got %v", err)
	}
}

func TestMarshalJSON_WrappedInnerMarshalFailure(t *testing.T) {
	t.Parallel()

	inner := mustErrPtr(t, xerrors.New(1, "inner").WithAttr("bad", make(chan int)))
	outer := mustErrPtr(t, xerrors.Wrap(inner, 2, "outer"))

	_, err := outer.MarshalJSON()
	if err == nil || !strings.Contains(err.Error(), "marshal extended error to JSON") {
		t.Fatalf("expected marshal error from nested *Error, got %v", err)
	}
}

func TestString_SuccessAndMarshalFailure(t *testing.T) {
	t.Parallel()

	ok := xerrors.New(77, "hello").WithAttr("k", "v")

	str := ok.String()
	if !json.Valid([]byte(str)) {
		t.Fatalf("expected JSON string, got %q", str)
	}

	bad := xerrors.New(1, "x").WithAttr("bad", make(chan int))

	str2 := bad.String()
	if !strings.HasPrefix(str2, "failed to marshal error to JSON:") {
		t.Fatalf("expected failure prefix, got %q", str2)
	}
}

func TestWithSkipBias_WithStripFilePrefixes_Chained(t *testing.T) {
	t.Parallel()

	_ = xerrors.New(1, "m").WithSkipBias(1).WithStripFilePrefixes("/tmp/")
}

func TestWithOptionsFromContext_SecondCallDoesNotRecapture(t *testing.T) {
	t.Parallel()

	ctx := xerrors.ContextWithErrorOptions(context.Background(), xerrors.WithCaptureCaller())
	errRecapture := xerrors.New(1, "x").WithOptionsFromContext(ctx)
	callerFirst := errRecapture.Caller()
	errRecapture = errRecapture.WithOptionsFromContext(ctx)

	callerSecond := errRecapture.Caller()
	if callerFirst != callerSecond {
		t.Fatalf(
			"expected same caller after second WithOptionsFromContext, got %+v vs %+v",
			callerFirst,
			callerSecond,
		)
	}
}

func TestWithCaller_SecondCallDoesNotRecapture(t *testing.T) {
	t.Parallel()

	errSecond := xerrors.New(1, "x").WithCaller()
	callerFirst := errSecond.Caller()

	errSecond = errSecond.WithCaller()

	callerSecond := errSecond.Caller()
	if callerFirst != callerSecond {
		t.Fatalf(
			"expected same caller after second WithCaller, got %+v vs %+v",
			callerFirst,
			callerSecond,
		)
	}
}

// xerrorsDeepEqual compares two [*xerrors.XError] values produced by JSON round-trips or equivalent construction.
func xerrorsDeepEqual(want, got *xerrors.XError) bool {
	if want.Code() != got.Code() || want.Error() != got.Error() {
		return false
	}

	if !reflect.DeepEqual(want.Attrs(), got.Attrs()) {
		return false
	}

	if want.Caller() != got.Caller() {
		return false
	}

	return unwrapEqual(want.Unwrap(), got.Unwrap())
}

func unwrapEqual(want, got error) bool {
	if want == nil && got == nil {
		return true
	}

	if want == nil || got == nil {
		return false
	}

	var wantX, gotX *xerrors.XError
	if errors.As(want, &wantX) && errors.As(got, &gotX) {
		return xerrorsDeepEqual(wantX, gotX)
	}

	return want.Error() == got.Error()
}

func roundTripJSON(t *testing.T, orig *xerrors.XError) *xerrors.XError {
	t.Helper()

	raw, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out xerrors.XError

	err = json.Unmarshal(raw, &out)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	return &out
}

func TestUnmarshalJSON_RoundTripPlain(t *testing.T) {
	t.Parallel()

	want := mustErrPtr(t, xerrors.New(42, "hello"))

	got := roundTripJSON(t, want)
	if !xerrorsDeepEqual(want, got) {
		t.Fatalf("round-trip mismatch\nwant: %#v\ngot:  %#v", want, got)
	}
}

func TestUnmarshalJSON_RoundTripAttrsAndCaller(t *testing.T) {
	t.Parallel()

	want := mustErrPtr(t, xerrors.New(3, "with-meta"))
	want = mustErrPtr(t, want.WithAttr("k", "v").WithAttrs(map[string]any{"n": float64(1)}))
	want = mustErrPtr(t, want.WithStripFilePrefixes("/tmp/").WithCaller())

	got := roundTripJSON(t, want)
	if !xerrorsDeepEqual(want, got) {
		t.Fatalf("round-trip mismatch\nwant: %#v\ngot:  %#v", want, got)
	}
}

func TestUnmarshalJSON_RoundTripNestedXError(t *testing.T) {
	t.Parallel()

	inner := mustErrPtr(t, xerrors.New(5, "inner"))
	want := mustErrPtr(t, xerrors.Wrap(inner, 6, "outer"))

	got := roundTripJSON(t, want)
	if !xerrorsDeepEqual(want, got) {
		t.Fatalf("round-trip mismatch\nwant: %#v\ngot:  %#v", want, got)
	}
}

func TestUnmarshalJSON_RoundTripStdWrapped(t *testing.T) {
	t.Parallel()

	want := mustErrPtr(t, xerrors.Wrap(errTestStd, 9, "outer"))

	got := roundTripJSON(t, want)
	if !xerrorsDeepEqual(want, got) {
		t.Fatalf("round-trip mismatch\nwant: %#v\ngot:  %#v", want, got)
	}
}

func TestUnmarshalJSON_InvalidJSON(t *testing.T) {
	t.Parallel()

	var dest xerrors.XError

	err := json.Unmarshal([]byte(`not json`), &dest)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUnmarshalJSON_DirectCallInvalidPayload(t *testing.T) {
	t.Parallel()

	var dest xerrors.XError

	err := (&dest).UnmarshalJSON([]byte(`not json`))
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "unmarshal extended error from JSON") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnmarshalJSON_WrappedErrorNotObject(t *testing.T) {
	t.Parallel()

	var dest xerrors.XError

	err := json.Unmarshal([]byte(`{"code":1,"message":"x","wrappedError":[]}`), &dest)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUnmarshalJSON_WrappedInnerInvalid(t *testing.T) {
	t.Parallel()

	var dest xerrors.XError

	// wrappedError is not a valid [wrappedEnvelope]; missing kind (and invalid nested JSON for this schema).
	raw := `{"code":1,"message":"outer","wrappedError":{"code":2,"message":true}}`

	err := json.Unmarshal([]byte(raw), &dest)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUnmarshalJSON_WrappedStdInvalid(t *testing.T) {
	t.Parallel()

	var dest xerrors.XError

	// Invalid envelope: "message" must be a string for unmarshaling into [wrappedEnvelope].
	raw := `{"code":1,"message":"outer","wrappedError":{"kind":"std","message":{}}}`

	err := json.Unmarshal([]byte(raw), &dest)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUnmarshalJSON_MissingWrappedKind(t *testing.T) {
	t.Parallel()

	// Old wire shape without tagged union is rejected.
	raw := `{"code":1,"message":"outer","wrappedError":{"code":2,"message":"inner"}}`

	var dest xerrors.XError

	err := json.Unmarshal([]byte(raw), &dest)
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), `wrappedError must include a valid "kind"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnmarshalJSON_UnknownWrappedKind(t *testing.T) {
	t.Parallel()

	raw := `{"code":1,"message":"x","wrappedError":{"kind":"other","message":"m"}}`

	var dest xerrors.XError

	err := json.Unmarshal([]byte(raw), &dest)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUnmarshalJSON_XErrorKindMissingNested(t *testing.T) {
	t.Parallel()

	raw := `{"code":1,"message":"x","wrappedError":{"kind":"xerrors.Error"}}`

	var dest xerrors.XError

	err := json.Unmarshal([]byte(raw), &dest)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUnmarshalJSON_XErrorNestedInnerInvalid(t *testing.T) {
	t.Parallel()

	// [wrappedEnvelope.nested] decodes, but inner [*XError.UnmarshalJSON] fails on the nested object.
	raw := `{"code":1,"message":"outer","wrappedError":{"kind":"xerrors.Error","nested":{"code":1,"message":true}}}`

	var dest xerrors.XError

	err := json.Unmarshal([]byte(raw), &dest)
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "unmarshal extended error from JSON") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnmarshalJSON_ResetsReceiver(t *testing.T) {
	t.Parallel()

	dest := *mustErrPtr(t, xerrors.New(1, "old").WithAttr("x", 1))

	raw, err := json.Marshal(xerrors.New(2, "new"))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	err = json.Unmarshal(raw, &dest)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if dest.Code() != 2 || dest.Error() != "new" {
		t.Fatalf("unexpected dest: code=%d msg=%q", dest.Code(), dest.Error())
	}

	if dest.Attrs() != nil {
		t.Fatalf("attrs should be cleared, got %#v", dest.Attrs())
	}
}

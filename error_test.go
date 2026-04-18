package xerrors_test

import (
	"context"
	"encoding/json"
	"errors"
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

func mustErrPtr(t *testing.T, xe xerrors.XError) *xerrors.Error {
	t.Helper()

	out, ok := xe.(*xerrors.Error)
	if !ok {
		t.Fatalf("expected *xerrors.Error, got %T", xe)
	}

	return out
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
		t.Fatalf("Unwrap nested *Error: got %v, want inner", chain.Unwrap())
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

	if wrapped["kind"] != "xerror" {
		t.Fatalf("expected kind xerror, got %#v", wrapped["kind"])
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

// xerrorsDeepEqual compares two [*xerrors.Error] values produced by JSON round-trips or equivalent construction.
func xerrorsDeepEqual(want, got *xerrors.Error) bool {
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

	var wantX, gotX *xerrors.Error
	if errors.As(want, &wantX) && errors.As(got, &gotX) {
		return xerrorsDeepEqual(wantX, gotX)
	}

	return want.Error() == got.Error()
}

func roundTripJSON(t *testing.T, orig *xerrors.Error) *xerrors.Error {
	t.Helper()

	raw, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out xerrors.Error

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

	var dest xerrors.Error

	err := json.Unmarshal([]byte(`not json`), &dest)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUnmarshalJSON_DirectCallInvalidPayload(t *testing.T) {
	t.Parallel()

	var dest xerrors.Error

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

	var dest xerrors.Error

	err := json.Unmarshal([]byte(`{"code":1,"message":"x","wrappedError":[]}`), &dest)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUnmarshalJSON_WrappedInnerInvalid(t *testing.T) {
	t.Parallel()

	var dest xerrors.Error

	// wrappedError is not a valid [wrappedEnvelope]; missing kind (and invalid nested JSON for this schema).
	raw := `{"code":1,"message":"outer","wrappedError":{"code":2,"message":true}}`

	err := json.Unmarshal([]byte(raw), &dest)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUnmarshalJSON_WrappedStdInvalid(t *testing.T) {
	t.Parallel()

	var dest xerrors.Error

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

	var dest xerrors.Error

	err := json.Unmarshal([]byte(raw), &dest)
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), `wrappedError must include "kind"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnmarshalJSON_UnknownWrappedKind(t *testing.T) {
	t.Parallel()

	raw := `{"code":1,"message":"x","wrappedError":{"kind":"other","message":"m"}}`

	var dest xerrors.Error

	err := json.Unmarshal([]byte(raw), &dest)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUnmarshalJSON_XErrorKindMissingNested(t *testing.T) {
	t.Parallel()

	raw := `{"code":1,"message":"x","wrappedError":{"kind":"xerror"}}`

	var dest xerrors.Error

	err := json.Unmarshal([]byte(raw), &dest)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUnmarshalJSON_XErrorNestedInnerInvalid(t *testing.T) {
	t.Parallel()

	// [wrappedEnvelope.nested] decodes, but inner [*Error.UnmarshalJSON] fails on the nested object.
	raw := `{"code":1,"message":"outer","wrappedError":{"kind":"xerror","nested":{"code":1,"message":true}}}`

	var dest xerrors.Error

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

package xerrors_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"go.innotegrity.dev/mod/xerrors"
)

func TestMarshalJSON_ErrorChain(t *testing.T) {
	err := errors.New("this is a regular error")
	xerr1 := xerrors.Wrap(1, err, "this is error 1 with no attributes")
	xerr2 := xerrors.Wrap(2, xerr1, "this is error 2 which wraps error 1").WithAttr("xerr2_a1", "b")
	xerr3 := xerrors.Wrap(3, xerr2, "this is error 3 which wraps error 2").WithAttrs(map[string]any{
		"xerr3_a1": "b",
		"xerr3_a2": "c",
	})
	xerr4 := xerrors.Wrap(4, xerr3, "this is error 4 which wraps error 3").WithAttr("xerr4_a1", "b")

	e1, err := json.Marshal(xerr1)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("e1: %s", string(e1))

	e2, err := json.Marshal(xerr2)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("e2: %s", string(e2))

	e3, err := json.Marshal(xerr3)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("e3: %s", string(e3))

	e4, err := json.Marshal(xerr4)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("e4: %s", string(e4))
}

func TestNew_Newf(t *testing.T) {
	xerrors.CaptureCallerInfo(false)
	xerrors.StripCallerFilePrefixes()

	e1 := xerrors.New(100, "plain message")
	if e1.Code() != 100 {
		t.Fatalf("expected code 100, got %d", e1.Code())
	}
	if e1.Error() != "plain message" {
		t.Fatalf("unexpected message: %q", e1.Error())
	}

	e2 := xerrors.Newf(200, "value=%d", 42)
	if e2.Error() != "value=42" {
		t.Fatalf("unexpected formatted message: %q", e2.Error())
	}
}

func TestWrap_Wrapf_Is(t *testing.T) {
	base := errors.New("base")

	e1 := xerrors.New(1, "top-level")
	if e1.Is(base) {
		t.Fatal("expected Is=false for non-wrapped error")
	}

	e2 := xerrors.Wrap(2, base, "wrapped")
	if !e2.Is(base) {
		t.Fatal("expected Is=true for wrapped base error")
	}
	if e2.Is(errors.New("other")) {
		t.Fatal("expected Is=false for unrelated error")
	}

	e3 := xerrors.Wrapf(3, base, "wrapped %s", "formatted")
	if e3.Error() != "wrapped formatted" {
		t.Fatalf("unexpected Wrapf message: %q", e3.Error())
	}
}

func TestMarshalJSON_WithStandardWrappedError(t *testing.T) {
	e := xerrors.Wrap(10, errors.New("std"), "wrapped std")
	raw, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	wrapped, ok := out["wrapped_error"].(map[string]any)
	if !ok {
		t.Fatalf("wrapped_error not found or wrong type: %#v", out["wrapped_error"])
	}
	if wrapped["message"] != "std" {
		t.Fatalf("expected wrapped std message, got %#v", wrapped["message"])
	}
}

func TestMarshalJSON_WithXErrorWrappedAndAttrs(t *testing.T) {
	inner := xerrors.New(5, "inner")
	attrs := map[string]any{"k1": "v1"}
	outer := xerrors.Wrap(6, inner, "outer").WithAttrs(attrs).WithAttr("k2", "v2")
	attrs["k1"] = "changed-after-copy"

	raw, err := json.Marshal(outer)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	wrapped, ok := out["wrapped_error"].(map[string]any)
	if !ok {
		t.Fatalf("wrapped_error not found or wrong type: %#v", out["wrapped_error"])
	}
	if wrapped["message"] != "inner" {
		t.Fatalf("expected wrapped xerror message, got %#v", wrapped["message"])
	}

	outAttrs, ok := out["attrs"].(map[string]any)
	if !ok {
		t.Fatalf("attrs not found or wrong type: %#v", out["attrs"])
	}
	if outAttrs["k1"] != "v1" || outAttrs["k2"] != "v2" {
		t.Fatalf("unexpected attrs: %#v", outAttrs)
	}
}

func TestString_MarshalFailurePath(t *testing.T) {
	e := xerrors.New(1, "msg").WithAttr("bad", func() {})
	s := e.String()
	if !strings.Contains(s, "failed to marshal error to JSON:") {
		t.Fatalf("expected marshal failure message, got %q", s)
	}
}

func TestError_ReturnsMessage(t *testing.T) {
	e := xerrors.Wrap(9, errors.New("root"), "outer message")
	if e.Error() != "outer message" {
		t.Fatalf("expected Error() to return outer message, got %q", e.Error())
	}
}

func TestString_SuccessPath(t *testing.T) {
	e := xerrors.New(77, "string message").WithAttr("k", "v")
	s := e.String()
	if !json.Valid([]byte(s)) {
		t.Fatalf("expected valid JSON string output, got %q", s)
	}

	var out map[string]any
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		t.Fatalf("failed to unmarshal String() output: %v", err)
	}
	if out["message"] != "string message" {
		t.Fatalf("unexpected message in String() output: %#v", out["message"])
	}
	if out["code"] != float64(77) {
		t.Fatalf("unexpected code in String() output: %#v", out["code"])
	}
}

func TestAttrs_ReturnsCurrentAttributes(t *testing.T) {
	e := xerrors.New(1, "attrs")
	if e.Attrs() != nil {
		t.Fatalf("expected nil attrs before adding attributes, got %#v", e.Attrs())
	}

	e.WithAttr("a", 1).WithAttrs(map[string]any{"b": "two"})
	attrs := e.Attrs()
	if attrs["a"] != 1 || attrs["b"] != "two" {
		t.Fatalf("unexpected attrs map from Attrs(): %#v", attrs)
	}
}

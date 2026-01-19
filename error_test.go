package xerrors_test

import (
	"encoding/json"
	"errors"
	"testing"

	"pkg.innotegrity.dev/go/mod/xerrors"
)

func TestError(t *testing.T) {
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

package xerrors_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"go.innotegrity.dev/mod/xerrors"
)

func TestCaller_DefaultWhenCaptureDisabled(t *testing.T) {
	xerrors.CaptureCallerInfo(false)
	xerrors.StripCallerFilePrefixes()

	e := xerrors.New(1, "msg")
	c := e.Caller()

	if c.File != "???" || c.Line != 0 || c.Func != "???" {
		t.Fatalf("expected default caller info, got %+v", c)
	}
}

func TestCaller_CapturedAndPrefixStripped(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	xerrors.CaptureCallerInfo(true)
	xerrors.StripCallerFilePrefixes(wd + "/")
	t.Cleanup(func() {
		xerrors.CaptureCallerInfo(false)
		xerrors.StripCallerFilePrefixes()
	})

	e := xerrors.New(1, "msg")
	c := e.Caller()

	if c.File == "???" || c.Line == 0 || c.Func == "???" {
		t.Fatalf("expected caller info to be captured, got %+v", c)
	}
	if strings.HasPrefix(c.File, wd+"/") {
		t.Fatalf("expected prefix %q to be stripped from %q", wd+"/", c.File)
	}
	if !strings.HasSuffix(c.File, "_test.go") {
		t.Fatalf("expected test file suffix in caller file, got %q", c.File)
	}
}

func TestGetCallerInfo_ReturnsCallerDetails(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	xerrors.StripCallerFilePrefixes(wd + "/")
	t.Cleanup(func() {
		xerrors.StripCallerFilePrefixes()
	})

	c := helperGetCallerInfo()
	if c.File == "???" || c.Line == 0 || c.Func == "???" {
		t.Fatalf("expected non-default caller info, got %+v", c)
	}
	if strings.HasPrefix(c.File, wd+"/") {
		t.Fatalf("expected stripped file prefix, got %q", c.File)
	}
}

func TestGetCallerInfo_RuntimeCallerFailure(t *testing.T) {
	// Use an extreme skip value so runtime.Caller cannot resolve a stack frame.
	c := xerrors.GetCallerInfo(1 << 20)
	if c.File != "???" || c.Line != 0 || c.Func != "???" {
		t.Fatalf("expected default caller info when runtime.Caller fails, got %+v", *c)
	}
}

func helperGetCallerInfo() xerrors.CallerInfo {
	return *xerrors.GetCallerInfo(0)
}

func TestCaller_CapturedForNewfWrapAndWrapf(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	xerrors.CaptureCallerInfo(true)
	xerrors.StripCallerFilePrefixes(wd + "/")
	t.Cleanup(func() {
		xerrors.CaptureCallerInfo(false)
		xerrors.StripCallerFilePrefixes()
	})

	base := errors.New("base")
	tests := []struct {
		name string
		make func() xerrors.Error
	}{
		{
			name: "newf",
			make: func() xerrors.Error {
				return xerrors.Newf(1, "formatted %d", 7)
			},
		},
		{
			name: "wrap",
			make: func() xerrors.Error {
				return xerrors.Wrap(2, base, "wrapped")
			},
		},
		{
			name: "wrapf",
			make: func() xerrors.Error {
				return xerrors.Wrapf(3, base, "wrapped %s", "formatted")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := tc.make()
			c := e.Caller()
			if c.File == "???" || c.Line == 0 || c.Func == "???" {
				t.Fatalf("expected captured caller info, got %+v", c)
			}
			if strings.HasPrefix(c.File, wd+"/") {
				t.Fatalf("expected stripped file prefix in caller file, got %q", c.File)
			}
		})
	}
}

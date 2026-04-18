package xerrors

import (
	"runtime"
	"strings"
)

const (
	// UnknownCallerFile is the string used to indicate that the caller file is unknown.
	UnknownCallerFile = "<unknown>"

	// UnknownCallerFunc is the string used to indicate that the caller function is unknown.
	UnknownCallerFunc = "<unknown>"
)

// CallerInfo holds information about the location from which the error was generated.
type CallerInfo struct {
	// File is the name of the file in which the error occurred, relative to the package root.
	File string `json:"file"`

	// Line is the line number at which the error occurred.
	Line int `json:"line"`

	// Func is the name of the function in which the error occurred.
	Func string `json:"func"`
}

// DefaultCallerInfo returns a default [CallerInfo] struct with unknown caller file, line, and function.
func DefaultCallerInfo() *CallerInfo {
	return &CallerInfo{
		File: UnknownCallerFile,
		Line: -1,
		Func: UnknownCallerFunc,
	}
}

// getCallerInfo retrieves the caller information from the runtime stack.
//
// The 'runtimeSkip' parameter indicates how many stack frames to ascend with 0 being the immediate caller of
// this function.
//
// The 'stripPrefixes' parameter is a list of file path prefixes to strip from the caller file path.
//
// The function returns a [CallerInfo] object containing the caller information.
//
// If the caller information is not available, a default [CallerInfo] is returned.
//
// If the caller information is available, the function returns a [CallerInfo] object containing the caller information.
func getCallerInfo(runtimeSkip int, stripPrefixes []string) *CallerInfo {
	pc, _, _, ok := runtime.Caller(1 + runtimeSkip) // skip the getCallerInfo call itself
	if !ok {
		return DefaultCallerInfo()
	}

	frames := runtime.CallersFrames([]uintptr{pc})
	fr, _ := frames.Next()

	file := fr.File
	line := fr.Line
	fnName := fr.Function

	for _, prefix := range stripPrefixes {
		if strings.HasPrefix(file, prefix) {
			file = file[len(prefix):]

			break
		}
	}

	return &CallerInfo{
		File: file,
		Line: line,
		Func: fnName,
	}
}

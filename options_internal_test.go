package xerrors

import "testing"

func TestErrorOptionFns(t *testing.T) {
	t.Parallel()

	var opts ErrorOptions
	WithCaptureCaller()(&opts)
	WithSkipBias(9)(&opts)
	WithStripFilePrefixes("/a/", "/b/")(&opts)
}

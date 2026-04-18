// Package xerrors provides structured errors with numeric codes, optional attributes, JSON serialization, and
// optional caller (file, line, function) recording at the point the error is built.
//
// # Basic construction
//
// Use [New] for a plain message and [Newf] when the message is formatted:
//
//	err := xerrors.New(404, "not found")
//	err = xerrors.Newf(500, "user %s failed", userID)
//
// Use [Wrap] and [Wrapf] to wrap an existing error while attaching a code and outer message. The wrapped error is
// available via [Error.Unwrap] and participates in [errors.Is] / [errors.As] chains:
//
//	if err != nil {
//		return xerrors.Wrap(err, 502, "downstream call failed")
//	}
//	return xerrors.Wrapf(err, 503, "retry %d exhausted", n)
//
// # Caller information
//
// By default, [Error.Caller] returns placeholder values until caller capture runs. There are two ways to record the
// call site: call [Error.WithCaller] on the error, or attach options to a [context.Context] and call
// [Error.WithOptionsFromContext].
//
// Call [Error.WithStripFilePrefixes] (and optionally [Error.WithSkipBias]) on the error before [Error.WithCaller] so
// recorded paths omit noisy prefixes (for example the module root or workspace directory) and the stack skip matches
// your wrappers. Then call [Error.WithCaller] once; further calls do not replace the stored caller. After capture,
// [Error.Caller] returns file, line, and function name for the [Error.WithCaller] invocation:
//
//	wd, _ := os.Getwd()
//	err := xerrors.New(1, "oops").
//		WithStripFilePrefixes(wd + string(os.PathSeparator)).
//		WithSkipBias(0).
//		WithCaller()
//	info := err.Caller()
//
// # Callers via context
//
// Build a context with [ContextWithErrorOptions], passing any combination of [WithCaptureCaller], [WithSkipBias],
// and [WithStripFilePrefixes]. When an error is configured with [Error.WithOptionsFromContext], those options are
// applied to the embedded [ErrorOptions], and if capture is enabled the caller is stored immediately (only the first
// successful capture is kept).
//
// Typical pattern at a request or handler boundary:
//
//	ctx := xerrors.ContextWithErrorOptions(ctx,
//		xerrors.WithCaptureCaller(),
//		xerrors.WithStripFilePrefixes(moduleRootPrefix),
//	)
//	// ... pass ctx down ...
//	err := xerrors.New(code, "failed").WithOptionsFromContext(ctx)
//
// Apply other [Error] configuration (such as [Error.WithAttr]) before [Error.WithOptionsFromContext] or
// [Error.WithCaller], and before reading caller details with [Error.Caller], so attributes and options are settled
// first.
//
// A nil context is treated like [context.Background] when attaching options; [ErrorOptionsFromContext] returns nil
// when the context carries no options.
//
// # Third-party libraries and context
//
// If your library returns [XError] (or [*Error]) from its API, accept a [context.Context] on those entry points and
// apply it when you construct the error so applications can inject options without importing your internals. After you
// build the error with [New], [Newf], [Wrap], or [Wrapf], call [Error.WithOptionsFromContext] with the same context you
// received (and configure attributes or other [Error] methods before that call, as described above):
//
//	func (c *Client) DoSomething(ctx context.Context, id string) xerrors.XError {
//		if id == "" {
//			return xerrors.New(400, "empty id").WithOptionsFromContext(ctx)
//		}
//		// ...
//	}
//
// Callers in an application combine their request context with [ContextWithErrorOptions] once, then pass the
// resulting context into library functions. Options such as [WithCaptureCaller] and [WithStripFilePrefixes] are then
// applied when the library calls [Error.WithOptionsFromContext]:
//
//	ctx := r.Context() // or another root context
//	ctx = xerrors.ContextWithErrorOptions(ctx,
//		xerrors.WithCaptureCaller(),
//		xerrors.WithStripFilePrefixes(appStripPrefix),
//	)
//	err := thirdpartyclient.DoSomething(ctx, id)
//
// The library is responsible for threading that context into every error it returns; if it omits
// [Error.WithOptionsFromContext], application-supplied options have no effect on those errors.
//
// # Custom errors, codes, and the XError interface
//
// Treat integer codes as part of your public contract: define named constants (or enums) in your own package, reserve
// ranges or namespaces for services or domains, and use those constants with [New], [Newf], [Wrap], and [Wrapf]. Fixed
// user-facing strings can live in the same place as constants, or you can build messages with [Newf] / [Wrapf] while
// still using a stable code per failure kind:
//
//	const (
//		CodeNotFound   = 404001
//		CodeValidation = 400001
//	)
//
//	func errUserMissing(id string) xerrors.XError {
//		return xerrors.Newf(CodeValidation, "user %q is required", id)
//	}
//
// The concrete type [*Error] already implements [XError]. To add your own methods or types while keeping JSON and
// caller behavior, embed a pointer to [*Error] and construct it only through this package (so codes and messages stay
// consistent). Methods on [*Error] promote to the outer type; returning [XError] from factories keeps callers decoupled
// from the concrete struct:
//
//	type NotFound struct{ *xerrors.Error }
//
//	func NewNotFound(resource string) xerrors.XError {
//		return &NotFound{Error: xerrors.New(CodeNotFound, "not found").(*xerrors.Error)}
//	}
//
// Implementing [XError] without embedding is possible but requires every method from the interface (including JSON and
// the With* helpers); embedding [*Error] is the supported approach.
//
// # JSON and wrapped errors
//
// The wrappedError field is a tagged union. Use kind "xerror" with a nested object for another [*Error], or kind
// "std" with message for opaque errors. JSON without "kind" in wrappedError is not accepted.
//
//	"wrappedError": {"kind":"xerror","nested":{"code":2,"message":"inner",...}}
//	"wrappedError": {"kind":"std","message":"upstream failed"}
//
// # Behavior notes
//
// [Error.Attrs] returns a shallow copy of the attribute map; mutating the returned map does not change the [Error].
// Values that are pointers or slices are still shared with the stored attribute value.
//
// [ContextWithErrorOptions] appends new [ErrorOptionFn] values after any options already stored on the context, so
// layered middleware can add options without dropping earlier ones. Option functions still apply in order and can
// overwrite the same [ErrorOptions] fields when run later.
package xerrors

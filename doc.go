// Package xerrors provides structured errors with numeric codes, optional attributes, JSON serialization, and
// optional caller (file, line, function) recording at the point the error is built.
//
// # Basic construction
//
// Use [New] for a plain message and [Newf] when the message is formatted. Both return the [Error] interface; the
// concrete type is [*XError]:
//
//	err := xerrors.New(404, "not found")
//	err = xerrors.Newf(500, "user %s failed", userID)
//
// Use [Wrap] and [Wrapf] to wrap an existing error while attaching a code and outer message. The wrapped error is
// available via [XError.Unwrap] and participates in [errors.Is] / [errors.As] chains:
//
//	if err != nil {
//		return xerrors.Wrap(err, 502, "downstream call failed")
//	}
//	return xerrors.Wrapf(err, 503, "retry %d exhausted", n)
//
// # Caller information
//
// By default, [XError.Caller] returns placeholder values until caller capture runs. There are two ways to record the
// call site: call [XError.WithCaller] on the error, or attach options to a [context.Context] and call
// [XError.WithOptionsFromContext].
//
// Call [XError.WithStripFilePrefixes] (and optionally [XError.WithSkipBias]) on the error before [XError.WithCaller]
// so recorded paths omit noisy prefixes (for example the module root or workspace directory) and the stack skip
// matches your wrappers. Then call [XError.WithCaller] once; further calls do not replace the stored caller. After
// capture, [XError.Caller] returns file, line, and function name for the [XError.WithCaller] invocation:
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
// and [WithStripFilePrefixes]. When an error is configured with [XError.WithOptionsFromContext], those options are
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
// Apply other [XError] configuration (such as [XError.WithAttr]) before [XError.WithOptionsFromContext] or
// [XError.WithCaller], and before reading caller details with [XError.Caller], so attributes and options are settled
// first.
//
// A nil context is treated like [context.Background] when attaching options; [ErrorOptionsFromContext] returns nil
// when the context carries no options.
//
// # Third-party libraries and context
//
// If your library returns [Error] (typically a [*XError] or a wrapper that embeds it) from its API, accept a
// [context.Context] on those entry points and apply it when you construct the error so applications can inject options
// without importing your internals. After you build the error with [New], [Newf], [Wrap], or [Wrapf], call
// [XError.WithOptionsFromContext] with the same context you received (and configure attributes or other [XError]
// methods before that call, as described above):
//
//	func (c *Client) DoSomething(ctx context.Context, id string) xerrors.Error {
//		if id == "" {
//			return xerrors.New(400, "empty id").WithOptionsFromContext(ctx)
//		}
//		// ...
//	}
//
// Callers in an application combine their request context with [ContextWithErrorOptions] once, then pass the
// resulting context into library functions. Options such as [WithCaptureCaller] and [WithStripFilePrefixes] are then
// applied when the library calls [XError.WithOptionsFromContext]:
//
//	ctx := r.Context() // or another root context
//	ctx = xerrors.ContextWithErrorOptions(ctx,
//		xerrors.WithCaptureCaller(),
//		xerrors.WithStripFilePrefixes(appStripPrefix),
//	)
//	err := thirdpartyclient.DoSomething(ctx, id)
//
// The library is responsible for threading that context into every error it returns; if it omits
// [XError.WithOptionsFromContext], application-supplied options have no effect on those errors.
//
// # NewXError for third-party and application wrapper types
//
// [NewXError] returns a concrete [*XError] in one step: it applies [XError.WithOptionsFromContext] using the given
// context, chooses [Newf]-style construction when the wrapped error is nil, and [Wrapf]-style construction when it is
// not, using the same message and variadic argument rules as those functions. That makes it a convenient building block
// when your public API returns a struct that embeds [*XError]: you avoid repeating
// `Newf(...).WithOptionsFromContext(ctx).(*xerrors.XError)` or the wrap equivalent at every return site.
//
// Define your type with an embedded [*XError] (see the section below), then wire the pointer from [NewXError]:
//
//	type ClientError struct{ *xerrors.XError }
//
//	func errUpstream(ctx context.Context, cause error) *ClientError {
//		return &ClientError{
//			XError: xerrors.NewXError(ctx, cause, CodeUpstream, "downstream request failed"),
//		}
//	}
//
//	func errValidation(ctx context.Context, field string) *ClientError {
//		return &ClientError{
//			XError: xerrors.NewXError(ctx, nil, CodeValidation, "field %s is required", field),
//		}
//	}
//
// Pass the same [context.Context] you received on the API boundary so caller capture and strip-prefix options from
// [ContextWithErrorOptions] apply automatically. If you need a custom concrete type other than [*XError], build with
// [NewXError] and pass the result to [XErrorAs], or continue using [NewAs] / [NewfAs] / [WrapAs] / [WrapfAs], which
// accept a constructor callback instead of returning [*XError] directly.
//
// # Generic constructors (NewAs, NewfAs, WrapAs, WrapfAs, XErrorAs)
//
// [NewAs], [NewfAs], [WrapAs], and [WrapfAs] take a constructor function as their first argument. The function
// receives the built [*XError] and returns your own type T constrained by [Error], so you avoid a type assertion at
// the call site and keep a concrete return type for APIs and tests:
//
//	func NewNotFound() *NotFound {
//		return xerrors.NewAs(func(e *xerrors.XError) *NotFound {
//			return &NotFound{XError: e}
//		}, CodeNotFound, "not found")
//	}
//
//	func NewNotFoundf(userID string) *NotFound {
//		return xerrors.NewfAs(func(e *xerrors.XError) *NotFound {
//			return &NotFound{XError: e}
//		}, CodeNotFound, "user %s not found", userID)
//	}
//
//	func WrapNotFound(cause error) *NotFound {
//		return xerrors.WrapAs(func(e *xerrors.XError) *NotFound {
//			return &NotFound{XError: e}
//		}, cause, CodeNotFound, "resource missing")
//	}
//
// [WrapfAs] passes trailing arguments to the same formatting rules as [Wrapf] (empty args means a literal message
// string):
//
//	func WrapRetry(cause error, n int) *NotFound {
//		return xerrors.WrapfAs(func(e *xerrors.XError) *NotFound {
//			return &NotFound{XError: e}
//		}, cause, CodeNotFound, "retry %d exhausted", n)
//	}
//
// If you already hold a [*XError] (for example after options or attributes) and only need to change the concrete
// type, use [XErrorAs]:
//
//	base := xerrors.New(CodeNotFound, "gone").(*xerrors.XError)
//	out := xerrors.XErrorAs(base, func(e *xerrors.XError) *NotFound { return &NotFound{XError: e} })
//
// Third-party modules can use the same pattern: define stable codes in the application, embed [*XError], and return
// [Error] or a pointer to the wrapper from public APIs.
//
// Implementing [Error] without embedding is possible but requires every method from the interface (including JSON and
// the With* helpers); embedding [*XError] is the supported approach.
//
// # JSON and wrapped errors
//
// The wrappedError field is a tagged union. Use kind "xerrors.Error" with a nested object for another [*XError], or
// kind "std" with message for opaque errors. JSON without "kind" in wrappedError is not accepted.
//
//	"wrappedError": {"kind":"xerrors.Error","nested":{"code":2,"message":"inner",...}}
//	"wrappedError": {"kind":"std","message":"upstream failed"}
//
// # Behavior notes
//
// [XError.Attrs] returns a shallow copy of the attribute map; mutating the returned map does not change the [*XError].
// Values that are pointers or slices are still shared with the stored attribute value.
//
// [ContextWithErrorOptions] appends new [ErrorOptionFn] values after any options already stored on the context, so
// layered middleware can add options without dropping earlier ones. Option functions still apply in order and can
// overwrite the same [ErrorOptions] fields when run later.
package xerrors

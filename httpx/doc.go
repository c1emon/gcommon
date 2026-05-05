// Package httpx provides a [ClientFactory]-centric wrapper around github.com/imroc/req/v3:
// multiple named profiles, shared defaults, optional per-client overrides, built-in
// JSON envelope error handling via interceptors.Error. Shared response/result and
// pagination value objects are provided by package github.com/c1emon/gcommon/vo.
// It also supports optional retry and rate limiting (golang.org/x/time/rate), and
// browser impersonation profiles. Use [InitDefaultClientFactory] and [GetDefaultClientFactory] for
// a package-level default [ClientFactory] when a single shared factory is enough.
//
// Rate limiting is registered as the first OnBeforeRequest hook so it always runs
// before user-defined request middleware. Response hooks run in registration order
// after req's internal parsers; the error interceptor is always registered last on the client.
package httpx

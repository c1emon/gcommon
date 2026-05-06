// Package httpx provides a [ClientFactory]-centric wrapper around github.com/imroc/req/v3:
// named profiles, shared defaults, and fresh caller-owned clients created from those
// profiles. Shared response/result and pagination value objects are provided by
// package github.com/c1emon/gcommon/v2/vo. It also supports optional retry, rate
// limiting (golang.org/x/time/rate), browser impersonation profiles, browser
// navigation headers, cookie jars, redirect policies, and opt-in JSON envelope
// error handling via interceptors.Error.
// Use [InitDefaultClientFactory] and [GetDefaultClientFactory] for a package-level
// default [ClientFactory] when a single shared factory is enough.
//
// Rate limiting is registered as the first OnBeforeRequest hook so it always runs
// before user-defined request middleware. Response hooks run in registration order
// after req's internal parsers. The business-error interceptor is registered only
// when enabled with [WithGlobalBusinessError] or [WithBusinessError].
//
// Browser navigation captures request-explicit headers before req merges common
// headers, then applies dynamic navigation headers in the client round-trip
// wrapper after URL/body parsing.
package httpx

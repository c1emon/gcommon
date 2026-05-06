package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/c1emon/gcommon/httpx/v2"
)

type seenNavigationHeaders struct {
	Accept                  string `json:"accept"`
	Referer                 string `json:"referer"`
	Origin                  string `json:"origin"`
	SecFetchSite            string `json:"secFetchSite"`
	SecFetchMode            string `json:"secFetchMode"`
	SecFetchDest            string `json:"secFetchDest"`
	SecFetchUser            string `json:"secFetchUser"`
	UpgradeInsecureRequests string `json:"upgradeInsecureRequests"`
	XRequestedWith          string `json:"xRequestedWith"`
	InternalRequestKind     string `json:"internalRequestKind"`
}

type navigationRecorder struct {
	server *httptest.Server
	seen   []seenNavigationHeaders
}

func newNavigationRecorder(t *testing.T) *navigationRecorder {
	t.Helper()

	rec := &navigationRecorder{}
	rec.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := seenNavigationHeaders{
			Accept:                  r.Header.Get("Accept"),
			Referer:                 r.Header.Get("Referer"),
			Origin:                  r.Header.Get("Origin"),
			SecFetchSite:            r.Header.Get("Sec-Fetch-Site"),
			SecFetchMode:            r.Header.Get("Sec-Fetch-Mode"),
			SecFetchDest:            r.Header.Get("Sec-Fetch-Dest"),
			SecFetchUser:            r.Header.Get("Sec-Fetch-User"),
			UpgradeInsecureRequests: r.Header.Get("Upgrade-Insecure-Requests"),
			XRequestedWith:          r.Header.Get("X-Requested-With"),
			InternalRequestKind:     r.Header.Get("X-Httpx-Browser-Request-Kind"),
		}
		rec.seen = append(rec.seen, h)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(rec.server.Close)
	return rec
}

func TestBrowserNavigation_firstNavigationHeaders(t *testing.T) {
	rec := newNavigationRecorder(t)
	f := httpx.NewClientFactory()
	f.RegisterProfile("browser",
		httpx.WithBaseURL(rec.server.URL),
		httpx.WithBrowserNavigation(),
	)
	c := f.MustNewClient("browser")

	if _, err := c.Req().Get("/login?next=%2Fhome"); err != nil {
		t.Fatal(err)
	}

	if len(rec.seen) != 1 {
		t.Fatalf("requests seen: want 1, got %d", len(rec.seen))
	}
	first := rec.seen[0]
	if first.SecFetchSite != "none" {
		t.Fatalf("Sec-Fetch-Site: want none, got %q", first.SecFetchSite)
	}
	if first.Referer != "" {
		t.Fatalf("Referer: want empty, got %q", first.Referer)
	}
	if !strings.Contains(first.Accept, "text/html") || !strings.Contains(first.Accept, "application/xhtml+xml") {
		t.Fatalf("Accept: want HTML navigation values, got %q", first.Accept)
	}
	if first.SecFetchMode != "navigate" {
		t.Fatalf("Sec-Fetch-Mode: want navigate, got %q", first.SecFetchMode)
	}
	if first.SecFetchDest != "document" {
		t.Fatalf("Sec-Fetch-Dest: want document, got %q", first.SecFetchDest)
	}
	if first.SecFetchUser != "?1" {
		t.Fatalf("Sec-Fetch-User: want ?1, got %q", first.SecFetchUser)
	}
	if first.UpgradeInsecureRequests != "1" {
		t.Fatalf("Upgrade-Insecure-Requests: want 1, got %q", first.UpgradeInsecureRequests)
	}
	if first.XRequestedWith != "" {
		t.Fatalf("X-Requested-With: want empty for navigation, got %q", first.XRequestedWith)
	}
}

func TestBrowserNavigation_sameOriginSecondRequestUsesPreviousFullURL(t *testing.T) {
	rec := newNavigationRecorder(t)
	f := httpx.NewClientFactory()
	f.RegisterProfile("browser",
		httpx.WithBaseURL(rec.server.URL),
		httpx.WithBrowserNavigation(),
	)
	c := f.MustNewClient("browser")

	if _, err := c.Req().Get("/login?next=%2Fhome"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Req().Get("/home"); err != nil {
		t.Fatal(err)
	}

	if len(rec.seen) != 2 {
		t.Fatalf("requests seen: want 2, got %d", len(rec.seen))
	}
	second := rec.seen[1]
	if second.SecFetchSite != "same-origin" {
		t.Fatalf("Sec-Fetch-Site: want same-origin, got %q", second.SecFetchSite)
	}
	wantReferer := rec.server.URL + "/login?next=%2Fhome"
	if second.Referer != wantReferer {
		t.Fatalf("Referer: want %q, got %q", wantReferer, second.Referer)
	}
}

func TestBrowserNavigation_crossOriginRefererUsesPreviousOrigin(t *testing.T) {
	first := newNavigationRecorder(t)
	second := newNavigationRecorder(t)

	f := httpx.NewClientFactory()
	f.RegisterProfile("browser", httpx.WithBrowserNavigation())
	c := f.MustNewClient("browser")

	if _, err := c.Req().Get(first.server.URL + "/start?ticket=abc"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Req().Get(second.server.URL + "/next"); err != nil {
		t.Fatal(err)
	}

	if len(second.seen) != 1 {
		t.Fatalf("cross-origin requests seen: want 1, got %d", len(second.seen))
	}
	got := second.seen[0]
	if got.SecFetchSite != "cross-site" {
		t.Fatalf("Sec-Fetch-Site: want cross-site, got %q", got.SecFetchSite)
	}
	wantReferer := first.server.URL + "/"
	if got.Referer != wantReferer {
		t.Fatalf("Referer: want %q, got %q", wantReferer, got.Referer)
	}
}

func TestBrowserNavigation_postAddsCurrentOrigin(t *testing.T) {
	rec := newNavigationRecorder(t)
	f := httpx.NewClientFactory()
	f.RegisterProfile("browser",
		httpx.WithBaseURL(rec.server.URL),
		httpx.WithBrowserNavigation(),
	)
	c := f.MustNewClient("browser")

	if _, err := c.Req().Post("/submit"); err != nil {
		t.Fatal(err)
	}

	if len(rec.seen) != 1 {
		t.Fatalf("requests seen: want 1, got %d", len(rec.seen))
	}
	if rec.seen[0].Origin != rec.server.URL {
		t.Fatalf("Origin: want %q, got %q", rec.server.URL, rec.seen[0].Origin)
	}
}

func TestBrowserNavigation_jsonRequestDefaultsToXHR(t *testing.T) {
	rec := newNavigationRecorder(t)
	f := httpx.NewClientFactory()
	f.RegisterProfile("browser",
		httpx.WithBaseURL(rec.server.URL),
		httpx.WithBrowserNavigation(),
	)
	c := f.MustNewClient("browser")

	if _, err := c.Req().SetBodyJsonMarshal(map[string]string{"hello": "world"}).Post("/api"); err != nil {
		t.Fatal(err)
	}

	if len(rec.seen) != 1 {
		t.Fatalf("requests seen: want 1, got %d", len(rec.seen))
	}
	got := rec.seen[0]
	if got.Accept != "application/json, text/plain, */*" {
		t.Fatalf("Accept: want JSON XHR value, got %q", got.Accept)
	}
	if got.SecFetchMode != "cors" {
		t.Fatalf("Sec-Fetch-Mode: want cors, got %q", got.SecFetchMode)
	}
	if got.SecFetchDest != "empty" {
		t.Fatalf("Sec-Fetch-Dest: want empty, got %q", got.SecFetchDest)
	}
	if got.SecFetchUser != "" {
		t.Fatalf("Sec-Fetch-User: want empty for XHR, got %q", got.SecFetchUser)
	}
	if got.UpgradeInsecureRequests != "" {
		t.Fatalf("Upgrade-Insecure-Requests: want empty for XHR, got %q", got.UpgradeInsecureRequests)
	}
}

func TestBrowserNavigation_asNavigationOverridesJSONRequestKind(t *testing.T) {
	rec := newNavigationRecorder(t)
	f := httpx.NewClientFactory()
	f.RegisterProfile("browser",
		httpx.WithBaseURL(rec.server.URL),
		httpx.WithBrowserNavigation(),
	)
	c := f.MustNewClient("browser")

	if _, err := c.Req().AsNavigation().SetBodyJsonMarshal(map[string]string{"hello": "world"}).Post("/form"); err != nil {
		t.Fatal(err)
	}

	if len(rec.seen) != 1 {
		t.Fatalf("requests seen: want 1, got %d", len(rec.seen))
	}
	got := rec.seen[0]
	if got.SecFetchMode != "navigate" {
		t.Fatalf("Sec-Fetch-Mode: want navigate, got %q", got.SecFetchMode)
	}
	if got.SecFetchDest != "document" {
		t.Fatalf("Sec-Fetch-Dest: want document, got %q", got.SecFetchDest)
	}
}

func TestBrowserNavigation_doesNotOverrideSingleRequestExplicitHeaders(t *testing.T) {
	rec := newNavigationRecorder(t)
	f := httpx.NewClientFactory()
	f.RegisterProfile("browser",
		httpx.WithBaseURL(rec.server.URL),
		httpx.WithHeader("Sec-Fetch-Site", "profile-default"),
		httpx.WithBrowserNavigation(),
	)
	c := f.MustNewClient("browser")

	if _, err := c.Req().
		SetHeader("Sec-Fetch-Site", "manual").
		SetHeader("Referer", "https://manual.example/path").
		SetHeader("Accept", "text/custom").
		Get("/manual"); err != nil {
		t.Fatal(err)
	}

	if len(rec.seen) != 1 {
		t.Fatalf("requests seen: want 1, got %d", len(rec.seen))
	}
	got := rec.seen[0]
	if got.SecFetchSite != "manual" {
		t.Fatalf("Sec-Fetch-Site: want manual, got %q", got.SecFetchSite)
	}
	if got.Referer != "https://manual.example/path" {
		t.Fatalf("Referer: want explicit value, got %q", got.Referer)
	}
	if got.Accept != "text/custom" {
		t.Fatalf("Accept: want explicit value, got %q", got.Accept)
	}
}

func TestBrowserNavigation_removesProfileDynamicHeadersWhenNavigationOmitsThem(t *testing.T) {
	rec := newNavigationRecorder(t)
	f := httpx.NewClientFactory()
	f.RegisterProfile("browser",
		httpx.WithBaseURL(rec.server.URL),
		httpx.WithHeader("Referer", "https://profile.example/previous"),
		httpx.WithHeader("Origin", "https://profile.example"),
		httpx.WithBrowserNavigation(),
	)
	c := f.MustNewClient("browser")

	if _, err := c.Req().Get("/first"); err != nil {
		t.Fatal(err)
	}

	if len(rec.seen) != 1 {
		t.Fatalf("requests seen: want 1, got %d", len(rec.seen))
	}
	got := rec.seen[0]
	if got.Referer != "" {
		t.Fatalf("Referer: want empty on first navigation, got %q", got.Referer)
	}
	if got.Origin != "" {
		t.Fatalf("Origin: want empty for GET, got %q", got.Origin)
	}
}

func TestBrowserNavigation_noReferrerPolicyRemovesProfileReferer(t *testing.T) {
	rec := newNavigationRecorder(t)
	f := httpx.NewClientFactory()
	f.RegisterProfile("browser",
		httpx.WithBaseURL(rec.server.URL),
		httpx.WithHeader("Referer", "https://profile.example/previous"),
		httpx.WithBrowserNavigation(
			httpx.WithReferrerPolicy(httpx.ReferrerPolicyNoReferrer),
		),
	)
	c := f.MustNewClient("browser")

	if _, err := c.Req().Get("/first"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Req().Get("/second"); err != nil {
		t.Fatal(err)
	}

	if len(rec.seen) != 2 {
		t.Fatalf("requests seen: want 2, got %d", len(rec.seen))
	}
	for i, got := range rec.seen {
		if got.Referer != "" {
			t.Fatalf("request %d Referer: want empty, got %q", i+1, got.Referer)
		}
	}
}

func TestBrowserNavigation_overridesBrowserProfileSecFetchSite(t *testing.T) {
	rec := newNavigationRecorder(t)
	f := httpx.NewClientFactory()
	f.RegisterProfile("browser",
		httpx.WithBaseURL(rec.server.URL),
		httpx.WithBrowser(httpx.BrowserChrome),
		httpx.WithBrowserNavigation(),
	)
	c := f.MustNewClient("browser")

	if _, err := c.Req().Get("/first"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Req().Get("/second"); err != nil {
		t.Fatal(err)
	}

	if len(rec.seen) != 2 {
		t.Fatalf("requests seen: want 2, got %d", len(rec.seen))
	}
	if rec.seen[1].SecFetchSite != "same-origin" {
		t.Fatalf("Sec-Fetch-Site: want same-origin, got %q", rec.seen[1].SecFetchSite)
	}
}

func TestBrowserNavigation_newClientStateIsolation(t *testing.T) {
	rec := newNavigationRecorder(t)
	f := httpx.NewClientFactory()
	f.RegisterProfile("browser",
		httpx.WithBaseURL(rec.server.URL),
		httpx.WithBrowserNavigation(),
	)

	first := f.MustNewClient("browser")
	second := f.MustNewClient("browser")
	if _, err := first.Req().Get("/first"); err != nil {
		t.Fatal(err)
	}
	if _, err := second.Req().Get("/second"); err != nil {
		t.Fatal(err)
	}

	if len(rec.seen) != 2 {
		t.Fatalf("requests seen: want 2, got %d", len(rec.seen))
	}
	got := rec.seen[1]
	if got.Referer != "" {
		t.Fatalf("second client Referer: want empty, got %q", got.Referer)
	}
	if got.SecFetchSite != "none" {
		t.Fatalf("second client Sec-Fetch-Site: want none, got %q", got.SecFetchSite)
	}
}

func TestBrowserNavigation_cloneSharesState(t *testing.T) {
	rec := newNavigationRecorder(t)
	f := httpx.NewClientFactory()
	f.RegisterProfile("browser",
		httpx.WithBaseURL(rec.server.URL),
		httpx.WithBrowserNavigation(),
	)
	original := f.MustNewClient("browser")

	if _, err := original.Req().Get("/first"); err != nil {
		t.Fatal(err)
	}
	clone := original.Clone()
	if _, err := clone.Req().Get("/second"); err != nil {
		t.Fatal(err)
	}
	if _, err := original.Req().Get("/third"); err != nil {
		t.Fatal(err)
	}

	if len(rec.seen) != 3 {
		t.Fatalf("requests seen: want 3, got %d", len(rec.seen))
	}
	wantCloneReferer := rec.server.URL + "/first"
	if rec.seen[1].Referer != wantCloneReferer {
		t.Fatalf("clone Referer: want %q, got %q", wantCloneReferer, rec.seen[1].Referer)
	}
	wantOriginalReferer := rec.server.URL + "/second"
	if rec.seen[2].Referer != wantOriginalReferer {
		t.Fatalf("original Referer after clone request: want %q, got %q", wantOriginalReferer, rec.seen[2].Referer)
	}
}

func TestBrowserNavigation_remembersFinalURLAfterRedirect(t *testing.T) {
	var seen []seenNavigationHeaders
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			http.Redirect(w, r, "/landing", http.StatusFound)
			return
		}
		seen = append(seen, seenNavigationHeaders{
			Referer:      r.Header.Get("Referer"),
			SecFetchSite: r.Header.Get("Sec-Fetch-Site"),
		})
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	f := httpx.NewClientFactory()
	f.RegisterProfile("browser",
		httpx.WithBaseURL(srv.URL),
		httpx.WithBrowserNavigation(),
	)
	c := f.MustNewClient("browser")

	if _, err := c.Req().Get("/start"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Req().Get("/next"); err != nil {
		t.Fatal(err)
	}

	if len(seen) != 2 {
		t.Fatalf("non-redirect requests seen: want 2, got %d", len(seen))
	}
	wantReferer := srv.URL + "/landing"
	if seen[1].Referer != wantReferer {
		t.Fatalf("Referer: want %q, got %q", wantReferer, seen[1].Referer)
	}
}

func TestBrowserNavigation_internalRequestKindMarkerDoesNotLeakWithNonCanonicalHeader(t *testing.T) {
	rec := newNavigationRecorder(t)
	f := httpx.NewClientFactory()
	f.RegisterProfile("browser",
		httpx.WithBaseURL(rec.server.URL),
		httpx.WithBrowserNavigation(),
	)
	c := f.MustNewClient("browser")

	if _, err := c.Req().
		SetHeaderNonCanonical("x-httpx-browser-request-kind", "xhr").
		Get("/api"); err != nil {
		t.Fatal(err)
	}

	if len(rec.seen) != 1 {
		t.Fatalf("requests seen: want 1, got %d", len(rec.seen))
	}
	got := rec.seen[0]
	if got.InternalRequestKind != "" {
		t.Fatalf("internal request kind header leaked: got %q", got.InternalRequestKind)
	}
	if got.SecFetchMode != "cors" {
		t.Fatalf("Sec-Fetch-Mode: want cors, got %q", got.SecFetchMode)
	}
	if got.SecFetchDest != "empty" {
		t.Fatalf("Sec-Fetch-Dest: want empty, got %q", got.SecFetchDest)
	}
	if got.SecFetchUser != "" {
		t.Fatalf("Sec-Fetch-User: want empty for XHR, got %q", got.SecFetchUser)
	}
	if got.UpgradeInsecureRequests != "" {
		t.Fatalf("Upgrade-Insecure-Requests: want empty for XHR, got %q", got.UpgradeInsecureRequests)
	}
}

func TestBrowserNavigation_retryPreservesPreviousURLAndRequestKind(t *testing.T) {
	var retrySeen []seenNavigationHeaders
	var otherSeen []seenNavigationHeaders
	retryAttempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := seenNavigationHeaders{
			Accept:                  r.Header.Get("Accept"),
			Referer:                 r.Header.Get("Referer"),
			Origin:                  r.Header.Get("Origin"),
			SecFetchSite:            r.Header.Get("Sec-Fetch-Site"),
			SecFetchMode:            r.Header.Get("Sec-Fetch-Mode"),
			SecFetchDest:            r.Header.Get("Sec-Fetch-Dest"),
			SecFetchUser:            r.Header.Get("Sec-Fetch-User"),
			UpgradeInsecureRequests: r.Header.Get("Upgrade-Insecure-Requests"),
			XRequestedWith:          r.Header.Get("X-Requested-With"),
			InternalRequestKind:     r.Header.Get("X-Httpx-Browser-Request-Kind"),
		}
		if r.URL.Path == "/retry" {
			retrySeen = append(retrySeen, h)
			retryAttempts++
			if retryAttempts == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		otherSeen = append(otherSeen, h)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	f := httpx.NewClientFactory()
	f.RegisterProfile("browser",
		httpx.WithBaseURL(srv.URL),
		httpx.WithBrowserNavigation(),
		httpx.WithRetry(httpx.RetryPolicy{
			Enabled:    true,
			MaxRetries: 1,
			MinBackoff: time.Millisecond,
			MaxBackoff: time.Millisecond,
		}),
	)
	c := f.MustNewClient("browser")

	if _, err := c.Req().Get("/before"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Req().AsXHR().Post("/retry"); err != nil {
		t.Fatal(err)
	}

	if len(retrySeen) != 2 {
		t.Fatalf("retry attempts seen: want 2, got %d", len(retrySeen))
	}
	if len(otherSeen) != 1 {
		t.Fatalf("non-retry requests seen: want 1, got %d", len(otherSeen))
	}
	wantReferer := srv.URL + "/before"
	for i, got := range retrySeen {
		if got.SecFetchMode != "cors" {
			t.Fatalf("attempt %d Sec-Fetch-Mode: want cors, got %q", i+1, got.SecFetchMode)
		}
		if got.SecFetchDest != "empty" {
			t.Fatalf("attempt %d Sec-Fetch-Dest: want empty, got %q", i+1, got.SecFetchDest)
		}
		if got.SecFetchUser != "" {
			t.Fatalf("attempt %d Sec-Fetch-User: want empty for XHR, got %q", i+1, got.SecFetchUser)
		}
		if got.UpgradeInsecureRequests != "" {
			t.Fatalf("attempt %d Upgrade-Insecure-Requests: want empty for XHR, got %q", i+1, got.UpgradeInsecureRequests)
		}
		if got.Referer != wantReferer {
			t.Fatalf("attempt %d Referer: want %q, got %q", i+1, wantReferer, got.Referer)
		}
	}
}

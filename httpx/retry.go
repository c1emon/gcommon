package httpx

import (
	"time"

	"github.com/imroc/req/v3"
)

// RetryPolicy configures client-level automatic retries (disabled by default).
type RetryPolicy struct {
	Enabled bool

	// MaxRetries is passed to req's SetCommonRetryCount (same semantics as imroc/req).
	MaxRetries int

	// MinBackoff and MaxBackoff enable capped exponential backoff between attempts.
	// Zero MinBackoff skips SetCommonRetryBackoffInterval (req default interval applies).
	MinBackoff time.Duration
	MaxBackoff time.Duration

	// RetryIf, when non-nil, is OR-combined with the default transport / HTTP error checks.
	RetryIf func(resp *req.Response, err error) bool
}

func defaultRetryCondition(resp *req.Response, err error) bool {
	if err != nil {
		return true
	}
	if resp == nil {
		return false
	}
	if resp.StatusCode == 429 {
		return true
	}
	return resp.StatusCode >= 500 && resp.StatusCode < 600
}

func applyRetryPolicy(c *req.Client, p RetryPolicy) {
	if !p.Enabled || p.MaxRetries == 0 {
		return
	}
	c.SetCommonRetryCount(p.MaxRetries)
	if p.MinBackoff > 0 && p.MaxBackoff > 0 {
		c.SetCommonRetryBackoffInterval(p.MinBackoff, p.MaxBackoff)
	}
	if p.RetryIf != nil {
		c.SetCommonRetryCondition(func(resp *req.Response, err error) bool {
			if defaultRetryCondition(resp, err) {
				return true
			}
			return p.RetryIf(resp, err)
		})
	} else {
		c.SetCommonRetryCondition(defaultRetryCondition)
	}
}

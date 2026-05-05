package httpx

import (
	"github.com/c1emon/gcommon/util"
	"github.com/imroc/req/v3"
)

type RedirectPolicy = req.RedirectPolicy

func WithRedirectPolicy(policies ...RedirectPolicy) ClientOption {
	return util.WrapFuncOption(func(o *clientRegisterOpts) {
		o.redirectPolicySet = true
		o.redirectPolicies = append([]RedirectPolicy(nil), policies...)
	})
}

func NoRedirectPolicy() RedirectPolicy                  { return req.NoRedirectPolicy() }
func DefaultRedirectPolicy() RedirectPolicy             { return req.DefaultRedirectPolicy() }
func MaxRedirectPolicy(noOfRedirect int) RedirectPolicy { return req.MaxRedirectPolicy(noOfRedirect) }
func SameDomainRedirectPolicy() RedirectPolicy          { return req.SameDomainRedirectPolicy() }
func SameHostRedirectPolicy() RedirectPolicy            { return req.SameHostRedirectPolicy() }
func AllowedHostRedirectPolicy(hosts ...string) RedirectPolicy {
	return req.AllowedHostRedirectPolicy(hosts...)
}
func AllowedDomainRedirectPolicy(hosts ...string) RedirectPolicy {
	return req.AllowedDomainRedirectPolicy(hosts...)
}
func AlwaysCopyHeaderRedirectPolicy(headers ...string) RedirectPolicy {
	return req.AlwaysCopyHeaderRedirectPolicy(headers...)
}

func (c *Client) SetRedirectPolicy(policies ...RedirectPolicy) *Client {
	if c == nil {
		return nil
	}
	if c.Client != nil {
		c.Client.SetRedirectPolicy(policies...)
	}
	return c
}

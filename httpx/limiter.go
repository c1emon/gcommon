package httpx

import "github.com/imroc/req/v3"

// applyLimiterMiddleware wires global/client limiters as the first request hook.
// imroc/req runs OnBeforeRequest middlewares in registration order.
func (f *ClientFactory) applyLimiterMiddleware(c *req.Client, o *clientRegisterOpts) {
	if f.globalLimiter == nil && (o.noClientLimiter || o.clientLimiter == nil) {
		return
	}

	gl, cl, skipClient := f.globalLimiter, o.clientLimiter, o.noClientLimiter
	c.OnBeforeRequest(func(_ *req.Client, r *req.Request) error {
		ctx := r.Context()
		if gl != nil {
			if err := gl.Wait(ctx); err != nil {
				return err
			}
		}
		if !skipClient && cl != nil {
			if err := cl.Wait(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

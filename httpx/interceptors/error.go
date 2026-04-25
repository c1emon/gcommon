package interceptors

import (
	"encoding/json"
	"mime"
	"strings"

	"github.com/c1emon/gcommon/errorx"
	"github.com/imroc/req/v3"
)

type resultEnvelope struct {
	Code int    `json:"code"`
	Msg  string `json:"msg,omitempty"`
	Data any    `json:"data,omitempty"`
}

// Error maps a JSON business envelope into [errorx.HttpError].
// If strictContentType is true, parsing only happens when Content-Type indicates JSON.
func Error(strictContentType bool) req.ResponseMiddleware {
	return func(_ *req.Client, resp *req.Response) error {
		if strictContentType && !isJSONContentType(resp) {
			return nil
		}

		b, err := resp.ToBytes()
		if err != nil {
			return errorx.NewIOError(err)
		}

		r := &resultEnvelope{}
		if err := json.Unmarshal(b, r); err != nil {
			// Not a JSON envelope; leave response as-is.
			return nil
		}
		if r.Code != 0 {
			return errorx.NewHttpError(resp.StatusCode, r.Code, r.Msg, r.Data)
		}
		return nil
	}
}

func isJSONContentType(resp *req.Response) bool {
	if resp == nil || resp.Response == nil {
		return false
	}
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		return false
	}
	mediaType, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return false
	}
	return mediaType == "application/json" || strings.HasSuffix(mediaType, "+json")
}

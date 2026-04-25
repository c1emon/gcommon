package httpx

import (
	"encoding/json"

	"github.com/c1emon/gcommon/errorx"
)

// ErrorInterceptor decodes a JSON [Result] envelope and maps business errors to
// [errorx.HttpError]. It uses Response.ToBytes so the payload stays available
// as Response.Bytes for later hooks and for callers (avoid reading
// http.Response.Body directly once req has buffered the body).
func ErrorInterceptor(client *Client, resp *Response) error {
	b, err := resp.ToBytes()
	if err != nil {
		return errorx.NewIOError(err)
	}

	r := &Result[any]{}
	if err := json.Unmarshal(b, r); err != nil {
		return errorx.NewJsonError(err)
	}

	if r.HasError() {
		return errorx.NewHttpError(resp.StatusCode, r.Code, r.Msg, r.Data)
	}

	return nil
}

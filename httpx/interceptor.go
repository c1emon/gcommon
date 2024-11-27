package httpx

import (
	"encoding/json"
	"io"

	"github.com/c1emon/gcommon/errorx"
)

func ErrorInterceptor(client *Client, resp *Response) error {

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return errorx.NewIOError(err)
	}

	r := &Result[any]{}
	err = json.Unmarshal(b, r)
	if err != nil {
		return errorx.NewJsonError(err)
	}

	if r.HasError() {
		return errorx.NewHttpError(resp.StatusCode, r.Code, r.Msg, r.Data)
	}

	return nil
}

package httpx_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/c1emon/gcommon/httpx"
)

func TestClient(t *testing.T) {

	client := httpx.NewClient(httpx.WithBaseUrl("http://baidu.com"), httpx.WithReqInterceptor(func(client *httpx.Client, req *httpx.Request) error {
		fmt.Printf("url=%s\n", client.BaseURL)
		return nil
	}))

	resp, err := client.Req().Get("/")
	if err != nil {
		fmt.Printf("err=%s", err)
		return
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("err=%s", err)
		return
	}
	fmt.Printf("resp=%s", b)
}

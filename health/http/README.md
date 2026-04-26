# health/http (`httphealth`)

提供标准库 **[`http.Handler`](https://pkg.go.dev/net/http#Handler)**，响应 JSON，字段由 hellofresh/health-go 定义（`status`、`timestamp`、`component`、`system` 等）。

## 用法

```go
package main

import (
	"log"
	"net/http"

	"github.com/c1emon/gcommon/health"
	httphealth "github.com/c1emon/gcommon/health/http"
)

func main() {
	h, err := httphealth.Handler(health.Config{
		ServiceName: "order-api",
		Version:     "1.4.2",
	})
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/healthz", h)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

包名 **`httphealth`**，建议始终用 **`import httphealth "github.com/c1emon/gcommon/health/http"`** 避免与标准库 `net/http` 混淆。

- **`ServiceName`**：必填，对应 JSON 里 `component.name`。
- **`Version`**：可选，对应 `component.version`。

## 与 `MustHandler`

若希望在启动阶段直接 `panic`（例如配置来自常量），可使用 **`MustHandler`**。

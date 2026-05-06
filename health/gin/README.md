# health/gin (`healthgin`)

提供 **[`gin.HandlerFunc`](https://pkg.go.dev/github.com/gin-gonic/gin#HandlerFunc)**，底层与 [`health/http`](../http) 相同，通过 **`gin.WrapH`** 挂载标准库的 health `http.Handler`。

## 用法

```go
package main

import (
	"log"

	"github.com/c1emon/gcommon/health/v2"
	healthgin "github.com/c1emon/gcommon/health/v2/gin"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.New()
	h, err := healthgin.Handler(health.Config{
		ServiceName: "order-api",
		Version:     "1.4.2",
	})
	if err != nil {
		log.Fatal(err)
	}
	r.GET("/healthz", h)
	_ = r.Run(":8080")
}
```

- **`ServiceName`**：必填。
- **`Version`**：可选。

## `MustHandler`

配置固定、希望在错误时直接崩溃时：

```go
r.GET("/healthz", healthgin.MustHandler(health.Config{ServiceName: "order-api"}))
```

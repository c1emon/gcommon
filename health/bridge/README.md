# health/bridge

对接 [`github.com/hellofresh/health-go/v5`](https://github.com/hellofresh/health-go) 的 **`NewHealth`**：根据根包 [`health.Config`](../config.go) 注册组件信息并开启 **`WithSystemInfo`**。

业务代码请优先使用 **[`health/http`](../http)** 或 **[`health/gin`](../gin)**；本包仅在需要自行基于 `*health.Health` 扩展（例如额外 `Register` 检查）时引用。

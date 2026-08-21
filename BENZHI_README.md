# task149-tsnsched Benzhi 评测说明

这是一个无外部服务依赖的 Go/SQLite 项目。评测镜像从 `benzhi.Dockerfile` 构建，源码根目录直接包含 `go.mod`、`web/`、`cmd/` 和 `internal/`。

```bash
GOTOOLCHAIN=local go test ./...
GOTOOLCHAIN=local go vet ./...
GOTOOLCHAIN=local go build ./...
GOTOOLCHAIN=local go run ./cmd/tsnsched --smoke-test --db :memory:
```

HTTP 服务启动后访问 `/` 可打开最小工程师操作页面；真实调度能力通过节点、端口、定向链路、实时流、草稿验证和原子提交 API 完成。SQLite 文件保存拓扑、草稿、分配、验证证据和活动版本，重新打开同一数据库即可恢复状态。

评测应分别验证 `linux/amd64` 与 `linux/arm64` 镜像构建；项目不需要外部数据库、消息队列或网络服务。

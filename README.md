# hwj-macgo-0019 重试预算状态引擎

本项目实现一个离线可运行的 Go 重试预算状态引擎，管理操作类别、预算窗口、尝试记录、结果和退避预留。

## 构建与测试

```bash
go build ./...
go test ./...
go test -race ./...
go vet ./...
```

## 离线自检

```bash
go run ./cmd/retryengine --self-check
```

## 目录结构

- `cmd`: 可执行入口、配置装配和离线 self-check
- `domain`: 聚合、值对象、状态机、跨实体不变量和领域错误
- `application`: 用例编排、事务边界、幂等与批量部分失败
- `repository`: 仓储接口、本地实现、乐观锁和深拷贝隔离
- `journal`: 带版本、长度和校验的写前日志
- `recovery`: 快照轮换、日志重放、安全截断和恢复校验
- `scheduler`: 后台任务、取消、超时、重试和退避
- `query`: 过滤、稳定排序、分页和派生查询
- `audit`: 审计事件、哈希链和导出校验
- `internal/clockidcodec`: 时钟、标识、编码和共享基础约束

## 配置

配置通过环境变量注入，当前支持的变量包括：

- `HWJ_LISTEN_ADDR`：监听地址，默认 `127.0.0.1:0`
- `HWJ_DATA_DIR`：数据目录，默认 `./data`
- `HWJ_BUDGET_WINDOW`：预算窗口时长，默认 `24h`
- `HWJ_CLOCK_SKEW`：允许时钟偏差，默认 `100ms`
- `HWJ_MAX_CONCURRENT`：最大并发数，默认 `16`
- `HWJ_SHUTDOWN_TIMEOUT`：关闭超时，默认 `5s`

所有时长值使用 Go `time.ParseDuration` 格式。

## 设计基线

详细设计基线见 `docs/PROJECT_DESIGN.md`。

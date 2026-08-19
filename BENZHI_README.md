# 本质评测环境说明

## 项目

- 项目编号：`hwj-macgo-0019`
- 项目名称：重试预算状态引擎
- 项目说明：管理操作类别、预算窗口、尝试记录、退避预留和结果，提供可恢复的重试预算控制。

## 固定环境

- Go toolchain：`go1.26.5`
- go.mod language version：`go 1.21`
- GOTOOLCHAIN：`local`
- 支持平台：`linux/amd64`、`linux/arm64`
- Docker 基础镜像：`golang:1.26.5-bookworm`
- Docker manifest：`golang@sha256:53eeac89074db483fdf0ab3be1df32bf6e47562263d2d0d6baa7f26acb4957dd`

## 构建

```bash
./build_benzhi_docker.sh hwj-macgo-0019:benzhi-amd64 linux/amd64
./build_benzhi_docker.sh hwj-macgo-0019:benzhi-arm64 linux/arm64
```

## 运行

```bash
docker run --rm -it --network none hwj-macgo-0019:benzhi-amd64 bash
```

## 容器内验证

```bash
go version
go env GOTOOLCHAIN GOPROXY GOMODCACHE GOCACHE
go test ./...
go vet ./...
go build ./...
```

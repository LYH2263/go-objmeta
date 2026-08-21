# go-objmeta

本地对象元数据层：Put/Get/Head/Delete、multipart 分片上传、前缀列举与条件写。

## Build / Test

```bash
go build ./...
go test ./... -count=1
```

## Daemon

```bash
go run ./cmd/objd -addr :8103 -root ./data -web web
```

管理页：http://127.0.0.1:8103/

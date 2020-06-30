# Echo Proxy

Echo Proxy 是 mixin api 的一个代理，
可以在没有 Auth Token 的情况下访问 ```api.mixin.one``` 的所有 GET 请求。适用于在没有 ```keystore``` 的情况下访问
用户详情，汇率等接口。

## Requirement

Go v1.15

## Usage

* Run `go run main.go --key keystore_path.json --port 8888`

See [main.go](main.go) for details


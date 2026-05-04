# wxcloudrun-golang
[![CI](https://github.com/xavierxcn/wxcloud-server/actions/workflows/ci.yml/badge.svg)](https://github.com/xavierxcn/wxcloud-server/actions/workflows/ci.yml)
[![GitHub license](https://img.shields.io/github/license/WeixinCloud/wxcloudrun-express)](https://github.com/WeixinCloud/wxcloudrun-express)
![Go version](https://img.shields.io/badge/golang-1.26.2-green)

微信云托管 Go 服务，用于承载公众号开放接口代理，并保留模板计数器接口。

![](https://qcloudimg.tencent-cloud.cn/raw/be22992d297d1b9a1a5365e606276781.png)


## 快速开始
前往 [微信云托管快速开始页面](https://developers.weixin.qq.com/miniprogram/dev/wxcloudrun/src/basic/guide.html)，选择相应语言的模板，根据引导完成部署。

## 本地调试

```bash
go test ./...
PORT=8080 go run .
```

本地计数器接口需要数据库环境变量，可参考 `.env.example`。公众号代理接口不依赖数据库。

## 实时开发
代码变动时，不需要重新构建和启动容器，即可查看变动后的效果。请参考[微信云托管实时开发指南](https://developers.weixin.qq.com/miniprogram/dev/wxcloudrun/src/guide/debug/dev.html)

## Dockerfile最佳实践
请参考[如何提高项目构建效率](https://developers.weixin.qq.com/miniprogram/dev/wxcloudrun/src/scene/build/speed.html)

## GitHub Actions

`.github/workflows/ci.yml` 会在 `main` 分支 push、pull request 和手动触发时运行：

- `go mod download`
- `go test ./...`
- `go build ./...`
- `docker build -t wxcloud-server:<commit> .`

## 目录结构说明
~~~
.
├── Dockerfile                Dockerfile 文件
├── LICENSE                   LICENSE 文件
├── README.md                 README 文件
├── container.config.json     模板部署「服务设置」初始化配置（二开请忽略）
├── db                        数据库逻辑目录
├── go.mod                    go.mod 文件
├── go.sum                    go.sum 文件
├── index.html                主页 html 
├── main.go                   主函数入口
└── service                   接口服务逻辑目录
~~~


## 服务 API 文档

### `GET /api/count`

获取当前计数

#### 请求参数

无

#### 响应结果

- `code`：错误码
- `data`：当前计数值

##### 响应结果示例

```json
{
  "code": 0,
  "data": 42
}
```

#### 调用示例

```
curl https://<云托管服务域名>/api/count
```



### `POST /api/count`

更新计数，自增或者清零

#### 请求参数

- `action`：`string` 类型，枚举值
  - 等于 `"inc"` 时，表示计数加一
  - 等于 `"clear"` 时，表示计数重置（清零）

##### 请求参数示例

```
{
  "action": "inc"
}
```

#### 响应结果

- `code`：错误码
- `data`：当前计数值

##### 响应结果示例

```json
{
  "code": 0,
  "data": 42
}
```

#### 调用示例

```
curl -X POST -H 'content-type: application/json' -d '{"action": "inc"}' https://<云托管服务域名>/api/count
```

### `GET /wechat/freepublish/batchget`

通过微信云托管开放接口服务调用微信公众号已发布内容列表接口。

#### 前置条件

- 已在「服务管理 / 云调用」开启开放接口服务。
- 已在「微信令牌权限」配置 `/cgi-bin/freepublish/batchget`。
- 服务已重新发布，使云调用配置对线上版本生效。

#### 响应结果

直接透传微信接口响应。

#### 调用示例

```
curl https://<云托管服务域名>/wechat/freepublish/batchget
```

#### 注意

服务内部请求 `http://api.weixin.qq.com/cgi-bin/freepublish/batchget`，不手动拼接 `access_token`。

## 环境变量

服务启动不依赖 MySQL。只有访问 `/api/count` 时才会初始化数据库；如果没有配置数据库，该接口会返回 JSON 错误，不会导致容器退出。

如需使用模板计数器接口，请在「服务设置」中补全以下环境变量：

- MYSQL_ADDRESS
- MYSQL_PASSWORD
- MYSQL_USERNAME
- MYSQL_DATABASE，默认 `golang_demo`

不要把数据库密码写入代码、README 或 `container.config.json`。

## License

[MIT](./LICENSE)

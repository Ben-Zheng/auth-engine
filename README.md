# api-engine

# # 代码检查

`docker run --rm -v ./:/app -w /app release.daocloud.io/aigc/gitlab-ci:v1.0.6  make staticcheck`

# 运行

```shell
go run .
```

# 编译

```shell
go build github.com/auth-engine

./api-engine
```

# 构建镜像
```shell
docker buildx build --platform=linux/amd64 -t hub.intranet.daocloud.io/pufa/auth-engine:${version} . --push
```

# 部署

```shell
kubectl apply -f deployments/*.yaml
```

# 开发目录

```shell
(base) ➜ auth-engine tree -a -I "vendor"
.
├── .gitignore
├── .gitlab-ci.yml
├── Dockerfile  # 开发环境 dockerfile
├── Dockerfile.release  # 生产环境 dockerfile
├── Makefile
├── README.md
├── atlas.hcl  # atlas配置文件
├── cmd
│   └── migrator
│       └── main.go # 迁移文件 main 文件
├── config  # 配置文件
│   ├── config.go 
│   ├── config.yaml 
│   └── kubeconfig # kubeconfig文件
├── deployments  # 部署文件
│   ├── configmap.yaml 
│   ├── deployment.yaml 
│   ├── service.yaml 
│   └── gproduct-proxy.yaml
├── docs
│   ├── docs.go # swagger配置文件
│   ├── swagger.json
│   └── swagger.yaml
├── go.mod
├── go.sum
├── internal # 业务代码
│   └── pkg
│       ├── dao # Dao 层，数据库表操作
│       │   ├── core.go  
│       │   ├── token.go
│       │   ├── json.go
│       │   └── mysql.go
│       ├── global  # 全局相关文件
│       │   ├── init.go
│       │   └── servers
│       │       └── server.go
│       ├── model # API层，API 数据结构定义
│       │   ├── atuh.go
│       │   └── common.go
│       ├── routers # 路由
│       │   ├── ctrl
│       │   │   ├── common
│       │   │   │   ├── common.go # 错误码、请求参数、返回值
│       │   │   │   ├── error.go 
│       │   │   │   ├── middleware.go # 中间件
│       │   │   │   └── request.go # 请求参数
│       │   │   ├── token.go # 业务路由
│       │   │   └── swagger.go
│       │   └── routeinit
│       │       └── route_init.go # 路由初始化
│       ├── services # Service 层，业务逻辑
│       │   ├── client_service # Client 层，调用第三方服务
│       │   └── token.go # Service 层，业务逻辑
│       └── utils # 工具类
│           ├── common.go
│           └── json
│               └── json.go
├── main.go
├── migrations # 迁移文件
├── pkg
│   ├── clients # Client 层，调用第三方服务
│   │   ├── dce_client.go
│   │   └── insight.go
│   ├── constants # 常量
│   │   └── const.go
│   └── log # 日志
│       └── log.go
├── scripts # 脚本
│   ├── get-version.sh
│   ├── util.sh
│   └── verify-staticcheck.sh
├── test # 测试文件
├── tools.go
├── wire_gen.go # 运行 `go generate` 或者 `wire` 生成
└── wrie.go  # wire依赖注入
```

# 模块规范

若是新作一个模块，请遵循以下规范：
1. dao层，新建一个文件，文件名为表名，如xxx.go，并实现对应表的定义；
2. model层，新建一个文件，文件名为表名，如xxx.go，并实现对应 API 结构定义；
3. service层，新建一个文件，文件名为表名，如xxx.go，并实现对应业务逻辑；
4. router层，新建一个文件，文件名为表名，如xxx.go，并实现对应路由定义 和 controller；
   > 注意：
   > 没定义一个 APIHandler，需要在 routeinit/route_init.go 中注册
5. wire：依赖注入，以上dao、service、roter 层的定义，需要在 wrie.go 中注册。
6. 每次修改 wire.go 文件后，需要运行 `go generate` 或者 `wire` 更新 wire_gen.go 文件
    > 安装 wire：`go install github.com/google/wire/cmd/wire@latest`
7. swagger 文档：在 docs/docs.go 中注册, 通过运行 `go generate` 生成 docs/swagger.json

# go mod

```shell
go mod init github.com/auth-engine （第一次生成，后续无需执行）
go mod tidy
go mod vendor
```
注意：每次修改依赖后，都需要执行 `go mod tidy`，当依赖是私有仓库时，需要配置 GOPRIVATE, 如
`export GOPRIVATE=gitlab.daocloud.cn/*`

安装单个依赖包，执行如`go get -u github.com/xxx@v1.0.0`


# 数据库迁移
确保本地已经有 atlas，如果没有请执行以下命令安装：`curl -sSf https://atlasgo.sh | sh`

1. 在`cmd/migrator/main.go`文件中注册model

2. 在migrations中新建一个文件夹，名称为版本号，修改[atlas.hcl](atlas.hcl)文件

   ```go
   migration {
      dir = "file://migrations"
   }
   ```
   修改成
   
   ```go
   migration {
      dir = "file://migrations/版本号"
   }
   ```
3. 生成迁移sql文件（生成增量SQL）

   修改项目根目录下的atlas.hcl文件，将dir字段修改为 `file://migrations`,执行命令

   ```shell
   atlas migrate diff --env gorm  v2.3.0(每次发版更新)
   ```

   !!! note
       -mod may only be set to readonly or vendor when in workspace mode, but it is set to "mod" 
       解决方法：
         ```shell
         export GOWORK=off
         atlas migrate diff --env gorm v1.0.0
         ``` 
4. 生成全量的sql文件

   修改项目根目录下的atlas.hcl文件，将dir字段修改为 file://migrations/v2.1.4 （修改为你的版本号）。执行命令：

   ```shell
   atlas migrate diff --env gorm  init
   ```
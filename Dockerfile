FROM release.daocloud.io/aigc/golang:1.23-git AS build-env

ENV GO111MODULE="on"
ENV GOFLAGS="-mod=vendor"
ENV GOPROXY="https://goproxy.cn"
ENV GOPRIVATE="gitlab.daocloud.cn"
WORKDIR /go/src/github.com/auth-engine
ADD . .
# 生成
RUN CGO_ENABLED=0 go generate
# 增加设置 版本信息, 默认根据git自动执行, 想要自定义的话, 可以将下面命令中的 $(git xxx) 替换为自定义的值, 不设置的话为unknown
# RUN CGO_ENABLED=0 go build -ldflags "-X 'github.com/auth-engine/pkg/info.EngineVersion=$(git describe --abbrev=0 --tags --always)' -X 'github.com/auth-engine/pkg/info.GitCommitID=$(git rev-parse --short HEAD)'" -o /go/engine .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X 'github.com/auth-engine/pkg/info.EngineVersion=v0.1.0'" -o /go/engine .

FROM release.daocloud.io/aigc/debian:12.6-slim-base-pkg
COPY --from=build-env /go/engine /bin/engine

ENV TZ="Asia/Shanghai"
ENV LANG="en_US.UTF-8"
ENV LC_ALL="en_US.UTF-8"

EXPOSE 8888
CMD ["/bin/engine"]
FROM golang:1.13.10 as builder

ARG ssh_prv_key

ARG ssh_pub_key

ARG PKG=polardb-sms

ARG DIR=$GOPATH/src/$PKG

WORKDIR $DIR

COPY go.mod go.sum ./

RUN go env -w GOPROXY=https://goproxy.cn,direct && \

RUN go mod download

COPY / $DIR

ARG CodeSource='https://github.com/ApsaraDB/PolarDB-Stack-Storage.git'

ARG CodeBranch='git branch --show-current'

ARG CodeVersion='git rev-parse HEAD'

ARG BuildDate='date -u +"%Y-%m-%dT%H:%M:%SZ"'

RUN CGO_ENABLED=0 go build -ldflags "-X '$PKG/pkg/version.GitBranch=$CodeBranch' \
                                     -X '$PKG/pkg/version.GitCommit=$CodeVersion'\
                                     -X '$PKG/pkg/version.BuildDate=$BuildDate' \
                                     -X '$PKG/pkg/version.Module=$CodeSource'" -o polardb-sms-manager $DIR/cmd/manager

FROM alpine:3.12

ARG CodeSource='https://github.com/ApsaraDB/PolarDB-Stack-Storage.git'

ARG CodeBranch='git branch --show-current'

ARG CodeVersion='git rev-parse HEAD'

ARG BuildDate='date -u +"%Y-%m-%dT%H:%M:%SZ"'

ARG APK_MIRROR=mirrors.aliyun.com

WORKDIR /bin/

ENTRYPOINT ["/bin/polardb-sms-manager"]

LABEL CodeSource=$CodeSource CodeBranch=$CodeBranch CodeVersion=$CodeVersion BuildDate=$BuildDate

COPY --from=builder /go/src/polardb-sms/polardb-sms-manager /bin/polardb-sms-manager

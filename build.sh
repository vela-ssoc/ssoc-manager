#!/bin/bash

# 1. 获取程序名。
DIR_NAME=$(basename $(pwd))
TZ="Asia/Shanghai"
NOW=$(date)
VER=$(date -d "$NOW" +"%y.%-m.%-d-%H%M%S")
BIN_NAME=${DIR_NAME}"-"v$VER$(go env GOEXE)
echo "程序名为："${BIN_NAME}

# 2. 如果执行的是清理命令，清理完就退出。
go clean -cache
if [ "$1" = "clean" ]; then
    rm -rf ${DIR_NAME}*
    echo "清理结束"
    exit 0
fi

export CGO_ENABLED=0
LDFLAGS="-s -w -extldflags -static -X 'github.com/vela-ssoc/vela-manager/banner.compileTime=$NOW'"
go build -o ${BIN_NAME} -trimpath -v -ldflags "$LDFLAGS" ./main

echo "编译结束"

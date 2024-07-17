#!/bin/bash

echo ""
echo "\033[1;33m   ______________  _____  \033[0m"
echo "\033[1;33m  / ___/ ___/ __ \/ ___/  \033[0m"
echo "\033[1;33m (__  |__  ) /_/ / /__    \033[0m"
echo "\033[1;33m/____/____/\____/\\___/   \033[0m \033[1;33m\033[0m \033[34mCOMPILE\033[0m"
echo "    Powered By: 东方财富安全团队"
echo ""

# 日志
# 定义颜色代码
DEFAULT="\033[0m"
RED="\033[0;31m"
GREEN="\033[0;32m"
YELLOW="\033[0;33m"
BLUE="\033[0;34m"

# 定义不同级别的日志函数
log_success() {
  local message="$1"
  echo -e "${GREEN}[SUCC] $message${DEFAULT}"
}

log_info() {
  local message="$1"
  echo -e "${BLUE}[INFO] $message${DEFAULT}"
}

log_warning() {
  local message="$1"
  echo -e "${YELLOW}[WARN] $message${DEFAULT}"
}

log_error() {
  local message="$1"
  echo -e "${RED}[ERRO] $message${DEFAULT}" >&2
}

#PS3="请选择操作系统："
#GOOS_OPTS=("linux" "windows" "darwin")
#select goos in "${GOOS_OPTS[@]}"; do
#  case $goos in
#  "linux")
#    export GOOS="linux"
#    break
#    ;;
#  "windows")
#    export GOOS="windows"
#    break
#    ;;
#  "darwin")
#    export GOOS="darwin"
#    break
#    ;;
#  *)
#    log_error "输入了无效的参数。"
#    exit 1
#    ;;
#  esac
#done

GOOS=$(go env GOOS)
ARCH=$(go env GOARCH)
GOPROXY="https://goproxy.cn,direct"

# 执行编译
export GOOS=$GOOS   # 设置 GOOS
export GOARCH=$ARCH # 设置 GOARCH

# 设置 GOPROXY
GOPROXY_BAK=$(go env GOPROXY)
log_info "已将 GOPROXY 设置为：$GOPROXY"
go env -w GOPROXY=$GOPROXY

log_info "缓存清理......"
go clean --cache

NOW=$(date)
if [ $(uname) = "Linux" ]; then
  NOW=$(date --iso-8601=seconds)
fi

LDFLAGS="-s -w -extldflags -static -X 'github.com/vela-ssoc/vela-manager/infra/banner.compileTime=$NOW'"

log_warning "编译信息：${GOOS}-${GOARCH}"
log_info "正在编译......"

export CGO_ENABLED=0

CURRENT_DATE=$(date +'%Y%m%d')
BINARY_NAME=ssoc-manager-${CURRENT_DATE}$(go env GOEXE)

go build -o ${BINARY_NAME} -ldflags "$LDFLAGS" \
  -gcflags="all=-trimpath=${PWD}" \
  -asmflags="all=-trimpath=${PWD}" \
  ./main


if [ $? -eq 0 ]; then
    # 检查 upx 命令是否存在
    if command -v upx &> /dev/null; then
        upx -9 ${BINARY_NAME}
    fi
fi

if [ $? -eq 0 ]; then
  log_success "编译成功."
else
  log_error "编译失败!"
fi

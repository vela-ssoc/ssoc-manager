package banner

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

// ANSI 打印 banner 到指定输出流。
//
// 何为 ANSI 转义序列：https://en.wikipedia.org/wiki/ANSI_escape_code.
func ANSI(w io.Writer) {
	onceParseArgs()

	_, _ = fmt.Fprintf(w, ansiLogo, version, pid, runtime.GOOS, runtime.GOARCH,
		hostname, username, workdir, compileAt, commitAt, path, revision)
}

const ansiLogo = "\033[1;33m" +
	"   ______________  _____  \n" +
	"  / ___/ ___/ __ \\/ ___/ \n" +
	" (__  |__  ) /_/ / /__    \n" +
	"/____/____/\\____/\\___/  \033[0mMANAGER\n" +
	"\033[1;31m:: Vela Security ::\033[0m     %s\n\n" +
	"    \033[1;36m进程 PID:\033[0m %d\n" +
	"    \033[1;36m操作系统:\033[0m %s\n" +
	"    \033[1;36m系统架构:\033[0m %s\n" +
	"    \033[1;36m主机名称:\033[0m %s\n" +
	"    \033[1;36m当前用户:\033[0m %s\n" +
	"    \033[1;36m工作目录:\033[0m %s\n" +
	"    \033[1;36m编译时间:\033[0m %s\n" +
	"    \033[1;36m提交时间:\033[0m %s\n" +
	"    \033[1;36m修订版本:\033[0m https://%s/tree/%s\n\n\n"

var (
	// version 项目发布版本号
	// 项目每次发布版本后会打一个 tag, 这个版本号就来自 git 最新的 tag
	version string

	// revision 修订版本, 代码最近一次的提交 ID
	revision string

	// compileAt 编译时间, 由编译脚本在编译时 -X 写入。
	compileTime string

	compileAt time.Time

	// commitAt 代码最近一次提交时间
	commitAt time.Time

	path string

	// pid 进程 ID
	pid int

	// username 当前系统用户名
	username string

	workdir string

	// hostname 主机名
	hostname string

	onceParseArgs = sync.OnceFunc(parseArgs)
)

// parseArgs 处理编译与运行时参数
func parseArgs() {
	pid = os.Getpid() // 获取 PID
	if cu, _ := user.Current(); cu != nil {
		username = cu.Username
	}
	hostname, _ = os.Hostname()
	workdir, _ = os.Getwd()
	compileAt = parseLocalTime(compileTime)

	info, _ := debug.ReadBuildInfo()
	if info == nil {
		return
	}

	path = info.Main.Path
	if version == "" {
		v := info.Main.Version
		v = strings.TrimLeft(v, "(")
		version = strings.TrimRight(v, ")")
	}

	settings := info.Settings
	for _, set := range settings {
		key, val := set.Key, set.Value
		switch key {
		case "vcs.revision":
			revision = val
		case "vcs.time":
			commitAt = parseLocalTime(val)
		}
	}
}

// parseLocalTime 给定一个字符串格式化为当前地区的时间。
//
// - time.RFC1123Z Linux `date -R` 输出的时间格式。
// - time.UnixDate macOS `date` 输出的时间格式。
func parseLocalTime(str string) time.Time {
	for _, layout := range []string{
		time.RFC1123Z, time.UnixDate, time.Layout, time.ANSIC,
		time.RubyDate, time.RFC822, time.RFC822Z, time.RFC850,
		time.RFC1123, time.RFC3339, time.RFC3339Nano, time.Kitchen,
		time.Stamp, time.StampMilli, time.StampMicro, time.StampNano,
		time.DateTime, time.DateOnly,
	} {
		if at, err := time.Parse(layout, str); err == nil {
			return at.Local()
		}
	}
	epoch := time.Date(0, 1, 1, 0, 0, 0, 0, time.Local)

	return epoch
}

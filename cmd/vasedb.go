// Copyright 2022 Leon Ding <ding@ibyte.me> https://vasedb.github.io

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/auula/vasedb/clog"
	"github.com/auula/vasedb/conf"
	"github.com/auula/vasedb/server"
	"github.com/auula/vasedb/utils"
	"github.com/auula/vasedb/vfs"
	"github.com/fatih/color"
)

const (
	version = "v0.1.1"
	website = "https://vasedb.github.io"
)

var (
	//go:embed banner.txt
	banner    string
	greenFont = color.New(color.FgMagenta)
	logo      = greenFont.Sprintf(banner, version, website)
	daemon    = false
)

// 初始化全局需要使用的组件
func init() {
	// 打印 Banner 信息
	fmt.Println(logo)

	// 解析命令行输入的参数，默认命令行参数优先级最高，但是相对于能设置参数比较少
	fl := parseFlags()

	// 根据命令行传入的配置文件地址，覆盖掉默认的配置
	if fl.config != conf.DefaultConfig.FilePath {
		if err := conf.Load(fl.config, conf.Settings); err != nil {
			clog.Failed(err)
		}
	}

	if fl.debug {
		conf.Settings.Debug = fl.debug
		// 覆盖掉默认 DEBUG 选项
		clog.IsDebug = conf.Settings.Debug
	}

	if fl.auth != conf.DefaultConfig.Password {
		conf.Settings.Password = fl.auth
	}

	if fl.path != conf.DefaultConfig.Path {
		conf.Settings.Path = fl.path
	}

	if fl.port != conf.DefaultConfig.Port {
		conf.Settings.Port = fl.port
	}

	clog.Debug(fmt.Sprintf("%+v\n", conf.Settings))

	// 设置一下运行过程中日志输出文件的路径
	clog.SetPath(conf.Settings.Logging)

	if err := vfs.InitFS(conf.Settings.Path); err != nil {
		clog.Failed(err)
	}
}

// flags 优先级别最高的参数，从命令行传入
type flags struct {
	auth   string
	port   int
	path   string
	config string
	debug  bool
}

// parseFlags 解析从命令行启动输入的主要参数
func parseFlags() (fl *flags) {
	fl = new(flags)
	flag.StringVar(&fl.auth, "auth", conf.DefaultConfig.Password, "--auth specify the server authentication password.")
	flag.StringVar(&fl.config, "config", conf.DefaultConfig.FilePath, "--config specify the configuration file path.")
	flag.StringVar(&fl.path, "path", conf.DefaultConfig.Path, "--path specify the data storage directory.")
	flag.IntVar(&fl.port, "port", conf.DefaultConfig.Port, "--port specify the HTTP server port.")
	flag.BoolVar(&fl.debug, "debug", conf.DefaultConfig.Debug, "--debug whether to enable debug mode.")
	flag.BoolVar(&daemon, "daemon", false, "--daemon whether to run with a daemon.")
	flag.Parse()
	return
}

func main() {
	// 检查是否启用了守护进程模式
	if daemon {
		// 后台守护进程模式启动，创建一个与当前程序相同的命令
		cmd := exec.Command(os.Args[0], utils.SplitArgs(utils.TrimDaemon(os.Args))...)

		// 如果需要传递环境变量信息
		cmd.Env = os.Environ()

		// 从当前进程启动守护进程
		if err := cmd.Start(); err != nil {
			clog.Failed(err)
		}

		clog.Info(fmt.Sprintf("Daemon launched PID: %d", cmd.Process.Pid))
	} else {
		// 开始执行正常的 vasedb 逻辑，这里会启动 HTTP 服务器让客户端连接
		hs := server.New(conf.Settings)
		hs.Startup()
	}
}

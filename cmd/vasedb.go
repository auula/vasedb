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

package cmd

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

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
	if conf.HasCustom(fl.config) {
		err := conf.Load(fl.config, conf.Settings)
		if err != nil {
			clog.Failed(err)
		}
		clog.Info("Loading custom config file was successfully")
	}

	if fl.debug {
		conf.Settings.Debug = fl.debug
		// 覆盖掉默认 DEBUG 选项
		clog.IsDebug = conf.Settings.Debug
	}

	// 命令行传入的密码优先级最高
	if fl.auth != conf.Default.Password {
		conf.Settings.Password = fl.auth
	} else {
		// 如果命令行没有传入密码，系统随机生成一串 16 位的密码
		conf.Settings.Password = utils.RandomString(16)
		clog.Infof("The default password is: %s", conf.Settings.Password)
	}

	// 设置数据存储路径和端口
	if fl.path != conf.Default.Path {
		conf.Settings.Path = fl.path
	}

	if fl.port != conf.Default.Port {
		conf.Settings.Port = fl.port
	}

	clog.Debug(conf.Settings)

	var err error = nil
	// 设置一下运行过程中日志输出文件的路径
	err = clog.SetOutput(conf.Settings.LogPath)
	if err != nil {
		clog.Failed(err)
	}

	clog.Info("Initialize log file output successfully")

	// 设置数据文件存储位置和相关的文件系统，移到 vfs 手动初始化
	err = vfs.SetupFS(conf.Settings.Path)
	if err != nil {
		clog.Failed(err)
	}

	clog.Info("Setup file system was successfully")
}

func StartApp() {
	// 检查是否启用了守护进程模式
	if daemon {
		runAsDaemon()
	} else {
		runServer()
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

// runAsDaemon 以守护进程模式运行
// 后台守护进程模式启动，创建一个与当前程序相同的命令
func runAsDaemon() {
	args := splitArgs(trimDaemon(os.Args))
	cmd := exec.Command(os.Args[0], args...)
	cmd.Env = os.Environ()

	err := cmd.Start()
	if err != nil {
		clog.Failed(err)
	}

	clog.Infof("Daemon launched PID: %d", cmd.Process.Pid)
}

// runServer 启动 Http API 服务器
func runServer() {
	hts, err := server.New(&server.Options{
		Port: conf.Settings.Port,
		Auth: conf.Settings.Password,
	})
	if err != nil {
		clog.Failed(err)
	}

	// vfs.OpenFS() 合并 vfs.SetupFS()
	fss, err := vfs.OpenFS()
	if err != nil {
		clog.Failed(err)
	}

	server.SetupFS(fss)

	go func() {
		err := hts.Startup()
		if err != nil {
			clog.Failed(err)
		}
	}()

	time.Sleep(500 * time.Millisecond)
	clog.Infof("HTTP server started at http://%s:%d 🚀", hts.IPv4(), hts.Port())

	select {}
}

// parseFlags 解析从命令行启动输入的主要参数
func parseFlags() (fl *flags) {
	fl = new(flags)
	flag.StringVar(&fl.auth, "auth", conf.Default.Password, "--auth specify the server authentication password.")
	flag.StringVar(&fl.config, "config", "", "--config specify the configuration file path.")
	flag.StringVar(&fl.path, "path", conf.Default.Path, "--path specify the data storage directory.")
	flag.IntVar(&fl.port, "port", conf.Default.Port, "--port specify the HTTP server port.")
	flag.BoolVar(&fl.debug, "debug", conf.Default.Debug, "--debug whether to enable debug mode.")
	flag.BoolVar(&daemon, "daemon", false, "--daemon whether to run with a daemon.")
	flag.Parse()
	return
}

// TrimDaemon 从 os.Args 中移除 "-daemon" 参数
func trimDaemon(args []string) []string {
	var newArgs []string

	// 遍历 args 切片
	for i := 1; i < len(args); i++ {
		// 当发现参数是符合标准时跳过当前参数
		if args[i] == "-daemon" || args[i] == "--daemon" {
			continue
		}
		newArgs = append(newArgs, args[i])
	}

	return newArgs
}

// SplitArgs 处理命令行参数以 = 分割
func splitArgs(args []string) []string {
	var newArgs []string

	for i := 1; i < len(args); i++ {
		// 分割 args 中的元素，否则命令行解析错误
		if strings.Contains(args[i], "=") && strings.Count(args[i], "=") == 1 {
			newArgs = append(newArgs, strings.Split(args[i], "=")...)
		} else {
			// 过滤掉 == 不合法过滤掉
			if strings.Count(args[i], "=") > 1 {
				continue
			}
			newArgs = append(newArgs, strings.Split(args[i], "=")...)
		}
	}

	return newArgs
}

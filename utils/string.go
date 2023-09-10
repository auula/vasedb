package utils

import "strings"

// TrimDaemon 从 os.Args 中移除 "-daemon" 参数
func TrimDaemon(args []string) []string {
	var newArgs []string

	// 遍历 args 切片
	for i := 1; i < len(args); i++ {
		// 巨坑无比，无法知道用户输入的参数是 -- 还是 - 开头
		if args[i] == "-daemon" || args[i] == "--daemon" {
			// 当发现 "-daemon" 参数时，跳过当前参数
			continue
		}
		newArgs = append(newArgs, args[i])
	}

	return newArgs
}

// SplitArgs 处理命令行参数以 = 分割
func SplitArgs(args []string) []string {
	var newArgs []string
	// i = 1 避免处理 ./vasedb 本身
	for i := 1; i < len(args); i++ {
		// 巨坑无比，如果没有这段代码 newArgs 元素中会出现 --port=2468 元素
		// 而不是分割好的 [--port , 2468] 这样的元素，否则命令行解析错误
		if strings.Contains(args[i], "=") && strings.Count(args[i], "=") == 1 {
			newArgs = append(newArgs, strings.Split(args[i], "=")...)
		} else {
			// 如果是 --port==8080 ，过滤掉 == 不合法过滤掉
			if strings.Count(args[i], "=") > 1 {
				continue
			}
			newArgs = append(newArgs, strings.Split(args[i], "=")...)
		}
	}

	return newArgs
}

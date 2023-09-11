package utils

import "strings"

// TrimDaemon 从 os.Args 中移除 "-daemon" 参数
func TrimDaemon(args []string) []string {
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
func SplitArgs(args []string) []string {
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

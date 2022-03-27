package common

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// 会员绑定邮箱日志
type taskLog struct {
	Flag string `msg:"flag" json:"flag"`
	Info string `msg:"info" json:"info"`
	Date string `msg:"date" json:"date"`
}

func Log(flag, format string, v ...interface{}) {

	_, file, line, _ := runtime.Caller(1)
	paths := strings.Split(file, "/")
	l := len(paths)
	if l > 2 {
		file = paths[l-2] + "/" + paths[l-1]
	}
	path := fmt.Sprintf("%s:%d", file, line)
	msg := fmt.Sprintf(format, v...)

	tm := time.Now().Format("2006-01-02 15:04:05")
	info := fmt.Sprintf("%s|%s|%s", tm, path, msg)

	fmt.Println(info)

	if flag == "" {
		flag = "task"
	}

	//lg := taskLog{
	//	Flag: flag,
	//	Info: info,
	//	Date: tm,
	//}
	//err := zlog.Post(esPrefix+"task_log", lg)
	//if err != nil {
	//	fmt.Println("task log error")
	//}
}

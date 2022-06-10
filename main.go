package main

import (
	"fmt"
	_ "go.uber.org/automaxprocs"
	"os"
	"strings"
	"task/modules/trc20"
)

var (
	gitReversion   = ""
	buildTime      = ""
	buildGoVersion = ""
)

type fn func([]string, string)

var cb = map[string]fn{
	"trc20": trc20.Parse, //rocketMQ trc20查单确认
}

func main() {

	argc := len(os.Args)
	if argc != 5 {
		fmt.Printf("%s <etcds> <cfgPath> [upgrade][transferConfirm][dividend][birthDividend][monthlyDividend][rebate][message] <proxy>\n", os.Args[0])
		return
	}

	endpoints := strings.Split(os.Args[1], ",")

	fmt.Printf("gitReversion = %s\r\nbuildGoVersion = %s\r\nbuildTime = %s\r\n", gitReversion, buildGoVersion, buildTime)

	if val, ok := cb[os.Args[3]]; ok {
		val(endpoints, os.Args[2])
	}

	fmt.Println(os.Args[3], "done")
}

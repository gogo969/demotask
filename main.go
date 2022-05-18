package main

import (
	"fmt"
	_ "go.uber.org/automaxprocs"
	"os"
	"strings"
	"task/modules/banner"
	"task/modules/bonus"
	"task/modules/evo"
	"task/modules/message"
	"task/modules/promo"
	"task/modules/risk"
)

var (
	gitReversion   = ""
	buildTime      = ""
	buildGoVersion = ""
)

type fnP func([]string, string, string)

var cbP = map[string]fnP{
	"promo": promo.Parse, //活动流水更新
}

type fn func([]string, string)

var cb = map[string]fn{
	"banner":  banner.Parse,  //活动流水更新
	"bonus":   bonus.Parse,   //电子游戏奖金池
	"risk":    risk.Parse,    //风控自动派单脚本
	"evo":     evo.Parse,     //evo
	"message": message.Parse, //站内信批量发送
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

	// 带参数的脚本
	if val, ok := cbP[os.Args[3]]; ok {
		val(endpoints, os.Args[2], os.Args[4])
	}

	fmt.Println(os.Args[3], "done")
}

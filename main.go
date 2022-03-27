package main

import (
	"fmt"
	_ "go.uber.org/automaxprocs"
	"os"
	"strings"
	"task/modules/banner"
	"task/modules/bet"
	"task/modules/bonus"
	"task/modules/dividend"
	"task/modules/evo"
	"task/modules/promo"
	"task/modules/risk"
	"task/party"
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
	"banner": banner.Parse, //活动流水更新
	"bonus":  bonus.Parse,  // 电子游戏奖金池
	"bet":    bet.Parse,    //投注
	"risk":   risk.Parse,   //风控自动派单脚本
	"evo":    evo.Parse,    //evo
}

func main() {

	argc := len(os.Args)
	if argc != 5 {
		fmt.Printf("%s <etcds> <cfgPath> [upgrade][transferConfirm][dividend][birthDividend][monthlyDividend][rebate][message] <proxy>\n", os.Args[0])
		return
	}

	endpoints := strings.Split(os.Args[1], ",")

	fmt.Printf("gitReversion = %s\r\nbuildGoVersion = %s\r\nbuildTime = %s\r\n", gitReversion, buildGoVersion, buildTime)

	// 红利单独处理
	// 红利批量发放
	if os.Args[3] == "dividend" {
		// 参数 dividend,socks5://35.240.197.24:9080 dividend_im,socks5://35.240.197.24:9080
		params := strings.Split(os.Args[4], ",")
		if len(params) == 2 {
			party.New(params[1])
			dividend.Parse(endpoints, os.Args[2], params[0])
		}

		return
	}

	// 场馆代理
	party.New(os.Args[4])

	if val, ok := cb[os.Args[3]]; ok {
		val(endpoints, os.Args[2])
	}

	// 带参数的脚本
	if val, ok := cbP[os.Args[3]]; ok {
		val(endpoints, os.Args[2], os.Args[4])
	}

	fmt.Println(os.Args[3], "done")
}

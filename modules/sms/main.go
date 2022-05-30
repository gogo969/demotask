package sms

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/panjf2000/ants/v2"
	cpool "github.com/silenceper/pool"
	"task/contrib/conn"
	"task/modules/common"
	"time"
)

var (
	td       *sqlx.DB
	beanPool cpool.Pool
	prefix   string
)

func Parse(endpoints []string, path string) {

	conf := common.ConfParse(endpoints, path)
	prefix = conf.Prefix

	// 初始化beanstalk
	beanPool = conn.InitBeanstalk(conf.Beanstalkd.Addr, 50, 50, 100)
	// 初始化td
	td = conn.InitTD(conf.Td.Addr, conf.Td.MaxIdleConn, conf.Td.MaxOpenConn)
	common.InitTD(td)

	tdTask()
}

// 批量红利派发
func tdTask() {

	common.Log("sms", "短信自动过期脚本开始")

	// 初始化红利批量发放任务队列协程池
	tdPool, _ := ants.NewPoolWithFunc(10, func(payload interface{}) {

		if fn, ok := payload.(common.BeansFnParam); ok {
			tdHandle(fn.M)
			// 删除job
			_ = fn.Conn.Delete(fn.ID)
		}
	})

	topic := fmt.Sprintf("%s_sms", prefix)
	fmt.Printf("topic : %s\n", topic)
	attr := common.BeansWatcherAttr{
		TubeName:       topic,
		ReserveTimeOut: 2 * time.Minute,
		Pool:           tdPool,
	}

	common.BeanstalkWatcher(beanPool, attr)
}

//短信自动过期
func tdHandle(m map[string]interface{}) {
	fmt.Printf("bean data %#v", m)
}

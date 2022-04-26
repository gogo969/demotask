package message

import (
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/jmoiron/sqlx"
	"github.com/olivere/elastic/v7"
	"github.com/panjf2000/ants/v2"
	cpool "github.com/silenceper/pool"
	"strings"
	"task/contrib/conn"
	"task/modules/common"
	"time"
)

var (
	db       *sqlx.DB
	zlog     *fluent.Fluent
	esCli    *elastic.Client
	beanPool cpool.Pool
	dialect  = g.Dialect("mysql")
	esPrefix string
)

func Parse(endpoints []string, path string) {

	conf := common.ConfParse(endpoints, path)

	esPrefix = conf.EsPrefix
	// 初始化db
	db = conn.InitDB(conf.Db.Master.Addr, conf.Db.Master.MaxIdleConn, conf.Db.Master.MaxIdleConn)
	// 初始化beanstalk
	beanPool = conn.InitBeanstalk(conf.Beanstalkd.Addr, 50, 50, 100)
	// 初始化es
	esCli = conn.InitES(conf.Es.Host, conf.Es.Username, conf.Es.Password)

	batchMessageTask()
}

// 场馆转账订单确认
func batchMessageTask() {

	// 初始化场馆转账订单任务队列协程池
	messagePool, _ := ants.NewPoolWithFunc(500, func(payload interface{}) {

		if fn, ok := payload.(common.BeansFnParam); ok {
			// 场馆转账订单确认
			messageHandle(fn.M)
			// 删除job
			_ = fn.Conn.Delete(fn.ID)
		}
	})

	attr := common.BeansWatcherAttr{
		TubeName:       "message",
		ReserveTimeOut: 2 * time.Minute,
		Pool:           messagePool,
	}

	// 场馆转账订单确认队列
	common.BeanstalkWatcher(beanPool, attr)
}

func messageHandle(param map[string]interface{}) {

	//1 发送站内信 2 删除站内信
	flag, ok := param["flag"].(string)
	if !ok {
		common.Log("message", "messageHandle flag param null : %v \n", param)
		return
	}

	switch flag {
	case "1":
		sendHandle(param)
	case "2":
		deleteHandle(param)
	}
}

func deleteHandle(param map[string]interface{}) {

}

func sendHandle(param map[string]interface{}) {

	msgID, ok := param["msg_id"].(string)
	if !ok {
		common.Log("message", "sendHandle msgID param null : %v \n", param)
		return
	}
	//标题
	title, ok := param["title"].(string)
	if !ok {
		common.Log("message", "sendHandle title param null : %v \n", param)
		return
	}
	//副标题
	subTitle, ok := param["sub_title"].(string)
	if !ok {
		common.Log("message", "sendHandle sub_title param null : %v \n", param)
		return
	}
	//内容
	content, ok := param["content"].(string)
	if !ok {
		common.Log("message", "sendHandle content param null : %v \n", param)
		return
	}
	//0不置顶 1置顶
	isTop, ok := param["is_top"].(string)
	if !ok {
		common.Log("message", "sendHandle is_top param null : %v \n", param)
		return
	}
	//0不推送 1推送
	isPush, ok := param["is_push"].(string)
	if !ok {
		common.Log("message", "sendHandle is_push param null : %v \n", param)
		return
	}
	//0非vip站内信 1vip站内信
	isVip, ok := param["is_vip"].(string)
	if !ok {
		common.Log("message", "messageHandle is_vip param null : %v \n", param)
		return
	}
	//1站内消息 2活动消息
	ty, ok := param["ty"].(string)
	if !ok {
		common.Log("message", "sendHandle ty param null : %v \n", param)
		return
	}
	//发送人名
	sendName, ok := param["send_name"].(string)
	if !ok {
		common.Log("message", "sendHandle send_name param null : %v \n", param)
		return
	}
	//商户前缀
	prefix, ok := param["prefix"].(string)
	if !ok {
		common.Log("message", "sendHandle prefix param null : %v \n", param)
		return
	}

	switch isVip {
	case "0": //站内消息
		//会员名
		usernames, ok := param["usernames"].(string)
		if !ok {
			common.Log("message", "sendHandle level param null : %v \n", param)
			return
		}

		names := strings.Split(usernames, ",")
		count := len(names)
		p := count / 100
		l := count % 100
		// 分页发送
		for j := 0; j < p; j++ {
			offset := j * 100
			sendMessage(msgID, title, subTitle, content, isPush, isTop, ty, sendName, prefix, names[offset:offset+100])
		}
		// 最后一页
		if l > 0 {
			sendMessage(msgID, title, subTitle, content, isPush, isTop, ty, sendName, prefix, names[p*100:])
		}
	case "1": //vip站内信
		//会员等级
		level, ok := param["level"].(string)
		if !ok {
			common.Log("message", "sendHandle level param null : %v \n", param)
			return
		}

		lvs := strings.Split(level, ",")
		for _, v := range lvs {
			sendLevelMessage(msgID, title, subTitle, content, isPush, isTop, ty, sendName, prefix, v)
		}
	}
}

func sendLevelMessage(msgID, title, subTitle, content, isPush, isTop, ty, sendName, prefix, level string) {

	ex := g.Ex{
		"level": level,
	}
	count, err := common.MembersCount(db, ex)
	if err != nil {
		common.Log("message", "error : %v", err)
		return
	}

	fmt.Printf("count : %d\n", count)

	if count == 0 {
		return
	}

	p := count / 100
	l := count % 100
	if l > 0 {
		p += 1
	}

	for j := 1; j <= p; j++ {
		ns, err := common.MembersPageNames(db, j, 100, ex)
		if err != nil {
			common.Log("upgrade", "error : %v", err)
			return
		}

		sendMessage(msgID, title, subTitle, content, isPush, isTop, ty, sendName, prefix, ns)
	}
}

func sendMessage(msgID, title, subTitle, content, isPush, isTop, ty, sendName, prefix string, names []string) {

}

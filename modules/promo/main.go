package promo

import (
	"context"
	g "github.com/doug-martin/goqu/v9"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/olivere/elastic/v7"
	"github.com/panjf2000/ants/v2"
	cpool "github.com/silenceper/pool"
	"task/contrib/conn"
	"task/contrib/helper"
	"task/modules/common"
	"time"
)

// 活动流水计算脚本
var (
	db         *sqlx.DB
	cli        *redis.Client
	beanPool   cpool.Pool
	prefix     string
	lang       string
	loc        *time.Location
	es         *elastic.Client
	reportEs   *elastic.Client
	ctx        = context.Background()
	dialect    = g.Dialect("mysql")
	colPromo   = helper.EnumFields(Promo{})
	field      = []interface{}{"period", "display_at", "start_at", "hide_at", "end_at"}
	esPrefix   string
	pullPrefix string
	devices    = map[string]string{
		"web":      "24",
		"apph5":    "25,26,27",
		"sport":    "28,29",
		"agent":    "30,31",
		"carousel": "24,25,26,27,28,29,30,31",
	}
)

var (
	promoTyDividend    = "0"
	promoTyStateChange = "1"
	promoTyIncite      = "4"
)

func Parse(endpoints []string, path, flag string) {

	conf := common.ConfParse(endpoints, path)
	// 获取语言
	lang = conf.Lang
	if lang == "cn" {
		loc, _ = time.LoadLocation("Asia/Shanghai")
	} else if lang == "vn" || lang == "th" {
		loc, _ = time.LoadLocation("Asia/Bangkok")
	}

	prefix = conf.Prefix
	esPrefix = conf.EsPrefix
	pullPrefix = conf.PullPrefix

	// 初始化redis
	cli = conn.InitRedisSentinel(conf.Redis.Addr, conf.Redis.Password, conf.Redis.Sentinel, conf.Redis.Db)
	// 初始化db
	db = conn.InitDB(conf.Db.Master.Addr, conf.Db.Master.MaxIdleConn, conf.Db.Master.MaxIdleConn)

	// 初始化es
	es = conn.InitES(conf.Es.Host, conf.Es.Username, conf.Es.Password)
	// 初始化es
	reportEs = conn.InitES(conf.Es.Host, conf.Es.Username, conf.Es.Password)
	// 初始化beanstalk
	beanPool = conn.InitBeanstalk(conf.Beanstalkd.Addr, 50, 50, 100)

	if flag == "signday" {

		tm := time.Now()
		//计算当天数据
		signReport(tm.Unix())

		h := tm.Hour()
		//如果当前时间1点以前需计算前一天数据
		if h == 0 {
			tm = tm.AddDate(0, 0, -1)
			signReport(tm.Unix())
		}

		return
	}

	if flag == "signweek" {

		tm := time.Now()
		d := tm.Weekday()
		//周一计算上周签到奖金
		if d == time.Monday {
			signHandoutAward(tm.Unix())
		}

		return
	}

	if flag == "lastsignload" {

		tm := time.Now()
		SignLoadRecord(tm.Unix())

		return
	}

	promoTask()
}

// 批量红利派发
func promoTask() {

	common.Log("promo", "会员场馆流水服务开始")

	// 初始化红利批量发放任务队列协程池
	promoPool, _ := ants.NewPoolWithFunc(10, func(payload interface{}) {

		if fn, ok := payload.(common.BeansFnParam); ok {
			promoHandle(fn.M)
			// 删除job
			_ = fn.Conn.Delete(fn.ID)
		}
	})

	attr := common.BeansWatcherAttr{
		TubeName:       "promo",
		ReserveTimeOut: 2 * time.Minute,
		Pool:           promoPool,
	}

	// 场馆转账订单确认队列
	common.BeanstalkWatcher(beanPool, attr)
}

func promoHandle(m map[string]interface{}) {

	if m == nil {
		return
	}

	beanTy, ok := m["bean_ty"].(string)
	if !ok {
		return
	}

	switch beanTy {
	case promoTyDividend:
		promoDividend(m)
	case promoTyStateChange:
		promoStateChange(m)
	case promoTyIncite:
		depositSendDividendPromo(m)
	}

	return
}

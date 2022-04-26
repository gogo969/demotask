package message

import (
	g "github.com/doug-martin/goqu/v9"
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/olivere/elastic/v7"
	"github.com/panjf2000/ants/v2"
	cpool "github.com/silenceper/pool"
	"task/contrib/conn"
	"task/modules/common"
	"time"
)

var (
	db       *sqlx.DB
	cli      *redis.Client
	zlog     *fluent.Fluent
	esCli    *elastic.Client
	beanPool cpool.Pool
	dialect       = g.Dialect("mysql")
	pageSize uint = 100
	esPrefix string
)

type member struct {
	ID       string `json:"uid" db:"uid"`
	Username string `json:"username" db:"username"`
}

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

}

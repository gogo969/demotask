package message

import (
	g "github.com/doug-martin/goqu/v9"
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/olivere/elastic/v7"
	"github.com/panjf2000/ants/v2"
	cpool "github.com/silenceper/pool"
	"strings"
	"task/contrib/conn"
	"task/message"
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

	// 初始化fluent bit
	zlog = conn.InitFluentd(conf.Zlog.Host, conf.Zlog.Port)
	// 初始化db
	db = conn.InitDB(conf.Db.Master.Addr, conf.Db.Master.MaxIdleConn, conf.Db.Master.MaxIdleConn)
	//初始化merchant redis
	cli = conn.InitRedisSentinel(conf.Redis.Addr, conf.Redis.Password, conf.Redis.Sentinel, conf.Redis.Db)
	// 初始化beanstalk
	beanPool = conn.InitBeanstalk(conf.Beanstalkd.Addr, 50, 50, 100)
	// 初始化es
	esCli = conn.InitES(conf.Es.Host, conf.Es.Username, conf.Es.Password)

	batchMessageTask()
}

// 场馆转账订单确认
func batchMessageTask() {

	// 初始化场馆转账订单任务队列协程池
	confirmPool, _ := ants.NewPoolWithFunc(500, func(payload interface{}) {

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
		Pool:           confirmPool,
	}

	// 场馆转账订单确认队列
	common.BeanstalkWatcher(beanPool, attr)
}

// map 参数
// ty : 1公告自动启停 2 站内信全部会员 3 站内信指定会员 4 更新公告 5 删除公告 6 更新站内信 7 删除站内信
//	ty = 1 : param = posts_id
//	ty = 2 : param = letter_id
//	ty = 3 : param = letter_id,username,uid
//  ty = 4 : param = post_id,state
//  ty = 5 : param = post_id
//  ty = 6 : param = letter_id,state
//  ty = 7 : param = letter_id
func messageHandle(param map[string]interface{}) {

	ty, ok := param["ty"]
	if !ok {
		common.Log("message", "%s ty error", param["ty"])
		return
	}

	switch ty.(string) {
	case message.PostStateManage: //公告自动启停

		postsStateChange(param)

	case message.LetterInsertOperating: //站内信全部会员

		letterInsertAll(param["letter_id"].(string))

	case message.LetterMembersInsertOperating: // 站内信指定会员

		letterInsertMembers(param)

	case message.PostUpdateOperating:

		p := map[string]string{}

		if state, ok := param["state"]; ok {
			p["state"] = state.(string)
		}

		if title, ok := param["title"]; ok {
			p["title"] = title.(string)
		}

		if content, ok := param["content"]; ok {
			p["content"] = content.(string)
		}

		_, err := message.EsPostsUpdate(esCli, esPrefix, param["posts_id"].(string), p)
		if err != nil {
			common.Log("message", "PostUpdateOperating err: %v", err)
		}
	case message.PostDeleteOperating:

		_, err := message.EsPostsDel(esCli, esPrefix, param["posts_id"].(string))
		if err != nil {
			common.Log("message", "PostDeleteOperating err: %v", err)
		}
	case message.LetterUpdateOperating:

		p := map[string]string{}

		if state, ok := param["state"]; ok {
			p["state"] = state.(string)
		}

		if title, ok := param["title"]; ok {
			p["title"] = title.(string)
		}

		if content, ok := param["content"]; ok {
			p["content"] = content.(string)
		}
		_, err := message.EsLetterUpdate(esCli, esPrefix, param["letter_id"].(string), p)
		if err != nil {
			common.Log("message", "LetterUpdateOperating err: %v", err)
		}
	case message.LetterDeleteOperating:
		_, err := message.EsLetterDel(esCli, esPrefix, param["letter_id"].(string))
		if err != nil {
			common.Log("message", "LetterUpdateOperating err: %v", err)
		}
	case message.SendNoticeOperating:
		//_ = message.SendNotice(esCli, cli, message.SceneRegister, param["uid"].(string), param["username"].(string), time.Now().Unix(), param)
		_, err := message.TaskSendNotice(esCli, cli, esPrefix, param)
		common.Log("message", "send notice: %v", err)
	default:
		common.Log("message", "ty not used: %v", ty)
	}
}

//公告自动启停
func postsStateChange(param map[string]interface{}) {

	var (
		now   = time.Now().Unix()
		state = param["state"].(string)
		id    = param["id"].(string)
	)

	if state == "4" {

		ex := g.Ex{
			"id": id,
		}

		record := g.Record{
			"state":         state,
			"review_at":     now,
			"review_uid":    0,
			"review_name":   "",
			"review_remark": "",
		}

		_, err := message.PostsUpdate(db, ex, record)
		if err != nil {
			common.Log("message", "posts state change: %v", err)
			return
		}
	}

	_, _ = message.LoadToCache(db, cli)
}

// 指定会员发送站内信
func letterInsertMembers(param map[string]interface{}) {

	letter, _, err := message.LetterFind(db, param["letter_id"].(string))
	if err != nil {
		common.Log("message", "err : %v", err)
		return
	}

	users := []member{
		{ID: param["uid"].(string), Username: param["username"].(string)},
	}

	var key string
	if letter.Ty == message.LetterTyNotify {
		key = message.UnreadKeyNotice
	}

	if letter.Ty == message.LetterTyPromo {
		key = message.UnreadKeyPromo
	}

	_, err = message.UnreadNumsIncr(cli, membersToMap(users), key, 1)
	if err != nil {
		common.Log("message", "err : %v", err)
	}

	_, err = message.EsLetterInsert(esCli, esPrefix, letter, membersToMap(users))
	if err != nil {

		_, readErr := message.UnreadNumsIncr(cli, membersToMap(users), key, -1)
		if readErr != nil {
			common.Log("message", "err : %v", readErr)
		}

		common.Log("message", "err : %v", err)
		return
	}
}

// 全部会员 发布站内信
func letterInsertAll(letterId string) {

	letter, _, err := message.LetterFind(db, letterId)
	if err != nil {
		common.Log("message", "err : %v", err)
		return
	}

	// 判断是否有vip等级限制
	var levels []string
	if letter.Level != "0" {
		levels = strings.Split(letter.Level, ",")
	}

	total := userTotal(levels)
	totalPage := (total + int(pageSize) - 1) / int(pageSize)

	for i := 0; i < totalPage; i++ {
		// 获取数据
		users, err := getMembers(uint(i+1), levels)
		if err != nil {
			common.Log("message", "err : %v", err)
			return
		}

		var key string
		if letter.Ty == message.LetterTyNotify {
			key = message.UnreadKeyNotice
		}

		if letter.Ty == message.LetterTyPromo {
			key = message.UnreadKeyPromo
		}

		_, err = message.UnreadNumsIncr(cli, membersToMap(users), key, 1)
		if err != nil {
			common.Log("message", "err : %v", err)
		}

		_, err = message.EsLetterInsert(esCli, esPrefix, letter, membersToMap(users))
		if err != nil {

			_, readErr := message.UnreadNumsIncr(cli, membersToMap(users), key, -1)
			if readErr != nil {
				common.Log("message", "err : %v", readErr)
			}

			common.Log("message", "err : %v", err)
			return
		}
	}
}

// 发布系统公告
func postInsert(postId string) {

	posts, _, err := message.PostsFind(db, postId)
	if err != nil {
		common.Log("message", "err : %v", err)
		return
	}

	_, err = message.EsPostsInsertSpecial(esCli, esPrefix, posts)
	if err != nil {
		common.Log("message", "err : %s", err)
	}
}

// 用户信息 转化为map {username:uid}
func membersToMap(users []member) map[string]string {

	data := make(map[string]string)
	for _, v := range users {
		data[v.Username] = v.ID
	}

	return data
}

// 获取可发布的用户信息 每次限量100 人
func getMembers(page uint, levels []string) ([]member, error) {

	var data []member
	offset := (page - 1) * pageSize

	//  用户状态为关闭，30天内有登录
	ex := g.Ex{
		"state":         1,
		"last_login_at": g.Op{"gt": time.Now().Add(-30 * 24 * time.Hour).Unix()},
	}
	if levels != nil {
		ex["level"] = levels
	}
	query, _, _ := dialect.From("tbl_members").Select("uid", "username").Where(ex).Offset(offset).Limit(pageSize).ToSQL()
	err := db.Select(&data, query)
	if err != nil {
		common.Log("message", "err : %v", err)
		return nil, err
	}

	return data, nil
}

// 获取 可批量发送的总人数
func userTotal(levels []string) int {

	total := 0
	//  用户状态为关闭，30天内有登录
	ex := g.Ex{
		"state":         1,
		"last_login_at": g.Op{"gt": time.Now().Add(-30 * 24 * time.Hour).Unix()},
	}
	if levels != nil {
		ex["level"] = levels
	}
	query, _, _ := dialect.From("tbl_members").Where(ex).Select(g.COUNT("uid")).ToSQL()
	err := db.Get(&total, query)
	if err != nil {
		common.Log("message", "err : %v", err)
	}

	return total
}

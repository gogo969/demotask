package common

import (
	"fmt"
	"github.com/beanstalkd/go-beanstalk"
	"github.com/panjf2000/ants/v2"
	cpool "github.com/silenceper/pool"
	"github.com/valyala/fasthttp"
	"strings"
	"time"
)

type BeansFnParam struct {
	Conn *beanstalk.Conn
	ID   uint64
	M    map[string]interface{}
}

type BeansProducerAttr struct {
	Tube    string                 //tube 名称
	Message map[string]interface{} //job body
	Delay   time.Duration          //延迟ready的秒数
	TTR     time.Duration          //允许worker执行的最大秒数
	PRI     uint32                 //优先级
}

type BeansWatcherAttr struct {
	TubeName       string
	ReserveTimeOut time.Duration
	Pool           *ants.PoolWithFunc
}

// beanstalk  投递消息
func BeansAddTask(pool cpool.Pool, data BeansProducerAttr) error {

	m := &fasthttp.Args{}
	for k, v := range data.Message {
		m.Set(k, v.(string))
	}

	c, err := pool.Get()
	if err != nil {
		return err
	}

	if conn, ok := c.(*beanstalk.Conn); ok {

		tube := &beanstalk.Tube{Conn: conn, Name: data.Tube}
		_, err = tube.Put(m.QueryString(), data.PRI, data.Delay, data.TTR)
		if err != nil {
			pool.Put(c)
			return err
		}
	}

	//将连接放回连接池中
	return pool.Put(c)
}

// beanstalk 通用消费方法
func BeanstalkWatcher(beansPool cpool.Pool, attr BeansWatcherAttr) {

	//循环监听
	for {
		v, err := beansPool.Get()
		if err != nil {
			fmt.Println("Beanstalk", 0, err.Error())
			_ = beansPool.Put(v)
			continue
		}

		if conn, ok := v.(*beanstalk.Conn); ok {

			c := beanstalk.NewTubeSet(conn, attr.TubeName)
			id, msg, err := c.Reserve(attr.ReserveTimeOut)
			//无job时会返回timeout ,不打印日志，不处理
			if err != nil {

				if strings.Contains(err.Error(), "deadline soon") {
					//超时,续期
					_ = conn.Touch(id)
				} else if !strings.Contains(err.Error(), "timeout") {
					fmt.Printf("tube: %s reserve error: %s\n", attr.TubeName, err.Error())
				}

				_ = beansPool.Put(v)
				continue
			}

			message := string(msg)
			//记录日志
			fmt.Printf("tube: %s msg: %s running\n", attr.TubeName, message)
			//避免过期，增加续约机制
			_ = conn.Touch(id)

			// 获取参数
			param := map[string]interface{}{}
			m := &fasthttp.Args{}
			m.Parse(message)
			if m.Len() == 0 {
				fmt.Printf("tube: %s msg: %s parse error, deleted!\n", attr.TubeName, message)
				_ = conn.Delete(id)
				_ = beansPool.Put(v)
				continue
			}

			m.VisitAll(func(key, value []byte) {
				param[string(key)] = string(value)
			})

			fn := BeansFnParam{
				Conn: conn,
				ID:   id,
				M:    param,
			}

			// 注单自动确认
			if err := attr.Pool.Invoke(fn); err != nil {
				fmt.Printf("invoke error: %s\n", err.Error())
				_ = beansPool.Put(v)
				continue
			}
		}

		_ = beansPool.Put(v)
	}
}

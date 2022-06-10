package trc20

import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/panjf2000/ants/v2"
)

func batchCgBetTask() {

	// 初始化投注记录任务队列协程池
	confirmPool, _ := ants.NewPoolWithFunc(500, func(bet interface{}) {

		if payload, ok := bet.(string); ok {
			// 站内信处理
			messageHandle(payload)
		}
	})

	topic := prefix + "_trc20"
	merchantConsumer.Subscribe(topic, consumer.MessageSelector{}, func(ctx context.Context,
		msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for i := range msgs {

			fmt.Println("trc20 收到消息：", string(msgs[i].Body))
			// 注单自动确认
			if err := confirmPool.Invoke(string(msgs[i].Body)); err != nil {
				fmt.Printf("invoke error: %s\n", err.Error())
				continue
			}
		}

		return consumer.ConsumeSuccess, nil
	})
}

// trc20查单程序
func messageHandle(payload string) {

}

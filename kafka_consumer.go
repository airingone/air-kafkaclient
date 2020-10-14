package air_kafkaclient

import (
	"context"
	"github.com/airingone/log"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"os"
	"os/signal"
	"syscall"
)

//kafka consumer
type KafkaConsumer struct {
	c         *kafka.Consumer //client
	addrs     string          //kafka broker addrs
	group     string          //group
	topics    []string
	timeoutMs uint32 //time out ms
	stop      bool
	handle    KafkaConsumerHandler
}

//消费处理函数
type KafkaConsumerHandler func(context.Context, *kafka.Message) error

//创建消费，会启动消费协程poll轮询消息，有消息后则会创建工作协程处理信息
//addrs: kafka broker地址
//group: 消费group
//topic: 主题
//timeoutMs: pool超时
//handle: 消息处理函数
func NewKafkaConsumer(addrs string, group string, topics []string, timeoutMs uint32, handle KafkaConsumerHandler) (*KafkaConsumer, error) {
	config := &kafka.ConfigMap{
		"bootstrap.servers":  addrs,
		"group.id":           group,
		"session.timeout.ms": 6000, //连接session，超过这个时间无消息或心跳则表示已断开了连接
		//"heartbeat.interval.ms": 3000, //心跳时间，默认3s
		"max.poll.interval.ms": int(timeoutMs), //使用消费者组管理时轮询调用之间的最大延迟，默认300000
		"auto.offset.reset":    "earliest",
	}
	c, err := kafka.NewConsumer(config)
	if err != nil {
		return nil, err
	}

	err = c.SubscribeTopics(topics, nil)
	if err != nil {
		return nil, err
	}

	cli := &KafkaConsumer{
		c:         c,
		addrs:     addrs,
		group:     group,
		topics:    topics,
		timeoutMs: timeoutMs,
		stop:      false,
		handle:    handle,
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.PanicTrack()
			}
		}()

		cli.consumer()
	}()

	return cli, nil
}

//消息协程
func (cli *KafkaConsumer) consumer() {
	defer func() {
		if cli.c != nil {
			cli.c.Close()
		}
	}()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	for cli.stop == false {
		select {
		case sig := <-sigchan:
			log.Error("[KafKa] Signal stop, %+v", sig)
			cli.stop = true
		default:
			ev := cli.c.Poll(500)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				go func() {
					defer func() {
						if r := recover(); r != nil {
							log.PanicTrack()
						}
					}()

					_ = cli.handle(context.Background(), e)
				}()
			case kafka.Error:
				log.Error("[KafKa] consumer msg err, %+v", e)
				if e.Code() == kafka.ErrAllBrokersDown {
					log.Error("[KafKa] all brokers down err, stop")
					cli.stop = true
				}
			default:
				log.Error("[KafKa] consumer msg ignored, %+v", e)
			}
		}
	}
}

//stop
func (cli *KafkaConsumer) Stop() {
	cli.stop = true
}

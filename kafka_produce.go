package air_kafkaclient

import (
	"errors"
	"github.com/airingone/log"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"time"
)

//kafka produce
type KafkaProduce struct {
	p            *kafka.Producer  //client
	addrs        string           //kafka broker addrs
	topic        string           //topic
	timeoutMs    uint32           //time out ms
	deliveryChan chan kafka.Event //delivery chan
}

//创建kafka生产者对象
//addrs: kafka broker地址
//topic: 主题
func NewKafkaProduce(addrs string, topic string, timeoutMs uint32) (*KafkaProduce, error) {
	config := &kafka.ConfigMap{
		"bootstrap.servers": addrs,
	}
	p, err := kafka.NewProducer(config)
	if err != nil {
		return nil, err
	}

	cli := &KafkaProduce{
		p:            p,
		addrs:        addrs,
		topic:        topic,
		timeoutMs:    timeoutMs,
		deliveryChan: make(chan kafka.Event),
	}

	return cli, nil
}

//发送消息到kafka
//data: 数据
func (cli *KafkaProduce) Produce(data []byte) error {
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &cli.topic, Partition: kafka.PartitionAny},
		Value:          data,
		Timestamp:      time.Now(),
	}
	err := cli.p.Produce(msg, cli.deliveryChan)
	if err != nil {
		return err
	}

	select {
	case e := <-cli.deliveryChan:
		m := e.(*kafka.Message)
		if m.TopicPartition.Error != nil {
			log.Error("[KafKa] produced message err, %+v", m.TopicPartition)
		} else {
			log.Info("[KafKa] produced message succ, %+v", m.TopicPartition)
		}
	case <-time.After(time.Duration(cli.timeoutMs) * time.Millisecond):
		return errors.New("delivery timeout")
	}

	return nil
}

//close
func (cli *KafkaProduce) Close() {
	cli.p.Close()
}

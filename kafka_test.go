package air_kafkaclient

import (
	"context"
	"fmt"
	"github.com/airingone/config"
	"github.com/airingone/log"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"testing"
	"time"
)

//kafka produce测试
func TestNewKafkaProduce(t *testing.T) {
	config.InitConfig()                     //配置文件初始化
	log.InitLog(config.GetLogConfig("log")) //日志初始化

	kafkaConfig := config.GetKafKaProductConfig("kafka_product")
	cli, err := NewKafkaProduce(kafkaConfig.Addr, kafkaConfig.Topic, kafkaConfig.TimeOutMs)
	if err != nil {
		log.Error("NewKafkaProduce err, %+v", err)
		return
	}

	for i := 0; i < 10; i++ {
		data := fmt.Sprintf("kafka test data%d", i)
		_ = cli.Produce([]byte(data))
	}

	cli.Close() //进程退出时关闭
}

//消费处理函数
func handleMsg(ctx context.Context, msg *kafka.Message) error {
	log.Error("consumer msg succ, msg: %+v", msg)
	log.Error("consumer msg succ, value: %s", string(msg.Value))

	return nil
}

//kafka consumer测试
func TestNewKafkaConsumer(t *testing.T) {
	config.InitConfig()                     //配置文件初始化
	log.InitLog(config.GetLogConfig("log")) //日志初始化

	kafkaConfig := config.GetKafKaConsumerConfig("kafka_consumer")
	cli, err := NewKafkaConsumer(kafkaConfig.Addr, kafkaConfig.Group,
		kafkaConfig.Topics, kafkaConfig.TimeOutMs, handleMsg) //创建消费协程，收到消息后会新创建一个协程执行handleMsg函数
	if err != nil {
		log.Error("NewKafkaConsumer err, %+v", err)
		return
	}

	time.Sleep(time.Second * 60)
	cli.Stop()
}

/*
brew cask install java6
brew install kafka
brew install zookeeper

vi /usr/local/etc/kafka/server.properties
#     listeners = PLAINTEXT://your.host.name:9092
listeners=PLAINTEXT://localhost:9092

如果想以服务的方式启动，那么可以:
$ brew services start zookeeper
$ brew services start kafka
如果只是临时启动，可以:
$ zkServer start
$ kafka-server-start /usr/local/etc/kafka/server.properties

创建Topic
$ kafka-topics --create --zookeeper localhost:2181 --replication-factor 1 --partitions 1 --topic topic01
产生消息
$ kafka-console-producer --broker-list localhost:9092 --topic topic01
>HELLO Kafka
消费
简单方式:
$ kafka-console-consumer --bootstrap-server localhost:9092 --topic topic01 --from-beginning
如果使用消费组:
kafka-console-consumer --bootstrap-server localhost:9092 --topic topic01 --group group01 --from-beginning
*/

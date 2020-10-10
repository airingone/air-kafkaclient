# kafka client组件
## 1.组件描述
kafka client组件封装了kafka product与kafka consumer。
## 2.如何使用
### 2.1 product client
```
import (
    "github.com/airingone/config"
    "github.com/airingone/log"
    "github.com/confluentinc/confluent-kafka-go/kafka"
    kafkaclient "github.com/airingone/air-kafkaclient"
)

func main() {
    config.InitConfig()                     //进程启动时调用一次初始化配置文件，配置文件名为config.yml，目录路径为../conf/或./
    log.InitLog(config.GetLogConfig("log")) //进程启动时调用一次初始化日志
    
    kafkaConfig := config.GetKafKaProductConfig("kafka_product")
	cli, err := kafkaclient.NewKafkaProduce(kafkaConfig.Addr, kafkaConfig.Topic, kafkaConfig.TimeOutMs)
	if err != nil {
		log.Error("NewKafkaProduce err, %+v", err)
		return
	}

	for i:=0; i<10; i++ {
		data := fmt.Sprintf("kafka test data%d", i)
		_ = cli.Produce([]byte(data))
	}

	cli.Close() //进程退出时关闭
}
```
### 2.2 consumer client
```
import (
    "github.com/airingone/config"
    "github.com/airingone/log"
    "github.com/confluentinc/confluent-kafka-go/kafka"
    kafkaclient "github.com/airingone/air-kafkaclient"
)

//消费处理函数
func handleMsg(ctx context.Context, msg *kafka.Message) error {
	log.Error("consumer msg succ, msg: %+v", msg)
	log.Error("consumer msg succ, value: %s", string(msg.Value))

	return nil
}

func main() {
    config.InitConfig()                     //进程启动时调用一次初始化配置文件，配置文件名为config.yml，目录路径为../conf/或./
    log.InitLog(config.GetLogConfig("log")) //进程启动时调用一次初始化日志
    
	kafkaConfig := config.GetKafKaConsumerConfig("kafka_consumer")
	cli, err := kafkaclient.NewKafkaConsumer(kafkaConfig.Addr, kafkaConfig.Group,
		kafkaConfig.Topics, kafkaConfig.TimeOutMs, handleMsg) //创建消费协程，收到消息后会新创建一个协程执行handleMsg函数
	if err != nil {
		log.Error("NewKafkaConsumer err, %+v", err)
		return
	}

    for {
	    time.Sleep(time.Second * 60)
    }
	cli.Stop()    
}
```
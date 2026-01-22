package k_kafka

//
//func NewCTGMQConsumer() {
//	var kafkaConsumer *kafka.Consumer
//	kafkaConfig := kafka.NewConsumerConfig(
//		conf.GetString("Recv.Kafka.Broker"),
//		conf.GetString("Recv.Kafka.Topic"),
//		conf.GetBool("Recv.Kafka.Oldest"),
//	)
//	kctx.Infof("kafak broker: %s, topic: %s", kafkaConfig.Brokers, kafkaConfig.Topics)
//	v := conf.GetString("Recv.Kafka.Version")
//	if v == "" {
//		kafkaConfig.Version = sarama.DefaultVersion.String()
//	} else {
//		kafkaConfig.Version = v
//	}
//	if conf.GetBool("Recv.Kafka.Sasl.Enabled") {
//		kafkaConfig.SaslEnable = true
//		kafkaConfig.UserName = conf.GetString("Recv.Kafka.Sasl.Username")
//		kafkaConfig.Password = conf.GetString("Recv.Kafka.Sasl.Password")
//		kafkaConfig.Algorithm = conf.GetString("Recv.Kafka.Sasl.Algorithm")
//	}
//
//	var err error
//	kafkaConsumer, err = kafka.NewConsumer(kafkaConfig)
//	if err != nil {
//		kctx.Errorf("Failed to create Kafka consumer,waiting s and retry... ... \n err: %v", err)
//		time.Sleep(time.Second * kafkaTimeout)
//		CollectKafka(model)
//		return
//	}
//}

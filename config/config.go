package config

import (
	"github.com/jinzhu/configor"
)

type Config struct {
	AppConfig     AppConfig `env:"APP_CONFIG"`
	DBConfig      DBConfig
	DataDogConfig DataDogConfig
	TheTVDBConfig TheTVDBConfig `env:"THETVDB"`
	KafkaConfig   KafkaConfig
	NatsConfig    NatsConfig
	ProducerType  string `default:"kafka" env:"PRODUCER_TYPE"`
}

type AppConfig struct {
	APPName string `default:"scraper-api"`
	Port    int    `env:"PORT" default:"3000"`
	Version string `default:"x.x.x" env:"VERSION"`
}

type DBConfig struct {
	Host     string `default:"localhost" env:"DBHOST"`
	DataBase string `default:"weeb" env:"DBNAME"`
	User     string `default:"weeb" env:"DBUSERNAME"`
	Password string `required:"true" env:"DBPASSWORD" default:"mysecretpassword"`
	Port     uint   `default:"3306" env:"DBPORT"`
	SSLMode  string `default:"disable" env:"DBSSL"`
}

type DataDogConfig struct {
	DD_AGENT_HOST string `env:"DD_AGENT_HOST" default:"localhost"`
	DD_AGENT_PORT int    `env:"DD_AGENT_PORT" default:"8125"`
}

type TheTVDBConfig struct {
	APIKey string `default:"" env:"API_KEY"`
	APIPIN string `default:"" env:"API_PIN"`
}

type KafkaConfig struct {
	ConsumerGroupName string `default:"image-sync-group" env:"KAFKA_CONSUMER_GROUP_NAME"`
	BootstrapServers  string `default:"localhost:9092" env:"KAFKA_BOOTSTRAP_SERVERS"`
	Offset            string `default:"earliest" env:"KAFKA_OFFSET"`
	ProducerTopic     string `default:"image-sync" env:"KAFKA_PRODUCER_TOPIC"`
}

// NatsConfig is the producer half of KafkaConfig, for PRODUCER_TYPE=nats.
//
// Which transport is used is decided at startup rather than by a command name:
// this producer is constructed inside the GraphQL handler and inside the sync
// job, so there is no single process to name the way the standalone consumers
// have.
type NatsConfig struct {
	URL string `default:"nats://localhost:4222" env:"NATSURL"`

	// Empty StreamName: this subject is produced here rather than by Debezium,
	// so nothing else declares a stream over it and the driver creates one from
	// the subject.
	StreamName string `env:"NATSSTREAMNAME"`

	ProducerSubject string `default:"image-sync" env:"NATSPRODUCERSUBJECT"`
}

func LoadConfigOrPanic() Config {
	var config = Config{}
	configor.Load(&config, "config/config.dev.json")

	return config
}

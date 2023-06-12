package kafka

import "github.com/Shopify/sarama"

type Config struct {
	Address []string
	Config  sarama.Config
	Group   string
}

// func (cf *Config)Init

func (cf *Config) NewProducer() (sarama.SyncProducer, error) {
	producer, err := sarama.NewSyncProducer(cf.Address, &cf.Config)
	if err != nil {
		return nil, err
	}
	return producer, nil
}

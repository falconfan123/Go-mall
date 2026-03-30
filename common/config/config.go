package config

import "fmt"

type PostgresConfig struct {
	DataSource  string
	Conntimeout int
}

type RabbitMQConfig struct {
	Host  string
	Port  int
	User  string
	Pass  string
	VHost string
}
type ElasticSearchConfig struct {
	Addr string
}
type GorseConfig struct {
	GorseAddr   string
	GorseApikey string
}

func (r *RabbitMQConfig) Dns() string {
	return fmt.Sprintf(
		"amqp://%s:%s@%s:%d/%s",
		r.User,
		r.Pass,
		r.Host,
		r.Port,
		r.VHost,
	)
}

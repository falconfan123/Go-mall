package config

import (
	"fmt"
	"net/url"
	"strings"
)

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
	Addr      string
	IndexName string
}
type GorseConfig struct {
	GorseAddr   string
	GorseApikey string
}

func (r *RabbitMQConfig) Dns() string {
	vhost := strings.TrimSpace(r.VHost)
	if vhost == "" {
		vhost = "/"
	}

	escapedVHost := "%2F"
	if vhost != "/" {
		escapedVHost = url.PathEscape(strings.TrimPrefix(vhost, "/"))
	}

	return fmt.Sprintf(
		"amqp://%s:%s@%s:%d/%s",
		r.User,
		r.Pass,
		r.Host,
		r.Port,
		escapedVHost,
	)
}

package impl

import (
	proxymode "museum/config/proxy-mode"
)

type EnvConfig struct {
	EtcdHost    string `env:"ETCD_HOST,required"`
	EtcdBaseKey string `env:"ETCD_BASE_KEY" envDefault:"museum"`
	NatsHost    string `env:"NATS_HOST"`
	NatsBaseKey string `env:"NATS_BASE_KEY" envDefault:"museum"`
	DockerHost  string `env:"DOCKER_HOST" envDefault:"unix:///var/run/docker.sock"`
	Hostname    string `env:"HOSTNAME" envDefault:"localhost"`
	Port        string `env:"PORT" envDefault:"8080"`
	JaegerHost  string `env:"JAEGER_HOST"`
	Environment string `env:"ENVIRONMENT" envDefault:"development"`
	ProxyMode   string `env:"PROXY_MODE" envDefault:"swarm-ext"`
	DevProxyUrl string `env:"DEV_PROXY_URL" envDefault:"http://localhost:3000"`
}

func (e EnvConfig) GetEtcdHost() string {
	return e.EtcdHost
}

func (e EnvConfig) GetEtcdBaseKey() string {
	return e.EtcdBaseKey
}

func (e EnvConfig) GetNatsHost() string {
	return e.NatsHost
}

func (e EnvConfig) GetNatsBaseKey() string {
	return e.NatsBaseKey
}

func (e EnvConfig) GetDockerHost() string {
	return e.DockerHost
}

func (e EnvConfig) GetHostname() string {
	return e.Hostname
}

func (e EnvConfig) GetPort() string {
	return e.Port
}

func (e EnvConfig) GetJaegerHost() string {
	return e.JaegerHost
}

func (e EnvConfig) GetEnvironment() string {
	return e.Environment
}

func (e EnvConfig) GetProxyMode() proxymode.Mode {
	switch e.ProxyMode {
	case "swarm":
		return proxymode.ModeSwarm
	case "swarm-ext":
		return proxymode.ModeSwarmExt
	default:
		panic("invalid proxy mode" + e.ProxyMode)
	}
}

func (e EnvConfig) GetDevProxyUrl() string {
	return e.DevProxyUrl
}

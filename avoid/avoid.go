package avoid

import (
	"fmt"

	"io/ioutil"
	"time"

	log "github.com/sirupsen/logrus"
	"gitlab.com/mergetb/tech/stor"
	grpc "google.golang.org/grpc"
	"gopkg.in/yaml.v2"
)

// TODO: Add grpc tls options
func WithAvoid(endpoint string, f func(TunnelClient) error) error {
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to connect to avoid service: %v", err)
	}

	client := NewTunnelClient(conn)
	defer conn.Close()

	return f(client)
}

// https://pulwar.isi.edu/sabres/orchestrator/-/blob/main/pkg/config.go
type ServiceConfig struct {
	Address  string          `yaml:address",omitempty"` // Address and Port should be on VPN for traffic to go over VPN
	Port     int             `yaml:port",omitempty"`    // Address and Port should be on VPN for traffic to go over VPN
	TLS      *stor.TLSConfig `yaml:tls",omitempty"`
	Timeout  int             `yaml:timeout",omitempty"`
	Identity string          `yaml:identity",omitempty"`
}

// ServicesConfig encapsulates information for communicating with services.
type ServicesConfig struct {
	Avoid *ServiceConfig `yaml:",omitempty"`
}

// Endpoint returns the endpoint string of a service config.
func (s *ServiceConfig) Endpoint() string {
	return fmt.Sprintf("%s:%d", s.Address, s.Port)
}

func LoadConfig(configPath string) (*ServicesConfig, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Errorf("could not read configuration file %s", configPath)
		return nil, err
	}

	log.Infof("%s", data)

	cfg := &ServicesConfig{}
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		log.Errorf("could not parse configuration file")
		return nil, err
	}

	log.WithFields(log.Fields{
		"config": fmt.Sprintf("%+v", *cfg),
	}).Debug("config")

	return cfg, nil
}

// TODO: When we persist data
func SetAvoidSettings(config *ServicesConfig) (*stor.Config, error) {
	cfg := &stor.Config{}

	if config.Avoid != nil {
		cfg.Address = config.Avoid.Address
		cfg.Port = config.Avoid.Port
		cfg.TLS = config.Avoid.TLS
		cfg.Timeout = time.Duration(config.Avoid.Timeout) * time.Millisecond
	} else {
		return nil, fmt.Errorf("No Avoid config found.\n")
	}

	return cfg, nil
}

package avoid

import (
	"fmt"
	"sync/atomic"

	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"github.com/slackhq/nebula/config"
	"gitlab.com/mergetb/tech/stor"
	grpc "google.golang.org/grpc"
	"gopkg.in/yaml.v2"
)

// TODO: Make this all atomics
type Avoid struct {
	client   atomic.Bool
	manager  atomic.Bool
	config   string
	identity string
	primary  *Endpoint
	backups  []*Endpoint
	l        *log.Logger
}

func NewAvoidFromConfig(l *log.Logger, c *config.C) *Avoid {
	av := &Avoid{l: l}

	av.reload(c, true)
	c.RegisterReloadCallback(func(c *config.C) {
		av.reload(c, false)
	})

	return av
}

func (av *Avoid) reload(c *config.C, initial bool) {
	// TODO: initial value
	var fp string
	if !c.IsSet("avoid") {
		av = &Avoid{}
		return
	} else {
		fp = c.GetString("avoid.config", "")
		if fp == "" {
			av.l.Errorf("Avoid not configured, but required\n")
			return
		}
		av.config = fp

		cfg, err := LoadConfig(fp)
		if err != nil {
			av.l.WithField("err", err).Errorf("Failed to load avoid config file\n")
			return
		}

		if (cfg.Manager == nil && cfg.Client == nil) || (cfg.Manager != nil && cfg.Client != nil) {
			av.l.Errorf("Client or Manager (mutually exclusive) must be set\n")
			return
		}
		if cfg.Manager != nil {
			av.manager.Store(true)
			av.client.Store(false)
			if cfg.Manager.EP != nil {
				av.primary = cfg.Manager.EP
			} else {
				av.l.Errorf("Manager requires an endpoint\n")
				return
			}
		} else {
			if cfg.Client.Identity == "" {
				av.l.Errorf("No identity specified\n")
				return
			}
			av.identity = cfg.Client.Identity

			av.client.Store(true)
			av.manager.Store(false)
			if cfg.Client.EPS != nil {
				if len(cfg.Client.EPS) < 1 {
					av.l.Errorf("Client must have at least 1 endpoint\n")
					return
				}
				if len(cfg.Client.EPS) >= 1 {
					av.backups = cfg.Client.EPS[:len(cfg.Client.EPS)]
				}
				for _, v := range cfg.Client.EPS {
					if v.Primary {
						av.primary = v
						index := 0
						for i, vv := range av.backups {
							if v == vv {
								index = i
								break
							}
						}
						av.backups = append(av.backups[:index], av.backups[index+1:]...)
						break
					}
				}
				if av.primary == nil {
					if len(av.backups) > 1 {
						av.primary, av.backups = av.backups[0], av.backups[1:]
					} else {
						av.primary = av.backups[0]
						av.backups = nil
					}
				}
			}
		}
	}
}

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

type Endpoint struct {
	Address string          `yaml:address",omitempty"` // Address and Port should be on VPN for traffic to go over VPN
	Port    int             `yaml:port",omitempty"`    // Address and Port should be on VPN for traffic to go over VPN
	TLS     *stor.TLSConfig `yaml:tls",omitempty"`
	Timeout int             `yaml:timeout",omitempty"`
	Primary bool            `yaml:primary",omitempty"`
}

// https://pulwar.isi.edu/sabres/orchestrator/-/blob/main/pkg/config.go
type ClientServiceConfig struct {
	EPS      []*Endpoint `yaml:eps",omitempty"`
	Identity string      `yaml:identity",omitempty"`
}

type ManagerServiceConfig struct {
	EP *Endpoint `yaml:ep",omitempty"`
}

// ServicesConfig encapsulates information for communicating with services.
type ServicesConfig struct {
	Client  *ClientServiceConfig  `yaml:client",omitempty"`
	Manager *ManagerServiceConfig `yaml:manager",omitempty"`
}

// Endpoint returns the endpoint string of a service config.
func (ep *Endpoint) ToAddr() string {
	return fmt.Sprintf("%s:%d", ep.Address, ep.Port)
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

	if cfg.Client != nil {
		log.WithFields(log.Fields{
			"client": fmt.Sprintf("%+v", *cfg.Client),
		}).Debug("client")
	}

	if cfg.Manager != nil {
		log.WithFields(log.Fields{
			"manager": fmt.Sprintf("%+v", *cfg.Manager),
		}).Debug("manager")
	}

	return cfg, nil
}

// TODO: When we persist data
/*
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
*/

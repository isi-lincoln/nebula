package avoid

import (
	"crypto/tls"
	"fmt"
	"sync/atomic"
	"time"

	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"github.com/slackhq/nebula/config"
	"gitlab.com/mergetb/tech/stor"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/yaml.v2"
)

var (
	DefaultAvoidManagerPort = 55554
	DefaultAvoidRelayPort   = 55555
	DefaultAvoidClientPort  = 55556
	DefaultAvoidConfigPath  = "/etc/avoid/avoid.conf"
)

// ServiceConfig encapsulates information for communicating with services.
type ServiceConfig struct {
	Address string
	Port    int
	TLS     *stor.TLSConfig
}

type AlertConfig struct {
	Address  string
	Port     int
	Username string
	Password string
	Type     string
}

// Endpoint returns the endpoint string of a service config.
func (s *ServiceConfig) Endpoint() string {
	return fmt.Sprintf("%s:%d", s.Address, s.Port)
}

// TODO: Make this all atomics
type Avoid struct {
	client   atomic.Bool
	manager  atomic.Bool
	config   string
	identity string
	primary  *Endpoint
	backups  []*Endpoint
	l        *log.Logger
	cert     string
	key      string
}

const (
	ClientOnline        = iota
	ClientOffline       = iota
	ClientDisconnecting = iota
	ClientDisconnected  = iota
	ClientMigrating     = iota
	ClientConnected     = iota
)

func NewAvoidFromConfig(l *log.Logger, c *config.C) *Avoid {
	av := &Avoid{l: l}

	av.reload(c, true)
	c.RegisterReloadCallback(func(c *config.C) {
		av.reload(c, false)
	})

	return av
}

func (av *Avoid) GetPrimary() *Endpoint {
	return av.primary
}

func (av *Avoid) GetBackups() []*Endpoint {
	return av.backups
}

func (av *Avoid) GetClient() bool {
	return av.client.Load()
}

func (av *Avoid) GetManager() bool {
	return av.manager.Load()
}

func (av *Avoid) GetIdentity() string {
	return av.identity
}

func (av *Avoid) GetCertificate() string {
	return av.cert
}

func (av *Avoid) GetKey() string {
	return av.key
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

func tlsHelper(tlsCfg *stor.TLSConfig) (credentials.TransportCredentials, error) {
	if tlsCfg == nil {
		log.Debug("TLS disabled")
		return nil, nil
	} else {
		log.Debug("TLS enabled")

		log.WithFields(log.Fields{
			"cacert": tlsCfg.Cacert,
			"cert":   tlsCfg.Cert,
			"key":    tlsCfg.Key,
		}).Debug("TLS config")

		if tlsCfg.Cacert != "" {
			creds, err := credentials.NewClientTLSFromFile(tlsCfg.Cacert, "")
			if err != nil {
				return nil, fmt.Errorf("failed to load cacert for credentials: %v", err)
			}
			return creds, err
		}
		if tlsCfg.Key != "" && tlsCfg.Cert != "" {
			cert, err := tls.LoadX509KeyPair(tlsCfg.Cert, tlsCfg.Key)
			if err != nil {
				return nil, fmt.Errorf("failed to load key pair for credentials: %v", err)
			}
			// TODO: https://pkg.go.dev/crypto/tls#ClientAuthType
			config := &tls.Config{Certificates: []tls.Certificate{cert}, ClientAuth: tls.NoClientCert}
			return credentials.NewTLS(config), nil
		}

		return nil, nil
	}

}

func WithAvoidRelay(endpoint string, tlsCfg *stor.TLSConfig, f func(AvoidRelayClient) error) error {
	var conn *grpc.ClientConn
	creds, err := tlsHelper(tlsCfg)
	if err != nil {
		return err
	}
	if creds != nil {
		conn, err = grpc.NewClient(endpoint, grpc.WithTransportCredentials(creds))
		if err != nil {
			return fmt.Errorf("failed to connect to avoid relay service (TLS): %v", err)
		}
	} else {
		conn, err = grpc.NewClient(endpoint, grpc.WithInsecure())
		if err != nil {
			return fmt.Errorf("failed to connect to avoid relay service (no TLS): %v", err)
		}
	}

	client := NewAvoidRelayClient(conn)
	defer conn.Close()

	return f(client)
}

func WithAvoidManager(endpoint string, tlsCfg *stor.TLSConfig, f func(AvoidManagerClient) error) error {
	var conn *grpc.ClientConn
	creds, err := tlsHelper(tlsCfg)
	if err != nil {
		return err
	}
	if creds != nil {
		conn, err = grpc.NewClient(endpoint, grpc.WithTransportCredentials(creds))
		if err != nil {
			return fmt.Errorf("failed to connect to avoid manager service (TLS): %v", err)
		}
	} else {
		conn, err = grpc.NewClient(endpoint, grpc.WithInsecure())
		if err != nil {
			return fmt.Errorf("failed to connect to avoid manager service (no TLS): %v", err)
		}
	}

	client := NewAvoidManagerClient(conn)
	defer conn.Close()

	return f(client)
}

func WithAvoidClient(endpoint string, tlsCfg *stor.TLSConfig, f func(AvoidClientClient) error) error {
	var conn *grpc.ClientConn
	creds, err := tlsHelper(tlsCfg)
	if err != nil {
		return err
	}
	if creds != nil {
		conn, err = grpc.NewClient(endpoint, grpc.WithTransportCredentials(creds))
		if err != nil {
			return fmt.Errorf("failed to connect to avoid client service (TLS): %v", err)
		}
	} else {
		conn, err = grpc.NewClient(endpoint, grpc.WithInsecure())
		if err != nil {
			return fmt.Errorf("failed to connect to avoid client service (no TLS): %v", err)
		}
	}

	client := NewAvoidClientClient(conn)
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
type EtcdConfig struct {
	Address string
	Port    int
	TLS     *stor.TLSConfig
	Quantum time.Duration
	Timeout time.Duration
}

// ServicesConfig encapsulates information for communicating with services.
type ServicesConfig struct {
	Client  *ClientServiceConfig  `yaml:client",omitempty"`
	Manager *ManagerServiceConfig `yaml:manager",omitempty"`
	Etcd    *EtcdConfig           `yaml:etcd",omitempty"`
}

// Endpoint returns the endpoint string of a service config.
func (ep *Endpoint) ToAddr() string {
	return fmt.Sprintf("%s:%d", ep.Address, ep.Port)
}

func (ep *Endpoint) GetKey() string {
	if ep.TLS != nil {
		return ep.TLS.Key
	}
	return ""
}

func (ep *Endpoint) GetCert() string {
	if ep.TLS != nil {
		return ep.TLS.Cert
	}
	return ""
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

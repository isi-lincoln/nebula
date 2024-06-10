package avoid

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/slackhq/nebula/config"
	"go.uber.org/atomic"
	grpc "google.golang.org/grpc"
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

type Avoid struct {
	manager  atomic.Bool
	client   atomic.Bool
	port     atomic.Uint32
	address  string
	identity string
	l        *log.Logger
	//endpoints atomic.Pointer[[]string]
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
	if initial {
		var isManager bool

		if c.IsSet("avoid.manager") {
			isManager = c.GetBool("avoid.manager", false)
		}

		av.manager.Store(isManager)
		av.client.Store(!isManager)

		if isManager {
			av.l.Info("avoid manager enabled")
		} else {
			av.l.Info("avoid client enabled")
		}

	} else if c.HasChanged("avoid.manager") || c.HasChanged("avoid.client") {
		//TODO:
		// when a client becomes manager or vise versa
		av.l.Warn("Changing avoid config with reload is not supported, ignoring.")
	}

	/*
		if initial || c.HasChanged("avoid.address") {
			av.address.Store(c.GetString("avoid.address"))
			if !initial {
				av.l.WithField("port", av.GetAddress()).Info("avoid.address changed")
			}
		}
	*/
	// TODO: put into pointer
	av.address = c.GetString("avoid.address", "")
	if av.address == "" {
		av.l.Errorf("No address has been set\n")
		log.Fatal("No address has been set")
	}

	if initial || c.HasChanged("avoid.port") {
		av.port.Store(c.GetUint32("avoid.port", 55554))
		if !initial {
			av.l.WithField("port", av.GetPort()).Info("avoid.port changed")
		}
	}

	/*
		if initial || c.HasChanged("avoid.identity") {
			av.identity.Store(c.GetString("avoid.identity"))
			if !initial {
				av.l.WithField("identity", av.GetIdentity()).Info("avoid.identity changed")
			}
		}
	*/
	av.identity = c.GetString("avoid.identity", "")
	if av.identity == "" {
		av.l.Errorf("No identity has been set\n")
		log.Fatal("No identity has been set")
	}
}

func (av *Avoid) GetManager() bool {
	return av.manager.Load()
}

func (av *Avoid) GetClient() bool {
	return av.client.Load()
}

func (av *Avoid) GetAddress() string {
	return av.address
}

func (av *Avoid) GetPort() uint32 {
	return av.port.Load()
}

func (av *Avoid) GetIdentity() string {
	return av.identity
}

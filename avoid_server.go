package nebula

import (
	"net"

	"github.com/sirupsen/logrus"
	"github.com/slackhq/nebula/avoid"
	"github.com/slackhq/nebula/avoid/service/tunnel"
	"github.com/slackhq/nebula/config"
	"google.golang.org/grpc"
)

func avoidTunnel(l *logrus.Logger, c *config.C, av *avoid.Avoid) func() {
	if av != nil {
		mgr := av.GetManager()
		primary := av.GetPrimary()
		if primary == nil {
			l.Errorf("No primary was found: %#v", av)
			return nil
		}
		if mgr {
			return func() {
				// TODO: maybe better to just use myVPNip
				addr := primary.ToAddr()
				l.Infof("starting avoid tunnel api: %s", addr)

				tunAddr, err := net.Listen("tcp", addr)
				if err != nil {
					l.Fatalf("avoid: failed to listen on %s: %v", addr, err)
				}

				grpcTunnelServer := grpc.NewServer()
				avoid.RegisterTunnelServer(
					grpcTunnelServer,
					tunnel.NewTunnelServer(),
				)
				grpcTunnelServer.Serve(tunAddr)
			}
		}
	}

	return nil
}

package nebula

import (
	"net"

	"github.com/sirupsen/logrus"
	"github.com/slackhq/nebula/avoid"
	"github.com/slackhq/nebula/avoid/service/tunnel"
	"github.com/slackhq/nebula/config"
	"google.golang.org/grpc"
)

func startTunnel(l *logrus.Logger, addr string) {
	// TODO: maybe better to just use myVPNip
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

func avoidTunnel(l *logrus.Logger, c *config.C, av *avoid.Avoid) func() {
	if av != nil {
		mgr := av.GetManager()
		primary := av.GetPrimary()
		if primary == nil {
			l.Errorf("No primary was found: %#v", av)
			return nil
		}
		if mgr {
			addr := primary.ToAddr()
			return func() {
				go startTunnel(l, addr)
			}
		}
	}

	return nil
}
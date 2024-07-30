package nebula

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/netip"

	log "github.com/sirupsen/logrus"
	"github.com/slackhq/nebula/avoid"
	"github.com/slackhq/nebula/config"
	"github.com/slackhq/nebula/iputil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func loadCredentials(cert, key string) (credentials.TransportCredentials, error) {
	// TODO: validate inputs
	serverCert, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		// TODO: https://pkg.go.dev/crypto/tls#ClientAuthType
		ClientAuth: tls.NoClientCert,
	}

	return credentials.NewTLS(config), nil
}

// TODO: implement this
func checkConfigForCerts(c *config.C) (bool, string, string, error) {
	return false, "", "", nil
}

type UEClient struct {
	avoid.UnimplementedAvoidClientServer
	log *log.Logger
}

func (s *UEClient) Action(ctx context.Context, req *avoid.ActionRequest) (*avoid.ConnectionInfo, error) {
	if req == nil {
		return nil, avoid.Error("invalid action request")
	}

	s.log.WithFields(log.Fields{"request": req}).Infof("Action Request")
	return &avoid.ConnectionInfo{}, nil
}

func (s *UEClient) HealthCheck(ctx context.Context, req *avoid.HealthRequest) (*avoid.HealthReply, error) {
	s.log.Infof("liveness check\n")
	return &avoid.HealthReply{}, nil
}

func startAvoidClient(l *log.Logger, addr, cert, key string) error {
	// TODO: maybe better to just use myVPNip
	l.Infof("starting avoid tunnel api: %s", addr)

	tunAddr, err := net.Listen("tcp", addr)
	if err != nil {
		// we fatal here because we dont want nebula to create a channel if we
		// dont have an avoid service running
		l.Fatalf("avoid: failed to listen on %s: %v", addr, err)
	}

	var grpcAvoidClientServer *grpc.Server

	// TODO: plumb certs
	if cert == "" || key == "" {
		grpcAvoidClientServer = grpc.NewServer()
	} else {
		creds, err := loadCredentials(cert, key)
		if err != nil {
			return err
		}
		grpcAvoidClientServer = grpc.NewServer(grpc.Creds(creds))
	}

	avoid.RegisterAvoidClientServer(
		grpcAvoidClientServer,
		&UEClient{log: l},
	)
	grpcAvoidClientServer.Serve(tunAddr)

	// never should reach here
	return nil
}

func checkIfStartAvoidClient(l *log.Logger, av *avoid.Avoid) func() {
	if av != nil {
		mgr := av.GetManager()
		primary := av.GetPrimary()
		if primary == nil {
			l.Errorf("No primary was found: %#v", av)
			return nil
		}
		if mgr {
			addr := primary.ToAddr()
			cert := primary.GetCert()
			key := primary.GetKey()
			return func() {
				go startAvoidClient(l, addr, cert, key)
			}
		}
	}

	return nil
}

type AvoidClient struct {
	avoid.UnimplementedAvoidClientServer
	token   string
	iface   *Interface
	hostmap *HostMap
}

func NewAvoidClient(token string, iface *Interface) *AvoidClient {
	return &AvoidClient{token: token, iface: iface}
}

func (s *AvoidClient) Action(ctx context.Context, req *avoid.ActionRequest) (*avoid.ConnectionInfo, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: Action")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}
	fields := log.Fields{"request": req}
	log.WithFields(fields).Info("Action Request")

	switch req.Action.Action {
	case avoid.ActionMessage_MIGRATE:
		break
		switch req.Action.Connection {
		case avoid.ActionMessage_RELAY:

			// get hostinfo of dst
			// TODO: Make better - add to hostmap as necessary
			// TODO: validate inputs
			if len(req.Values) != 2 {
				return nil, fmt.Errorf("invalid number of values.  Need [0] = dst, [1] = new relay.  got: %v", req.Values)
			}
			addr, err := netip.ParseAddr(req.Values[0])
			if err != nil {
				log.WithError(err).Error("Migrate Action Value not valid ip: %v", req.Values[0])
				return nil, err
			}

			// addr is 128 for ipv6, but nebula is ipv4
			dstip := iputil.Ip2VpnIp(addr.AsSlice())

			hostinfo := s.hostmap.QueryVpnIp(dstip)
			if hostinfo == nil {
				return nil, fmt.Errorf("destination not found in hostmap: %s", dstip.String())
			}

			// set new peer ip
			// TODO: all the parsing things
			addr, err = netip.ParseAddr(req.Values[1])
			if err != nil {
				return nil, err
			}

			// addr is 128 for ipv6, but nebula is ipv4
			peer := iputil.Ip2VpnIp(addr.AsSlice())

			// setup the relay change
			err = MigrateRelayUsed(hostinfo, peer, s.iface.l, s.iface)
			if err != nil {
				log.WithError(err).Errorf("failed in Action: Migrate")
				return nil, err
			}

			return &avoid.ConnectionInfo{Relay: peer.String()}, nil

		case avoid.ActionMessage_LIGHTHOUSE:
			log.WithFields(fields).Error("Lighthouse not implemented")
			return nil, fmt.Errorf("Lighthouse not implemented")
		default:
			log.WithFields(fields).Errorf("Unsupported Connection Type: %#v", req.Action.Connection)
			return nil, fmt.Errorf("Unsupported Connection Type: %#v", req.Action.Connection)
		}
	case avoid.ActionMessage_DISCONNECT:
		log.WithFields(fields).Error("Disconnect not implemented")
		return nil, fmt.Errorf("Disconnect not implemented")
	default:
		log.WithFields(fields).Errorf("Unsupported Action: %#v", req.Action.Action)
		return nil, fmt.Errorf("Unsupported Action: %#v", req.Action.Action)
	}
	return &avoid.ConnectionInfo{}, nil
}

func (s *AvoidClient) HealthCheck(ctx context.Context, req *avoid.HealthRequest) (*avoid.HealthReply, error) {
	log.Infof("liveness check\n")

	return &avoid.HealthReply{}, nil
}

func startAvoidClientService(addr, token string, iface *Interface) {
	clientAddr, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on tunnel addr: %v", err)
	}
	grpcAvoidClientServer := grpc.NewServer()
	avoid.RegisterAvoidClientServer(grpcAvoidClientServer, NewAvoidClient(token, iface))
	grpcAvoidClientServer.Serve(clientAddr)
}

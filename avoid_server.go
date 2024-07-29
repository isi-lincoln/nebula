package nebula

import (
	"context"
	"crypto/tls"
	"net"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/slackhq/nebula/avoid"
	"github.com/slackhq/nebula/config"
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
	token string
}

func NewAvoidClient(token string) *AvoidClient {
	return &AvoidClient{token: token}
}

func (s *AvoidClient) Action(ctx context.Context, req *avoid.ActionRequest) (*avoid.ConnectionInfo, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: Action")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	log.WithFields(log.Fields{"request": req}).Info("Action Request")

	return &avoid.ConnectionInfo{}, nil
}

func (s *AvoidClient) HealthCheck(ctx context.Context, req *avoid.HealthRequest) (*avoid.HealthReply, error) {
	log.Infof("liveness check\n")

	return &avoid.HealthReply{}, nil
}

func startAvoidClientService(addr, token string) {
	clientAddr, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on tunnel addr: %v", err)
	}
	grpcAvoidClientServer := grpc.NewServer()
	avoid.RegisterAvoidClientServer(grpcAvoidClientServer, NewAvoidClient(token))
	grpcAvoidClientServer.Serve(clientAddr)
}

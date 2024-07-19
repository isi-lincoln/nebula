package nebula

import (
	"crypto/tls"
	"net"

	"github.com/sirupsen/logrus"
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
}

func startAvoidClient(l *logrus.Logger, addr, cert, key string) error {
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
		UEClient{},
	)
	grpcAvoidClientServer.Serve(tunAddr)

	// never should reach here
	return nil
}

func checkIfStartAvoidClient(l *logrus.Logger, av *avoid.Avoid) func() {
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

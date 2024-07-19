package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/coreos/etcd/clientv3"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/slackhq/nebula/avoid"
	"github.com/slackhq/nebula/avoid/service/relay"
	"gitlab.com/mergetb/tech/stor"
	"google.golang.org/grpc"
)

var (
	Build string
	etcd  *clientv3.Client
)

type AvoidRelay struct {
	avoid.UnimplementedAvoidRelayServer
}

func NewAvoidRelay() *AvoidRelay {
	return &AvoidRelay{}
}

func (s *AvoidRelay) Register(ctx context.Context, req *avoid.RegisterRequest) (*avoid.RegisterReply, error) {
	if req == nil {
		return nil, avoid.Error("Invalid Request: Register")
	}
	if req.Name == "" {
		return nil, avoid.Error("Registration requires a name")
	}
	if req.Ip == "" {
		return nil, avoid.Error("Registration requires an ip address")
	}
	if req.Port == 0 {
		return nil, avoid.Error("Registration requires a port address")
	}

	// TODO: sanitize name, ip, port

	// get our host name to put in as the EP
	hostname, err := os.Hostname()
	if err != nil {
		return -1, err
	}

	// TODO: validate that name should be in avoid network

	// Once we know this is a good host, generate a token
	// TODO: use TOTP/HOTP, or another method - as this just means you listen
	// to Register requests to find tokens - even encrypted no bueno

	token := uuid.Must(uuid.NewV4()).String()

	reg := &avoid.Registration{
		UE:          req.Name,
		EP:          hostname,
		Token:       token,
		IP:          req.Ip,
		DNS:         req.Dns,
		Certificate: req.Cert,
		Port:        req.Port,
	}

	fields := log.Fields{"registration": reg}

	// now we need to store the registration object to be used later
	err = stor.WriteObjects(am, reg)
	if err != nil {
		return nil, err
	}

	avoid.InfoF("Registration Complete", fields)

	return &avoid.RegisterReply{Token: token}, nil
}

func (s *AvoidClient) HealthCheck(ctx context.Context, req *avoid.HealthRequest) (*avoid.HealthReply, error) {
	log.Debugf("liveness check\n")
	return &avoid.HealthReply{}, nil
}

func main() {
	printVersion := flag.Bool("version", false, "Print version")
	printUsage := flag.Bool("help", false, "Print command line usage")

	relayPort := flag.Int("port", avoid.DefaultAvoidRelayPort, "port to configure relay server")
	relayServer := flag.String("addr", "0.0.0.0", "relay server address")

	debug := flag.Bool("debug", false, "enable extra debugging")
	avoidConf := flag.String("conf", avoid.DefaultAvoidConfigPath, "avoid configuration file path")

	flag.Parse()

	if *printVersion {
		fmt.Printf("Version: %s\n", Build)
		os.Exit(0)
	}

	if *printUsage {
		flag.Usage()
		os.Exit(0)
	}

	if *debug {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	log.Infof("starting avoid relay api: %s:%d", *relayServer, *relayPort)

	relayAddr, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *relayServer, *relayPort))
	if err != nil {
		log.Fatalf("failed to listen on relay addr: %v", err)
	}

	cfg, err := avoid.LoadConfig(avoidConf)
	if err != nil {
		log.Fatalf("%v", err)
	}

	etcdCfg, err := avoid.SetEtcdSettings(cfg)
	if err != nil {
		log.Fatalf("%v", err)
	}

	stor.SetConfig(*etcdCfg)

	err := avoid.EnsureEtcd(&etcd)
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("connected to etcd")

	grpcTunnelServer := grpc.NewServer()
	avoid.RegisterTunnelServer(grpcTunnelServer, relay.NewAvoidRelayServer())
	grpcTunnelServer.Serve(relayAddr)

	os.Exit(0)
}

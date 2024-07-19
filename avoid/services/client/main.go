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
	"github.com/slackhq/nebula/avoid/service/tunnel"
	"gitlab.com/mergetb/tech/stor"
	"google.golang.org/grpc"
)

var (
	Build string
	etcd  *clientv3.Client
)

type AvoidClient struct {
	avoid.UnimplementedAvoidClientServer
}

func NewAvoidClient() *AvoidClient {
	return &AvoidClient{}
}

func (s *AvoidClient) Action(ctx context.Context, req *avoid.ActionRequest) (*avoid.ConnectionInfo, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: Action")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	avoid.InfoF("Action Request", log.fields{"request": req})

	// TODO: this code is mainly for testing
	// so implement some more functions here

	return &avoid.ConnectionInfo{}, nil
}

func (s *AvoidClient) HealthCheck(ctx context.Context, req *avoid.HealthRequest) (*avoid.HealthReply, error) {
	log.Infof("liveness check\n")

	return &avoid.HealthReply{}, nil
}

func main() {
	printVersion := flag.Bool("version", false, "Print version")
	printUsage := flag.Bool("help", false, "Print command line usage")

	clientPort := flag.Int("port", avoid.DefaultAvoidClientPort, "port to configure tunnel server")
	clientServer := flag.String("server", "0.0.0.0", "tunnel server address or interface")

	debug := flag.Bool("debug", false, "enable extra debugging")

	flag.Parse()

	if *printVersion {
		fmt.Printf("Version: %s\n", Build)
		os.Exit(0)
	}

	if *printUsage {
		flag.Usage()
		os.Exit(0)
	}

	// daemon mode
	if *debug {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	log.Infof("starting avoid client api: %s:%d", *clientServer, *clientPort)

	clientAddr, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *clientServer, *clientPort))
	if err != nil {
		log.Fatalf("failed to listen on tunnel addr: %v", err)
	}

	cfg, err := avoid.LoadConfig(EtcdConfigPath)
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
	log.Debug("connected to etcd")

	grpcTunnelServer := grpc.NewServer()
	avoid.RegisterAvoidClientServer(grpcAvoidClientServer, tunnel.NewAvoidClientServer())
	grpcAvoidClientServer.Serve(tunAddr)

	os.Exit(0)
}

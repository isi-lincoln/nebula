package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/slackhq/nebula/avoid"
	clientv3 "go.etcd.io/etcd/client/v3"
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

	log.WithFields(log.Fields{"request": req}).Info("Action Request")

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

	clientPort := flag.Int("port", avoid.DefaultAvoidClientPort, "server port")
	clientServer := flag.String("server", "0.0.0.0", "server address")

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

	// daemon mode
	if *debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.Infof("starting avoid client api: %s:%d", *clientServer, *clientPort)

	clientAddr, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *clientServer, *clientPort))
	if err != nil {
		log.Fatalf("failed to listen on tunnel addr: %v", err)
	}

	cfg, err := avoid.LoadConfig(*avoidConf)
	if err != nil {
		log.Fatalf("%v", err)
	}

	etcdCfg, err := avoid.GetEtcdConfig(cfg)
	if err != nil {
		log.Fatalf("%v", err)
	}

	err = avoid.SetConfig(etcdCfg)
	if err != nil {
		log.Fatalf("%v", err)
	}

	err = avoid.EnsureEtcd(&etcd)
	if err != nil {
		log.Fatal(err)
	}
	log.Debug("connected to etcd")

	grpcAvoidClientServer := grpc.NewServer()
	avoid.RegisterAvoidClientServer(grpcAvoidClientServer, NewAvoidClient())
	grpcAvoidClientServer.Serve(clientAddr)

	os.Exit(0)
}

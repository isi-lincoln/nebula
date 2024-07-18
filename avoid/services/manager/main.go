package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
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

type ClientState struct {
	Current    int
	Transition int
	Action     *avoid.ConnectionReply
}

type AvoidManager struct {
	avoid.UnimplementedAvoidManagerServer
	updates map[string]*ClientState
}

func NewAvoidManager() *AvoidManager {
	return &AvoidManager{updates: make(map[string]*ClientState)}
}

func (s *AvoidManager) ListConnections(ctx context.Context, req *avoid.ListRequest) (*avoid.ListReply, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: ListConnections")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	log.Infof("List Request")

	lr := make([]*avoid.ConnectionInfo, 0)
	for k, _ := range s.updates {
		tmp := &avoid.ConnectionInfo{
			Name: k,
		}
		lr = append(lr, tmp)
	}

	// TODO: More data from the connection

	return &avoid.ListReply{Info: lr}, nil
}

func (s *AvoidManager) GetStats(ctx context.Context, req *avoid.StatsRequest) (*avoid.StatsReply, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: GetStats")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	uuid := uuid.New()
	log.Infof("Get Stats: %s", req.Name)

	// TODO: Return Stats

	return &avoid.StatsReply{}, nil
}

func sendToPending(ar *avoid.ActionRequest) error {
	// create pending action
	pending := &avoid.Pending{
		ActionKey: ar.Key(),
	}

	var err error
	// TODO: retry loop, make better to handle backoff, etc.
	for i := 0; i < 10; i++ {
		// write objects as all or nothing
		err = stor.WriteObjects(am, pending)
		if err != nil {
			log.Warnf("%d: Retrying Write to Pending: %v", i, err)
			continue
		}
		return nil
	}

	return err
}

func waitForAction(ar *avoid.ActionRequest, timeout int) (*avoid.ConnectionInfo, error) {
	key := fmt.Sprintf("%s/%s", avoid.ConnPrefix, ar.Uuid)
	ctx, cancel := context.WithTimeout(context.TODO(), timeout*time.Second)
	defer cancel()

	avoid.EnsureEtcd(*etcd)
	rch := (*etcd).Watch(ctx, key, clientv3.WithPrefix())

	// we've found a key if rch != nil
	for wresp := range rch {
		for _, x := range wresp.Events {
			// cleaning up key event
			ci := &avoid.ConnectionInfo{}
			err := json.Unmarshal(x.Kv.Value, ci)
			if err != nil {
				// TODO: manage this situation
				log.Errorf("failed to unmarshal for %s: %v", key, err)
				return nil, err
			}
			return ci, nil
		}

	}

	// our key was nil because we timed out
	return nil, nil
}

func (s *AvoidManager) Migrate(ctx context.Context, req *avoid.MigrateRequest) (*avoid.MigrateReply, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: Migrate")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}
	if req.Migrate == nil {
		errMsg := fmt.Sprintf("Invalid Migrate: %v", req)
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	uuid := uuid.New()
	log.Infof("Migrate: %s: %v", uuid.String(), req.Migrate)

	ar := req.Migrate
	if ar.Action == nil {
		errMsg := fmt.Sprintf("Migrate missing action: %v", req)
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}
	ar.Action.Uuid = uuid
	ar.Action.Version = 0

	//TODO: sanity check identifier, values, Action

	err = sendToPending(ar)
	if err != nil {
		return nil, err
	}

	// timeout 10 seconds
	// wait to see if pending task is picked up
	// wait until we see the action being finished
	msg, err = waitForAction(ar, 10)
	if err != nil {
		return nil, err
	}

	if msg == nil {
		errMsg := "Action returned an empty message"
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	return &avoid.MigrateReply{Migrate: msg}, nil
}

func (s *AvoidManager) Disconnect(ctx context.Context, req *avoid.DisconnectRequest) (*avoid.DisconnectReply, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: Disconnect")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	uuid := uuid.New()
	log.Infof("Disconnect: %s: %v", uuid.String(), req)

	return &avoid.DisconnectReply{}, nil
}

func (s *AvoidManager) Register(ctx context.Context, req *avoid.RegisterRequest) (*avoid.RegisterReply, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: Register")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	log.Infof("Register request: %v\n", req)

	return &avoid.RegisterReply{Token: req.Req}, nil
}

func (s *AvoidManager) HealthCheck(ctx context.Context, req *avoid.HealthRequest) (*avoid.HealthReply, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: Health")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	log.Infof("HC\n")

	return &avoid.HealthReply{Json: "TODO"}, nil
}

func main() {
	printVersion := flag.Bool("version", false, "Print version")
	printUsage := flag.Bool("help", false, "Print command line usage")

	tunnelPort := flag.Int("tunport", 55554, "port to configure tunnel server")
	tunnelServer := flag.String("tunserver", "0.0.0.0", "tunnel server address or interface")

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

	log.Infof("starting avoid tunnel api: %s:%d", *tunnelServer, *tunnelPort)

	tunAddr, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *tunnelServer, *tunnelPort))
	if err != nil {
		log.Fatalf("failed to listen on tunnel addr: %v", err)
	}

	cfg, err := config.LoadConfig(EtcdConfigPath)
	if err != nil {
		log.Fatalf("%v", err)
	}

	// read in environment variables for container
	err = config.ReadENVSettings(cfg)
	if err != nil {
		log.Fatalf("%v", err)
	}

	etcdCfg, err := config.SetEtcdSettings(cfg)
	if err != nil {
		log.Fatalf("%v", err)
	}

	stor.SetConfig(*etcdCfg)

	err := avoid.EnsureEtcd(&etcd)
	if err != nil {
		log.Fatal(err)
	}
	log.Trace("connected to etcd")

	grpcTunnelServer := grpc.NewServer()
	avoid.RegisterTunnelServer(grpcTunnelServer, tunnel.NewTunnelServer())
	grpcTunnelServer.Serve(tunAddr)

	os.Exit(0)
}

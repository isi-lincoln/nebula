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
	"gitlab.com/mergetb/tech/stor"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

var (
	Build string
)

type AvoidManager struct {
	avoid.UnimplementedAvoidManagerServer
}

func NewAvoidManager() *AvoidManager {
	return &AvoidManager{}
}

func (s *AvoidManager) ListConnections(ctx context.Context, req *avoid.ListRequest) (*avoid.ListReply, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: ListConnections")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	log.Infof("List Request")

	// TODO: More data from the connection

	return &avoid.ListReply{}, nil
}

func (s *AvoidManager) GetStats(ctx context.Context, req *avoid.StatsRequest) (*avoid.ConnectionInfo, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: GetStats")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	//uuid := uuid.New()
	log.Infof("Get Stats: %s", req.Name)

	// TODO: Return Stats

	return &avoid.ConnectionInfo{}, nil
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
		objs := []stor.Object{ar, pending}
		err = stor.WriteObjects(objs, true)
		if err != nil {
			log.Warnf("%d: Retrying Write to Pending: %v", i, err)
			continue
		}
		return nil
	}

	return err
}

func waitForAction(ar *avoid.ActionRequest, timeout int) (connInfo *avoid.ConnectionInfo, err error) {
	key := fmt.Sprintf("%s/%s", avoid.ConnPrefix, ar.Uuid)
	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(timeout)*time.Second)
	defer cancel()

	err = stor.WithEtcd(func(etcdp *clientv3.Client) error {
		rch := (*etcdp).Watch(ctx, key, clientv3.WithPrefix())

		// we've found a key if rch != nil
		for wresp := range rch {
			for _, x := range wresp.Events {
				// cleaning up key event
				ci := &avoid.ConnectionInfo{}
				err := json.Unmarshal(x.Kv.Value, ci)
				if err != nil {
					// TODO: manage this situation
					log.Errorf("failed to unmarshal for %s: %v", key, err)
					return err
				}
				connInfo = ci
				return nil
			}

		}

		return nil
	})

	return
}

func (s *AvoidManager) Action(ctx context.Context, req *avoid.ActionRequest) (*avoid.ConnectionInfo, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: Migrate")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	uuidTracker := uuid.New()
	log.Infof("Action: %s: %v", uuidTracker.String(), req)

	ar := req
	if ar.Action == nil {
		errMsg := fmt.Sprintf("Migrate missing action: %v", req)
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}
	ar.Action.Uuid = uuidTracker.String()
	ar.Action.Version = 0

	//TODO: sanity check identifier, values, Action

	err := sendToPending(ar)
	if err != nil {
		return nil, err
	}

	// timeout 10 seconds
	// wait to see if pending task is picked up
	// wait until we see the action being finished
	msg, err := waitForAction(ar, 10)
	if err != nil {
		return nil, err
	}

	if msg == nil {
		errMsg := "Action returned an empty message"
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	return &avoid.ConnectionInfo{}, nil
}

func main() {
	printVersion := flag.Bool("version", false, "Print version")
	printUsage := flag.Bool("help", false, "Print command line usage")

	mgmtPort := flag.Int("port", avoid.DefaultAvoidManagerPort, "management server port")
	mgmtServer := flag.String("addr", "0.0.0.0", "mgmt server address or interface")

	avoidConf := flag.String("conf", avoid.DefaultAvoidConfigPath, "avoid configuration file path")
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

	log.Infof("starting avoid manager api: %s:%d", *mgmtServer, *mgmtPort)

	mgmtAddr, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *mgmtServer, *mgmtPort))
	if err != nil {
		log.Fatalf("failed to listen on mgmt addr: %v", err)
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

	grpcTunnelServer := grpc.NewServer()
	avoid.RegisterAvoidManagerServer(grpcTunnelServer, NewAvoidManager())
	grpcTunnelServer.Serve(mgmtAddr)

	os.Exit(0)
}

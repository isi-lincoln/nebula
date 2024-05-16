package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/slackhq/nebula/avoid"
	"google.golang.org/grpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	Build string
)

type ManagementServer struct {
	avoid.UnimplementedManagementServer
}

func (s *ManagementServer) ListConnections(ctx context.Context, req *avoid.ListRequest) (*avoid.ListReply, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: ListConnections")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	// TODO: Connection Management

	return &avoid.ListReply{}, nil
}

func (s *ManagementServer) GetStats(ctx context.Context, req *avoid.StatsRequest) (*avoid.StatsReply, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: GetStats")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	// TODO: Statistics Management

	return &avoid.StatsReply{}, nil
}

func (s *ManagementServer) Migrate(ctx context.Context, req *avoid.MigrateRequest) (*avoid.MigrateReply, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: Migrate")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	// TODO: Migration

	return &avoid.MigrateReply{}, nil
}

func (s *ManagementServer) Shutdown(ctx context.Context, req *avoid.ShutdownRequest) (*avoid.ShutdownReply, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: Shutdown")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	// TODO: Shutdown/Revocation

	return &avoid.ShutdownReply{}, nil
}

type TunnelServer struct {
	avoid.UnimplementedTunnelServer
	mu sync.RWMutex
	// If shutdown is true, it's expected all serving status is NOT_SERVING, and
	// will stay in NOT_SERVING.
	shutdown bool
	// statusMap stores the serving status of the services this Server monitors.
	statusMap map[string]*avoid.ConnectionReply
	updates   map[string]map[avoid.Tunnel_WatchServer]chan *avoid.ConnectionReply
}

func (s *TunnelServer) Register(ctx context.Context, req *avoid.RegisterRequest) (*avoid.RegisterReply, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: Register")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	// TODO: Register

	return &avoid.RegisterReply{}, nil
}

func (s *TunnelServer) HealthCheck(ctx context.Context, req *avoid.HealthRequest) (*avoid.HealthReply, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: Health")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	// TODO: Health Checks

	return &avoid.HealthReply{}, nil
}

// originally from here: https://github.com/grpc/grpc-go/blob/v1.64.0/health/server.go
// APL2 license
func (s *TunnelServer) Watch(in *avoid.ConnectionRequest, stream avoid.Tunnel_WatchServer) error {
	service := in.Name
	// update channel is used for getting service status updates.
	update := make(chan *avoid.ConnectionReply, 1)
	s.mu.Lock()
	// Puts the initial status to the channel.
	if servingStatus, ok := s.statusMap[service]; ok {
		update <- servingStatus
	} else {
		//update <- healthpb.TunnelCheckResponse_SERVICE_UNKNOWN
	}

	// Registers the update channel to the correct place in the updates map.
	if _, ok := s.updates[service]; !ok {
		s.updates[service] = make(map[avoid.Tunnel_WatchServer]chan *avoid.ConnectionReply)
	}
	s.updates[service][stream] = update
	defer func() {
		s.mu.Lock()
		delete(s.updates[service], stream)
		s.mu.Unlock()
	}()
	s.mu.Unlock()

	var lastSentStatus *avoid.ConnectionReply = &avoid.ConnectionReply{}
	for {
		select {
		// Status updated. Sends the up-to-date status to the client.
		case servingStatus := <-update:
			if lastSentStatus == servingStatus {
				continue
			}
			lastSentStatus = servingStatus
			err := stream.Send(servingStatus)
			if err != nil {
				return status.Error(codes.Canceled, "Stream has ended.")
			}
			// Context done. Removes the update channel from the updates map.
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "Stream has ended.")
		}
	}
}

func main() {
	printVersion := flag.Bool("version", false, "Print version")
	printUsage := flag.Bool("help", false, "Print command line usage")

	mgmtPort := flag.Int("mgmtport", 55555, "port to configure management server")
	mgmtServer := flag.String("mgmtserver", "0.0.0.0", "management server address or interface")
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

	log.Infof("starting avoid mgmt api: %s:%d", mgmtServer, mgmtPort)
	log.Infof("starting avoid tunnel api: %s:%d", tunnelServer, tunnelPort)

	mgmtAddr, err := net.Listen("tcp", fmt.Sprintf("%s:%d", mgmtServer, mgmtPort))
	if err != nil {
		log.Fatalf("failed to listen on mgmt addr: %v", err)
	}

	tunAddr, err := net.Listen("tcp", fmt.Sprintf("%s:%d", tunnelServer, tunnelPort))
	if err != nil {
		log.Fatalf("failed to listen on tunnel addr: %v", err)
	}

	grpcManagementServer := grpc.NewServer()
	avoid.RegisterManagementServer(grpcManagementServer, &ManagementServer{})
	grpcTunnelServer := grpc.NewServer()
	avoid.RegisterTunnelServer(grpcTunnelServer, &TunnelServer{})
	go func() {
		grpcTunnelServer.Serve(tunAddr)
	}()
	grpcManagementServer.Serve(mgmtAddr)

	os.Exit(0)
}

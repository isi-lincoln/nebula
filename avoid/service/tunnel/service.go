package tunnel

import (
	"context"
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/slackhq/nebula/avoid"
)

// TODO: This approach is non-persistent
type TunnelUpdate struct {
	Watch   *avoid.Tunnel_WatchServer
	Ch      chan avoid.ConnectionReply_ServingStatus
	Current avoid.ConnectionReply_ServingStatus
	Data    *avoid.ConnectionReply
}

type TunnelServer struct {
	avoid.UnimplementedTunnelServer
	mu      sync.RWMutex
	updates map[string]*TunnelUpdate
}

func NewTunnelServer() *TunnelServer {
	return &TunnelServer{updates: make(map[string]*TunnelUpdate)}
}

func (s *TunnelServer) ListConnections(ctx context.Context, req *avoid.ListRequest) (*avoid.ListReply, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: ListConnections")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	log.Infof("List Request")

	lr := make([]*avoid.ConnInfo, 0)
	for k, _ := range s.updates {
		tmp := &avoid.ConnInfo{
			Name: k,
		}
		lr = append(lr, tmp)
	}

	// TODO: Connection Management

	return &avoid.ListReply{Info: lr}, nil
}

func (s *TunnelServer) GetStats(ctx context.Context, req *avoid.StatsRequest) (*avoid.StatsReply, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: GetStats")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	log.Infof("Get Stats: %s", req.Name)

	// TODO: Statistics Management

	return &avoid.StatsReply{}, nil
}

func (s *TunnelServer) Migrate(ctx context.Context, req *avoid.MigrateRequest) (*avoid.MigrateReply, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: Migrate")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	log.Infof("Migrate: %v", req.Migrate)
	rm := req.Migrate

	s.mu.Lock()
	_, ok := s.updates[req.Name]
	if !ok {
		s.updates[req.Name] = &TunnelUpdate{
			Watch:   nil,
			Ch:      make(chan avoid.ConnectionReply_ServingStatus, 1),
			Current: avoid.ConnectionReply_SERVICE_UNKNOWN,
			Data: &avoid.ConnectionReply{
				Status:     avoid.ConnectionReply_SERVICE_UNKNOWN,
				Connection: avoid.ConnectionReply_Lighthouse,
				Value:      "",
			},
		}
	} else {
		tmp := &avoid.ConnectionReply{
			Status:     avoid.ConnectionReply_SERVING,
			Value:      rm.Value,
			Connection: rm.Connection,
		}
		s.updates[req.Name].Data = tmp
	}
	s.updates[req.Name].Ch <- avoid.ConnectionReply_SERVING
	s.mu.Unlock()

	return &avoid.MigrateReply{}, nil
}

func (s *TunnelServer) Shutdown(ctx context.Context, req *avoid.ShutdownRequest) (*avoid.ShutdownReply, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: Shutdown")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	log.Infof("Shutdown: %v", req)

	// TODO: Shutdown/Revocation

	return &avoid.ShutdownReply{}, nil
}

func (s *TunnelServer) Register(ctx context.Context, req *avoid.RegisterRequest) (*avoid.RegisterReply, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: Register")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	log.Infof("Register\n")

	// TODO: Register

	return &avoid.RegisterReply{Token: "TODO"}, nil
}

func (s *TunnelServer) HealthCheck(ctx context.Context, req *avoid.HealthRequest) (*avoid.HealthReply, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: Health")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	log.Infof("HC\n")

	// TODO: Health Checks

	return &avoid.HealthReply{Json: "TODO"}, nil
}

// originally from here: https://github.com/grpc/grpc-go/blob/v1.64.0/health/server.go
// APL2 license
// TODO: I dont think this is scalable.
func (s *TunnelServer) Watch(in *avoid.ConnectionRequest, stream avoid.Tunnel_WatchServer) error {

	log.Infof("Watch: %s\n", in.Name)
	service := in.Name

	s.mu.Lock()
	val, ok := s.updates[service]
	s.mu.Unlock()
	if !ok {
		tu := &TunnelUpdate{
			Watch:   &stream,
			Ch:      make(chan avoid.ConnectionReply_ServingStatus, 1),
			Current: avoid.ConnectionReply_SERVICE_UNKNOWN,
			Data: &avoid.ConnectionReply{
				Status:     avoid.ConnectionReply_SERVICE_UNKNOWN,
				Connection: avoid.ConnectionReply_Lighthouse,
				Value:      "",
			},
		}
		s.mu.Lock()
		s.updates[service] = tu
		log.Infof("Added new: %#v\n", s.updates[service])
		s.updates[service].Ch <- avoid.ConnectionReply_SERVICE_UNKNOWN
		s.mu.Unlock()
		for {
			select {
			case servingStatus := <-s.updates[service].Ch:
				log.Infof("Something in channel: %s\n", servingStatus)
				if servingStatus == avoid.ConnectionReply_SERVICE_UNKNOWN {
					log.Infof("Nothing sent\n")
					continue
				}

				s.mu.Lock()
				tu := s.updates[service]
				log.Infof("Data: %#v\n", tu)
				if tu.Watch != nil {
					(*tu.Watch).Send(tu.Data)
					// kill the connection? or not?
					// break
				} else {
					log.Errorf("Invalid tunnel update. Watch is nil: %v\n", s.updates[service].Data)
				}
				s.mu.Unlock()
			}
		}
	} else {
		stream.Send(val.Data)
	}

	return nil
}

func (s *TunnelServer) Watcher(service string) error {
	return nil
}

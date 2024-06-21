package tunnel

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/slackhq/nebula/avoid"
)

// TODO: This approach is non-persistent
type TunnelUpdate struct {
	rwmutex     sync.RWMutex
	Watch       *avoid.Tunnel_WatchServer
	SendChannel chan avoid.ConnectionMessage_ServingStatus
	RecvChannel chan avoid.ConnectionMessage_ServingStatus
	Current     avoid.ConnectionMessage_ServingStatus
	Transition  avoid.ConnectionMessage_ServingStatus
	Action      *avoid.ConnectionReply
}

type TunnelServer struct {
	avoid.UnimplementedTunnelServer
	mu           sync.RWMutex
	updates      map[string]*TunnelUpdate
	notification chan string
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

func (s *TunnelServer) manageRecv(ueUuid, uuid string, status avoid.ConnectionMessage_ServingStatus) (*avoid.ConnectionRequest, error) {

	log.Infof("manage Recv: %s\n", ueUuid)

	s.mu.Lock()
	tu, ok := s.updates[ueUuid]
	s.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("could not find %s in TunnelUpdates", ueUuid)
	}

	// TODO: pass in timeout
	for i := 0; i < 50; i++ {
		tu.rwmutex.Lock()
		select {
		case servingStatus := <-s.updates[ueUuid].RecvChannel:
			log.Infof("Recv in channel: %s\n", servingStatus)
			if servingStatus == status {

				tuCurrent := s.updates[ueUuid]
				if tuCurrent != nil {
					log.Errorf("TunnelUpdate is nil for %s\n", ueUuid)
				}

				log.Infof("Recv Data: %#v\n", tu)
				if tu.Watch != nil {
					resp, err := (*tu.Watch).Recv()
					if err == io.EOF {
						time.Sleep(100 * time.Millisecond)
						continue
					}
					if err != nil {
						log.Errorf("Recv error in stream: %v\n", err)
						time.Sleep(100 * time.Millisecond)
						continue
					}
					tu.rwmutex.Unlock()
					s.mu.Lock()
					s.notifications <- ueUuid
					s.mu.Unlock()
					return resp, nil
				} else {
					log.Errorf("Invalid tunnel update. Watch is nil: %v\n", s.updates[ueUuid])
				}
			} else {
				// TODO: this is where in concurrency something could fall through the cracks
			}
		}

		tu.rwmutex.Unlock()
		time.Sleep(100 * time.Millisecond)
	}

	return nil, fmt.Errorf("Message Recv Timeout\n")
}

/*
 * Helper function to manage all services that need to use the Watch channel
 * This function is responsible for taking inputs on type, status, action, value
 * to populate a ConnectionMessage to prepare.
 * First it is responsible for checking the state of the global lock and updating
 * global state before updating the channel.
 */
func (s *TunnelServer) manageSend(
	ueUuid string,
	ct avoid.ConnectionMessage_ConnType,
	status avoid.ConnectionMessage_ServingStatus,
	action avoid.ConnectionMessage_Action,
	value []string,
	uuid string,
) error {
	s.mu.Lock()
	_, ok := s.updates[ueUuid]
	s.mu.Unlock()
	if !ok {
		s.mu.Lock()
		s.updates[ueUuid] = &TunnelUpdate{
			Watch:       nil,
			SendChannel: make(chan avoid.ConnectionMessage_ServingStatus, 1),
			Current:     avoid.ConnectionMessage_SERVICE_UNKNOWN,
			Transition:  status,
			Action: &avoid.ConnectionReply{
				Connection: &avoid.ConnectionMessage{
					Status:     status,
					Connection: ct,
					Action:     action,
					Uuid:       ueUuid,
				},
				Value: value,
			},
		}
		s.updates[ueUuid].SendChannel <- avoid.ConnectionMessage_SERVING
		s.mu.Unlock()
	} else {
		s.mu.Lock()
		tu := s.updates[ueUuid]
		s.mu.Unlock()
		if tu == nil {
			log.Errorf("something happened to our TunnelUpdates for %s\n", ueUuid)
			return fmt.Errorf("key missing from updates struct: %s", ueUuid)
		}
		tu.rwmutex.Lock()

		tu.Transition = status
		tua := tu.Action
		if tua == nil {
			log.Errorf("action is missing from TunnelUpdates for %s: %+v\n", ueUuid, tu)
			tu.rwmutex.Unlock()
			return fmt.Errorf("action missing from updates struct: %s", ueUuid)
		}

		tuac := tua.Connection
		if tuac == nil {
			log.Errorf("connection is missing from TunnelUpdates for %s: %+v\n", ueUuid, tua)
			tu.rwmutex.Unlock()
			return fmt.Errorf("connection missing from updates struct: %s", ueUuid)
		}

		tuac.Status = status
		tuac.Connection = ct
		tuac.Action = action
		tuac.Uuid = uuid

		s.updates[ueUuid].SendChannel <- status
		tu.rwmutex.Unlock()
	}

	s.mu.Lock()
	s.notifications <- ueUuid
	s.mu.Unlock()

	return nil
}

func (s *TunnelServer) GetStats(ctx context.Context, req *avoid.StatsRequest) (*avoid.StatsReply, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: GetStats")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	uuid := uuid.New()
	log.Infof("Get Stats: %s", req.Name)

	err := s.manageSend(req.Name, 0, avoid.ConnectionMessage_SERVING, avoid.ConnectionMessage_STATISTICS, nil, uuid.String())
	if err != nil {
		// TODO
		log.Errorf("")
		return nil, err
	}

	msg, err := s.manageRecv(req.Name, uuid.String(), avoid.ConnectionMessage_SERVING)
	if err != nil {
		// TODO
		log.Errorf("")
		return nil, err
	}

	return &avoid.StatsReply{Stats: msg.Stats}, nil
}

func (s *TunnelServer) Migrate(ctx context.Context, req *avoid.MigrateRequest) (*avoid.MigrateReply, error) {
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

	err := s.manageSend(req.Name,
		req.Migrate.Connection.Connection,
		avoid.ConnectionMessage_SERVING,
		avoid.ConnectionMessage_MIGRATE,
		req.Migrate.Value,
		uuid.String(),
	)
	if err != nil {
		// TODO
		log.Errorf("")
		return nil, err
	}

	msg, err := s.manageRecv(req.Name, uuid.String(), avoid.ConnectionMessage_SERVING)
	if err != nil {
		// TODO
		log.Errorf("")
		return nil, err
	}

	return &avoid.MigrateReply{Migrate: msg.Stats}, nil
}

func (s *TunnelServer) Disconnect(ctx context.Context, req *avoid.DisconnectRequest) (*avoid.DisconnectReply, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: Disconnect")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	uuid := uuid.New()
	log.Infof("Disconnect: %s: %v", uuid.String(), req)

	err := s.manageSend(req.Name,
		0,
		avoid.ConnectionMessage_SERVING,
		avoid.ConnectionMessage_MIGRATE,
		req.Disconnect.Value,
		uuid.String(),
	)
	if err != nil {
		// TODO
		log.Errorf("")
		return nil, err
	}

	msg, err := s.manageRecv(req.Name, uuid.String(), avoid.ConnectionMessage_SERVING)
	if err != nil {
		// TODO
		log.Errorf("")
		return nil, err
	}

	return &avoid.DisconnectReply{Disconnect: msg.Stats}, nil
}

func (s *TunnelServer) Register(ctx context.Context, req *avoid.RegisterRequest) (*avoid.RegisterReply, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: Register")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	log.Infof("Register request: %v\n", req)

	// TODO: Register

	return &avoid.RegisterReply{Token: req.Req}, nil
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
func (s *TunnelServer) Watch(stream avoid.Tunnel_WatchServer) error {

	// When we get an update in the channel if its a send, we need to send to the other side

	// If its on the recv, we need to do something with it.

	for {
		select {
		// we've received a notification, lets check
		case ueUuid := <-s.notifications:
			s.mu.Lock()
			val, ok := s.updates[service]
			s.mu.Unlock()
			break

		}
	}

	/*
		var recv *avoid.ConnectionRequest
		var err error
		for {
			recv, err = stream.Recv()
			if err == io.EOF {
				continue
			}
			if err != nil {
				log.Errorf("err with stream: %v\n", err)
				continue
			}
			break
		}

		log.Infof("Watch: %s\n", recv.Name)
		service := recv.Name

		s.mu.Lock()
		val, ok := s.updates[service]
		s.mu.Unlock()
		if !ok {
			tu := &TunnelUpdate{
				Watch:       &stream,
				SendChannel: make(chan avoid.ConnectionMessage_ServingStatus, 1),
				Current:     avoid.ConnectionMessage_SERVICE_UNKNOWN,
				Action: &avoid.ConnectionReply{
					Connection: &avoid.ConnectionMessage{
						Status: avoid.ConnectionMessage_SERVICE_UNKNOWN,
					},
				},
			}
			s.mu.Lock()
			s.updates[service] = tu
			log.Infof("Added new: %#v\n", s.updates[service])
			s.updates[service].SendChannel <- avoid.ConnectionMessage_SERVICE_UNKNOWN
			s.mu.Unlock()
			for {
				select {
				case servingStatus := <-s.updates[service].SendChannel:
					log.Infof("Something in channel: %s\n", servingStatus)
					if servingStatus == avoid.ConnectionMessage_SERVICE_UNKNOWN {
						log.Infof("Nothing sent\n")
						continue
					}

					s.mu.Lock()
					tu := s.updates[service]
					log.Infof("Data: %#v\n", tu)
					if tu.Watch != nil {
						(*tu.Watch).Send(tu.Action)
					} else {
						log.Errorf("Invalid tunnel update. Watch is nil: %v\n", s.updates[service].Action)
					}
					s.mu.Unlock()
				}
			}
		} else {
			stream.Send(val.Action)
		}
	*/

	return nil
}

func (s *TunnelServer) Watcher(service string) error {
	return nil
}

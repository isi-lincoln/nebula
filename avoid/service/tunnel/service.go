package tunnel

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/slackhq/nebula/avoid"
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

/*
	tu := s.updates[ueUuid]
	tu.Transition = status
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
*/

func (s *AvoidManager) GetStats(ctx context.Context, req *avoid.StatsRequest) (*avoid.StatsReply, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: GetStats")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	uuid := uuid.New()
	log.Infof("Get Stats: %s", req.Name)

	// TODO: Return Stats

	return &avoid.StatsReply{Stats: msg.Stats}, nil
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

	am := &avoid.ActionMessage{}

	// TODO: write am to

	return &avoid.MigrateReply{Migrate: msg.Stats}, nil
}

func (s *AvoidManager) Disconnect(ctx context.Context, req *avoid.DisconnectRequest) (*avoid.DisconnectReply, error) {
	if req == nil {
		errMsg := fmt.Sprintf("Invalid Request: Disconnect")
		log.Errorf("%s", errMsg)
		return nil, fmt.Errorf("%s", errMsg)
	}

	uuid := uuid.New()
	log.Infof("Disconnect: %s: %v", uuid.String(), req)

	return &avoid.DisconnectReply{Disconnect: msg.Stats}, nil
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

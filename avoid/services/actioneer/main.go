package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/slackhq/nebula/avoid"
	"gitlab.com/mergetb/tech/stor"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	cfgPath string
	timeout = 5 * time.Second
	Build   string
)

type Runner struct {
	Host    string
	Pid     int
	Uuid    string
	LeaseId string
	Keys    []string
	Version int64
}

func (x *Runner) Key() string {
	return fmt.Sprintf("%s/%s", avoid.RunnerPrefix, x.Uuid)
}
func (x *Runner) SetVersion(v int64) { x.Version = v }
func (x *Runner) GetVersion() int64  { return x.Version }
func (x *Runner) Value() interface{} { return x }

func actionFunc() {
	actionID := uuid.New().String()
	timeout = 5 * time.Second

	host, err := os.Hostname()
	if err != nil {
		log.Fatal("failure to get hostname: %v", err)
	}

	ar := &Runner{
		Uuid: actionID,
		Host: host,
		Pid:  os.Getpid(),
	}

	// TODO: handle multiple runners at once via leases

	/*
		ctx, cancel := context.WithTimeout(context.TODO(), timeout)
		lease, err := c.Grant(ctx, 5)
		defer cancel()
		if err != nil {
			log.Fatal("failure to get lease: %v", err)
		}

		ar.LeaseId = lease.ID
		stor.Write(ar, clientv3.WithLease(lease.ID))
	*/

	stor.WriteObjects([]stor.Object{ar}, true)
	fields := log.Fields{"runner": ar}

	key := fmt.Sprintf("%s", avoid.PendingPrefix)

	err = stor.WithEtcd(func(etcdp *clientv3.Client) error {
		rch := (*etcdp).Watch(context.Background(), key, clientv3.WithPrefix())
		log.Debugf("begin watch on key: %s", key)
		for wresp := range rch {

			for _, x := range wresp.Events {
				if x.Type == mvccpb.DELETE {
					log.Debugf("delete event for key: %s", x.Kv.Key)
					continue
				}
				pending := &avoid.Pending{}
				err := json.Unmarshal(x.Kv.Value, pending)
				if err != nil {
					fields := log.Fields{"key": x.Kv.Key, "value": x.Kv.Value}
					avoid.ErrorEF("unmarshalling pending failed", err, fields)
				}

				// we've gotten an event, we need to handle it now.
				err = avoid.RUC(pending, func(o stor.Object) {
					o.(*avoid.Pending).Owner = actionID
				})
				// unable to get key
				if err != nil {
					fields["txn key"] = x.Kv.Value
					avoid.ErrorEF("unable to read update commit pending", err, fields)
					continue
				}

				// have key, now need to do something
				fields["actionKey"] = pending.ActionKey

				// TODO: fix this when leases are implemented
				leaseID := 12

				err = actionHandler(leaseID, actionID, pending.ActionKey)
				if err != nil {
					fields["actionError"] = err
					avoid.ErrorEF("action handler failed", err, fields)
					errCount := pending.ErrCount
					if errCount >= 2 {
						avoid.ErrorF("unable to resolveerrors moving to failed", fields)
						// delete pending
						// add fail
						// TODO
					} else {
						avoid.ErrorF("incrementing failed count", fields)
						err = avoid.RUC(pending, func(o stor.Object) {
							o.(*avoid.Pending).Owner = ""
							o.(*avoid.Pending).ErrCount = o.(*avoid.Pending).ErrCount + 1
						})
					}
				}

				log.WithFields(fields).Info("handled action")
			}
		}

		return nil
	})

	log.Fatalf("action failed: %v", err)
}

func actionHandler(lease int, aid, actionKey string) error {
	key := strings.TrimLeft(actionKey, fmt.Sprintf("%s/", avoid.ActionPrefix))
	fields := log.Fields{"key": key}
	ak := &avoid.ActionRequest{
		Uuid: key,
	}
	// read the action
	err := avoid.ReadStandard(ak)
	if err != nil {
		return avoid.ErrorEF("failed action obj read", err, fields)
	}
	// TODO: validate action request

	newAk := ak.Action
	newAk.Uuid = ""

	reg := &avoid.Registration{
		UE: ak.Identifier,
	}

	fields["ue"] = ak.Identifier

	// get the ue details to connect to it
	err = avoid.ReadStandard(reg)
	if err != nil {
		return avoid.ErrorEF("failed registration obj read", err, fields)
	}
	// TODO validate registration
	fields["token"] = reg.Token
	fields["addr"] = fmt.Sprintf("%s:%d", reg.UE, reg.Port)

	ar := &avoid.ActionRequest{
		Identifier: ak.Identifier,
		Values:     ak.Values,
		Token:      reg.Token,
		Action:     newAk,
	}

	ueAddr := fmt.Sprintf("%s:%s", reg.UE, reg.Port)

	// TODO: TLS
	connInfo := &avoid.ConnectionInfo{}
	err = avoid.WithAvoidClient(ueAddr, nil, func(c avoid.AvoidClientClient) error {
		// TODO: add context timeout
		resp, err := c.Action(context.TODO(), ar)
		if err != nil {
			return avoid.ErrorEF("failed client conn", err, fields)
		}

		connInfo = resp
		return nil
	})
	if err != nil {
		return avoid.ErrorEF("failed action on client", err, fields)
	}

	pending := &avoid.Pending{ActionKey: actionKey}

	// remove pending, action, and save results in connections
	tx := avoid.ObjectTx{
		Put:    []stor.Object{connInfo},
		Delete: []stor.Object{pending, ak},
	}

	err = avoid.RunObjectTx(tx)
	if err != nil {
		return avoid.ErrorEF("failed to txn action handler", err, fields)
	}

	return nil
}

func manageActions(actioneers int) {

	for i := 0; i < actioneers; i++ {
		go actionFunc()
	}

	for {
		// TODO: goroutine management
	}
}

func main() {
	printVersion := flag.Bool("version", false, "Print version")
	printUsage := flag.Bool("help", false, "Print command line usage")
	debug := flag.Bool("debug", false, "enable extra debugging")
	avoidConf := flag.String("conf", avoid.DefaultAvoidConfigPath, "set avoid configuration path")
	actioneers := flag.Int("actions", 1, "set actioneer threads")

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

	manageActions(*actioneers)
}

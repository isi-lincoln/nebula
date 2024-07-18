package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/slackhq/nebula/avoid"
	"gitlab.com/mergetb/tech/stor"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	cfgPath    string
	etcd       *clientv3.Client
	actioneers int
	timeout    = 5 * time.Second
)

type Runner struct {
	Host    string
	Pid     string
	Uuid    string
	LeaseId string
	Keys    []string
}

func (x *Runner) Key() string {
	return fmt.Sprintf("%s/%s/%s", avoid.RunnerPrefix, x.Uuid)
}
func (x *Runner) SetVersion(v int64) { x.Version = v }
func (x *Runner) Value() interface{} { return x }

func actionFunc() {
	actionid = uuid.Must(uuid.NewV4()).String()
	timeout = 5 * time.Second

	host, err := os.Hostname()
	if err != nil {
		return -1, err
	}

	ar := &Runner{
		Uuid: actionid,
		Host: host,
		Pid:  os.Getpid(),
	}

	ctx, cancel := context.WithTimeout(context.TODO(), timeout)
	lease, err := c.Grant(ctx, 5)
	defer cancel()
	if err != nil {
		return -1, err
	}

	ar.LeaseId = lease.ID

	stor.Write(ar, clientv3.WithLease(lease.ID))
	fields := log.Fields{"runner": ar}

	key := fmt.Sprintf("%s", avoid.PendingPrefix)
	rch := (*etcd).Watch(context.Background(), key, clientv3.WithPrefix())
	log.Debugf("begin watch on key: %s", key)
	for wresp := range rch {
		avoid.EnsureEtcd(*etcd)

		for _, x := range wresp.Events {
			if x.Type == mvccpb.DELETE {
				log.Debugf("delete event for key: %s", x.Kv.Key)
				continue
			}
			pending := &avoid.Pending{}
			err := json.Unmarshal(x.Kv.Value, pending)
			if err != nil {
				fields := log.Fields{}
				avoid.ErrorEF(fields, err)
			}

			// we've gotten an event, we need to handle it now.
			err = avoid.RUC(s, func(o avoid.Object) {
				o.(*avoid.Pending).Lease = actionId
			})
			// unable to get key
			if err != nil {
				fields["txn key"] = x.Kv.Value
				avoid.ErrorEF(fields, err)
				continue
			}

			// have key, now need to do something
			fields["actionKey"] = pending.ActionKey
			err = actionHandler(lease.ID, actionId, pending.ActionKey)
			if err != nil {
				fields["actionError"] = err
				avoid.ErrorEF(fields, err)
				errCount = o.(*avoid.Pending).ErrCount
				if errCount >= 2 {
					avoid.ErrorF(fields, "unable to resolveerrors moving to failed")
					// delete pending
					// add fail
				} else {
					avoid.ErrorF(fields, "incrementing failed count")
					err = avoid.RUC(s, func(o avoid.Object) {
						o.(*avoid.Pending).Lease = ""
						o.(*avoid.Pending).ErrCount = o.(*avoid.Pending).ErrCount + 1
					})
				}
			}

			avoid.LogF(fields, "handled action")
		}
	}
}

func actionHandler(lease int, aid int, actionKey string) error {
	key := strings.TrimLeft(actionKey, fmt.Sprintf("%s/", avoid.ActionPrefix))
	fields := log.Fields{"key": key}
	ak := &avoid.ActionRequest{
		Uuid: key,
	}
	// read the action
	err = avoid.Read(ak)
	if err != nil {
		return avoid.ErrorEF("failed action obj read", err, fields)
	}
	// TODO: validate action request

	newAk = ak.Action
	newAk.Uuid = 0

	reg := &avoid.Registration{
		UE: ak.Identifier,
	}

	fields["ue"] = ak.Identifier

	// get the ue details to connect to it
	err = avoid.Read(reg)
	if err != nil {
		return avoid.ErrorE("failed registration obj read", err, fields)
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

	connInfo := &avoid.ConnectionInfo{}
	err := avoid.WithAvoidClient(ueAddr, func(c avoid.TunnelClient) error {
		// TODO: add context timeout
		resp, err := c.Action(context.TODO(), ar)
		if err != nil {
			return avoid.ErrorEF("failed client conn", err, fields)
		}
		connInfo = resp
	})
	if err != nil {
		return avoid.ErrorEF("failed action on client", err, fields)
	}

	pending := &avoid.Pending{ActionKey: actionKey}

	// remove pending, action, and save results in connections
	tx := common.ObjectTx{
		Put:    []common.Object{connInfo},
		Delete: []common.Object{pending, ak},
	}

	err = common.RunObjectTx(tx)
	if err != nil {
		return avoid.ErrorEF("failed to txn action handler", err)
	}

	return nil
}

func manageActions() {

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
	cfgPath = flag.String("config", avoid.AvoidConfigPath, "set avoid configuration path")
	actioneers = flag.Int("actions", 1, "set actioneer threads")

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

	cfg, err := avoid.LoadConfig(cfgPath)
	if err != nil {
		log.Fatalf("%v", err)
	}

	// read in environment variables for container
	err = avoid.ReadENVSettings(cfg)
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
	log.Trace("connected to etcd")

	manageActions()
}

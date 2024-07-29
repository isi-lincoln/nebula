package avoid

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"gitlab.com/mergetb/tech/stor"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	MaxMessageSize = 1024 * 1024 * 4
)

const (
	rucRetry = 10
)

const (
	Unspecified  = iota
	ErrNotFound  = iota
	TxnFailedNum = iota
	ConnFailed   = iota
)

var (
	ObjNotFound  = &ObjectError{Code: ErrNotFound, Message: "Object not Found"}
	TxnFailed    = &ObjectError{Code: TxnFailedNum, Message: "Transaction Failed"}
	ClientFailed = &ObjectError{Code: ConnFailed, Message: "Transaction Failed"}
)

type ObjectError struct {
	Code    int
	Message string
}

func (e *ObjectError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

func (e *ObjectError) ToError() error {
	return fmt.Errorf("%d: %s", e.Code, e.Message)
}

func ReadStandard(obj stor.Object) error {
	err := Read(obj)
	if err != nil {
		return err.ToError()
	}
	return nil
}

// Read reads an object from the datastore
func Read(obj stor.Object) *ObjectError {
	n, err := ReadObjects([]stor.Object{obj})
	if err != nil {
		return err
	}
	if n == 0 {
		return ObjNotFound
	}

	return nil
}

// ReadNew reads an object form the datastore, and does not throw an error if
// the object is not found.
func ReadNew(obj stor.Object) error {

	eo := Read(obj)
	if eo != nil && eo.Code != ErrNotFound {
		return fmt.Errorf("%s", eo.Message)
	}

	return nil
}

// ReadTimer tracks timing information about a read.
type ReadTimer struct {
	Period  time.Duration
	Timeout time.Duration
}

// ReadWait attempts to read an object repeatedly until a timeout threshold is
// reached defined by timer. If timer is nil the defaults of 30 seconds with a
// retry period of 250 milliseconds is applied
func ReadWait(obj stor.Object, timer *ReadTimer) error {

	if timer == nil {
		timer = &ReadTimer{
			Period:  250 * time.Millisecond,
			Timeout: 30 * time.Second,
		}
	}

	count := 0
	var eo *ObjectError
	for eo = Read(obj); eo.Code == ErrNotFound; eo = Read(obj) {
		time.Sleep(timer.Period)
		count++
		if time.Duration(count)*timer.Period > timer.Timeout {
			break
		}
	}

	return eo.ToError()
}

// ReadWaitObjects is readwait for many objects
// we check that the number of reads is equal to the number
// of objects (which is what ErrNotFound was doing anyway
func ReadWaitObjects(objs []stor.Object, timer *ReadTimer) error {
	if timer == nil {
		timer = &ReadTimer{
			Period:  250 * time.Millisecond,
			Timeout: 30 * time.Second,
		}
	}

	count := 0
	n, err := ReadObjects(objs)

	for n != len(objs) {
		time.Sleep(timer.Period)
		count++
		if time.Duration(count)*timer.Period > timer.Timeout {
			break
		}
		n, err = ReadObjects(objs)
	}

	return err.ToError()
}

// FromJSON reads reads on object from byte array encoded json. If the object is
// a protobuf, then protobuf is used instead.
func FromJSON(o stor.Object, b []byte) {

	// if this is a protobuf, unmarshal as such
	msg, ok := o.Value().(proto.Message)
	if ok {
		log.Trace("from: as protobuf")
		err := jsonpb.Unmarshal(bytes.NewReader(b), msg)
		if err == nil {
			return
		}
		WarnEF("fail to unmarshal once", err, log.Fields{"object": b})
		//fallthrough to json
	}

	err := json.Unmarshal(b, o.Value())
	if err != nil {
		panic(err) //all merge objects must be json serializable, the end
	}

}

// ReadObjects reads a set of objects from the datastore in a one-shot
// transaction.
func ReadObjects(objs []stor.Object) (int, *ObjectError) {

	var ops []clientv3.Op
	omap := make(map[string]stor.Object)

	names := make([]string, 0)
	for _, o := range objs {
		names = append(names, o.Key())
		omap[o.Key()] = o
		ops = append(ops, clientv3.OpGet(o.Key()))
	}

	n := 0
	objsize := 0
	code := 0
	err := stor.WithEtcd(func(c *clientv3.Client) error {

		kvc := clientv3.NewKV(c)
		ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Minute)
		resp, err := kvc.Txn(ctx).Then(ops...).Commit()
		cancel()
		if err != nil {
			code = ConnFailed
			return err
		}
		if !resp.Succeeded {
			code = TxnFailedNum
			return fmt.Errorf("")
		}

		for _, r := range resp.Responses {
			rr := r.GetResponseRange()
			if rr == nil {
				continue
			}

			for _, kv := range rr.Kvs {
				n++
				o := omap[string(kv.Key)]

				switch t := o.Value().(type) {
				case *string:
					*t = string(kv.Value)
				default:
					FromJSON(o, kv.Value)
					o.SetVersion(kv.Version)
				}
				objsize += len([]byte(kv.Value))
			}
		}

		return nil

	})
	if err != nil {
		return 0, &ObjectError{Code: code, Message: fmt.Sprintf("failed to read: %v", err)}
	}

	return n, nil
}

// ToJSON marshals an object to JSON form. If the object is a protobuf, protobuf
// is used instead.
func ToJSON(o stor.Object) string {

	// if this is a protobuf, marshal as such
	msg, ok := o.Value().(proto.Message)
	if ok {
		log.Trace("from: as protobuf")
		var buf bytes.Buffer
		m := jsonpb.Marshaler{}
		err := m.Marshal(&buf, msg)
		if err == nil {
			return buf.String()
		}
		WarnEF("failed to pbmarshal", err, log.Fields{"object": o})
		//fallthrough to json
	}

	buf, err := json.MarshalIndent(o.Value(), "", "  ")
	if err != nil {
		panic(err) //all merge objects must be json serializable, the end
	}
	return string(buf)
}

// Write persists an object to the datastore.
func Write(obj stor.Object, opts ...clientv3.OpOption) error {
	return WriteObjects([]stor.Object{obj}, false, opts...).ToError()
}

// WriteObjects writes objects to the datastore in a single shot transaction. If
// fresh is true, then all objects must be the most recent version, or the write
// will fail.
func WriteObjects(objs []stor.Object, fresh bool, opts ...clientv3.OpOption) *ObjectError {

	var ops []clientv3.Op
	var ifs []clientv3.Cmp

	names := make([]string, 0)
	objsize := 0
	for _, obj := range objs {
		names = append(names, obj.Key())

		var value string
		switch t := obj.Value().(type) {
		case *string:
			value = *t
		default:
			value = ToJSON(obj)
		}
		objsize += len([]byte(value))

		ops = append(ops, clientv3.OpPut(obj.Key(), value, opts...))
		if fresh {
			ifs = append(ifs,
				clientv3.Compare(clientv3.Version(obj.Key()), "=", obj.GetVersion()))
		}

	}

	log.Tracef("Write (%v) Size: %d", names, objsize)

	code := 0
	err := stor.WithEtcd(func(c *clientv3.Client) error {
		kvc := clientv3.NewKV(c)
		if kvc == nil {
			log.Error("failed to create clientv3 client")
			return fmt.Errorf("failed to create clientv3 client")
		}

		ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Minute)
		resp, err := kvc.Txn(ctx).If(ifs...).Then(ops...).Commit()
		cancel()
		if err != nil {
			return err
		}
		if !resp.Succeeded {
			code = TxnFailedNum
			return fmt.Errorf("state has changed since read")
		}

		for _, o := range objs {
			o.SetVersion(o.GetVersion() + 1)
		}
		return nil
	})

	return &ObjectError{Code: code, Message: fmt.Sprintf("%v", err)}
}

// Touch update the key, but not value in data store
func Touch(obj stor.Object) error {
	return TouchObjects([]stor.Object{obj})
}

// TouchObjects updates multiple keys
func TouchObjects(objs []stor.Object) error {

	var ops []clientv3.Op
	var ifs []clientv3.Cmp

	names := make([]string, 0)
	objsize := 0
	for _, obj := range objs {
		names = append(names, obj.Key())

		ops = append(ops, clientv3.OpPut(obj.Key(), "", clientv3.WithIgnoreValue()))
	}

	log.Tracef("Write (%v) Size: %d", names, objsize)

	code := 0
	err := stor.WithEtcd(func(c *clientv3.Client) error {

		kvc := clientv3.NewKV(c)
		if kvc == nil {
			log.Error("failed to create clientv3 client")
			return fmt.Errorf("failed to create clientv3 client")
		}

		ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Minute)
		resp, err := kvc.Txn(ctx).If(ifs...).Then(ops...).Commit()
		cancel()
		if err != nil {
			code = ConnFailed
			return err
		}
		if !resp.Succeeded {
			code = TxnFailedNum
			return fmt.Errorf("state has changed since read")
		}
		return nil
	})

	if err != nil {
		eo := &ObjectError{Code: code, Message: fmt.Sprintf("%v", err)}
		return eo.ToError()
	}

	return nil
}

// DeleteObjects deletes a set of objects from the datastore.
func DeleteObjects(objs []stor.Object) error {

	var ops []clientv3.Op

	for _, obj := range objs {
		ops = append(ops, clientv3.OpDelete(obj.Key()))
	}

	code := 0
	err := stor.WithEtcd(func(c *clientv3.Client) error {
		kvc := clientv3.NewKV(c)
		ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Minute)
		resp, err := kvc.Txn(ctx).Then(ops...).Commit()
		cancel()
		if err != nil {
			return err
		}
		if !resp.Succeeded {
			code = TxnFailedNum
			return fmt.Errorf("delete objects failed")
		}
		return nil
	})

	if err != nil {
		eo := &ObjectError{Code: code, Message: fmt.Sprintf("%v", err)}
		return eo.ToError()
	}

	return nil
}

// Delete deletes an object from the datastore.
func Delete(obj stor.Object) error {

	return DeleteObjects([]stor.Object{obj})

}

// ObjectTx encapsulates a set of put and delete operations into a single
// transaction.
type ObjectTx struct {
	Put    []stor.Object
	Delete []stor.Object
}

// RunObjectTx runs an object transaction.
func RunObjectTx(otx ObjectTx) error {

	var ops []clientv3.Op

	names := make([]string, 0)
	objsize := 0
	for _, x := range otx.Put {
		names = append(names, x.Key())

		var value string
		switch t := x.Value().(type) {
		case *string:
			value = *t
		default:
			buf, err := json.MarshalIndent(x.Value(), "", "  ")
			if err != nil {
				return err
			}
			value = string(buf)
		}
		objsize += len([]byte(value))

		ops = append(ops, clientv3.OpPut(x.Key(), string(value)))
	}

	log.Tracef("WriteTx (%v) Size: %d", names, objsize)
	names = make([]string, 0)
	for _, x := range otx.Delete {
		names = append(names, x.Key())
		ops = append(ops, clientv3.OpDelete(x.Key()))
	}
	log.Tracef("DeleteTx: (%v)", names)

	code := 0
	err := stor.WithEtcd(func(c *clientv3.Client) error {
		kvc := clientv3.NewKV(c)
		ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Minute)
		resp, err := kvc.Txn(ctx).Then(ops...).Commit()
		cancel()
		if err != nil {
			return err
		}
		if !resp.Succeeded {
			code = TxnFailedNum
			return fmt.Errorf("run object txn failed")
		}
		return nil
	})

	if err != nil {
		eo := &ObjectError{Code: code, Message: fmt.Sprintf("%v", err)}
		return eo.ToError()
	}

	return nil
}

// RunObjectTxPrefix a bastaradization of RunObjectTx, making it so
// put is still object array, but delete is a prefix for the txn
func RunObjectTxPrefix(puts []stor.Object, deletePrefix string) error {

	if deletePrefix == "" {
		return fmt.Errorf("attempted to txn delete db")
	}

	var ops []clientv3.Op
	names := make([]string, 0)
	objsize := 0
	// put as in RunObjectTx
	for _, x := range puts {
		names = append(names, x.Key())
		var value string
		switch t := x.Value().(type) {
		case *string:
			value = *t
		default:
			buf, err := json.MarshalIndent(x.Value(), "", "  ")
			if err != nil {
				return err
			}
			value = string(buf)
		}
		objsize += len([]byte(value))
		ops = append(ops, clientv3.OpPut(x.Key(), string(value)))
	}

	// delete on prefix
	ops = append(ops, clientv3.OpDelete(deletePrefix, clientv3.WithPrefix()))
	log.Tracef("PutTx: (%v)", names)
	log.Tracef("DeletePrefixTx: (%v)", deletePrefix)

	code := 0
	err := stor.WithEtcd(func(c *clientv3.Client) error {
		kvc := clientv3.NewKV(c)
		ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Minute)
		resp, err := kvc.Txn(ctx).Then(ops...).Commit()
		cancel()
		if err != nil {
			return err
		}
		if !resp.Succeeded {
			code = TxnFailedNum
			return fmt.Errorf("run object prefix txn failed")
		}
		return nil
	})

	if err != nil {
		eo := &ObjectError{Code: code, Message: fmt.Sprintf("%v", err)}
		return eo.ToError()
	}

	return nil
}

// ReadRevision reads an object, and returns the clientv3 key revision for that object
func ReadRevision(obj stor.Object) (revision int64, err error) {
	objsize := 0
	revision = 0
	err = stor.WithEtcd(func(c *clientv3.Client) error {
		kvc := clientv3.NewKV(c)
		ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Minute)
		resp, err := kvc.Get(ctx, obj.Key())
		cancel()
		if err != nil {
			return err
		}
		if len(resp.Kvs) > 0 {
			kv := resp.Kvs[0]

			FromJSON(obj, kv.Value)
			obj.SetVersion(kv.Version)
			objsize += len([]byte(kv.Value))
			revision = resp.Header.Revision
		} else {
			return fmt.Errorf("%s", ErrNotFound)
		}
		return nil
	})
	if err != nil {
		return 0, ErrorE("failed to read", err)
	}
	log.Tracef("ReadRev (%v) Size: %d", obj.Key(), objsize)

	return revision, err
}

// RUC performs a read-update-commit on the provided object using the specified
// update function.
func RUC(o stor.Object, update func(o stor.Object)) error {
	for i := 0; i < rucRetry; i++ {
		update(o)

		eo := WriteObjects([]stor.Object{o}, false)
		if eo == nil {
			return nil
		}
		if eo.Code == TxnFailedNum {
			eo = Read(o)
			if eo != nil {
				return eo.ToError()
			}
			continue

		}

		return eo.ToError()
	}

	eo := &ObjectError{Code: TxnFailedNum, Message: "unable to complete transaction"}
	return eo.ToError()
}

// GetEtcdConfig sets the global etcd configuration settings
func GetEtcdConfig(cfg *ServicesConfig) (*EtcdConfig, error) {
	if cfg == nil {
		return nil, fmt.Errorf("Service configuration is undefined")
	}
	if cfg.Etcd == nil {
		return nil, fmt.Errorf("etcd service not found in config with key: etcd")
	}
	return cfg.Etcd, nil
}

func SetConfig(cfg *EtcdConfig) error {
	if cfg == nil {
		return fmt.Errorf("attempt to load etcd configuration is nil")
	}
	//etcdConfig = cfg
	scfg := stor.Config{
		Address: cfg.Address,
		Port:    cfg.Port,
		TLS:     cfg.TLS,
		Quantum: cfg.Quantum,
		Timeout: cfg.Timeout,
	}

	stor.SetConfig(scfg)

	return nil
}

/*

// EtcdConnect Try to get a etcd client- assumption EtcdClient is async until used
func EtcdConnect() (*clientv3.Client, error) {
	log.Trace("connecting to etcd...")

	etcd, err := EtcdClient()
	if err == nil {
		return etcd, nil
	}
	return nil, ErrorE("failed to connect to etcd", err)
}

// EnsureEtcd Make sure we always have an etcd connection
func EnsureEtcd(etcdp **clientv3.Client) error {

	// ensure we have a usable etcd connection
	var err error
	if *etcdp == nil {
		log.Debugf("etcd connection nil - connecting")
		*etcdp, err = EtcdConnect()
		if err != nil {
			return err
		}
	}

	// experimental: https://github.com/grpc/grpc-go/pull/1430
	state := (**etcdp).ActiveConnection().GetState()

	//https://github.com/grpc/grpc/blob/master/doc/connectivity-semantics-and-api.md
	if state != connectivity.Ready && state != connectivity.Idle {

		WarnE("etcd status check error - reconnecting", err)

		(*etcdp).Close()
		*etcdp, err = EtcdConnect()
		if err != nil {
			return err
		}

		state := (**etcdp).ActiveConnection().GetState()

		if state != connectivity.Ready && state != connectivity.Idle {
			return ErrorE("etcd status check failed - giving up", err)
		}
	}

	return err

}

// WithEtcd executes a function against an etcd client with a managed
// connection lifetime.
func WithEtcd(f func(*clientv3.Client) error) error {

	cli, err := EtcdConnect()
	if err != nil {
		return err
	}
	defer cli.Close()

	return f(cli)

}

// EtcdClient read the specific configuration files to initiate etcd connection
func EtcdClient() (*clientv3.Client, error) {

	log.Trace("creating new client...")

	cfg := etcdConfig

	var tlsc *tls.Config
	if cfg.TLS != nil {

		log.Trace("etcd tls enabled")

		log.WithFields(log.Fields{
			"cacert": cfg.TLS.Cacert,
			"cert":   cfg.TLS.Cert,
			"key":    cfg.TLS.Key,
		}).Trace("tls config")

		capool := x509.NewCertPool()
		capem, err := ioutil.ReadFile(cfg.TLS.Cacert)
		if err != nil {
			return nil, ErrorE("failed to read cacert", err)
		}
		ok := capool.AppendCertsFromPEM(capem)
		if !ok {
			return nil, ErrorF("capem is not ok", log.Fields{"ok": ok})
		}

		cert, err := tls.LoadX509KeyPair(
			cfg.TLS.Cert,
			cfg.TLS.Key,
		)
		if err != nil {
			return nil, ErrorE("failed to load cert/key pair", err)
		}

		tlsc = &tls.Config{
			RootCAs:      capool,
			Certificates: []tls.Certificate{cert},
		}
	} else {
		log.Trace("etcd tls disabled")
	}

	connstr := fmt.Sprintf("%s:%d", cfg.Address, cfg.Port)
	log.WithFields(log.Fields{
		"connstr": connstr,
	}).Trace("etcd connection string")

	// The issue here with etcd is that the connection to the database is
	// sticking around for 2MSL (2 minutes), so we will run into issues with
	// max number of connections.
	// So we will pass a dialoption, with a dialler, that overwrites the
	// standard tcp connection with SO_LINGER.  Setting to 0 deletes immediately,
	// > 1 is seconds, < 0 is backgrounded. non-zero leaves time_wait for 2MSL
	//TODO: There should be a better way of tracking down why the connection isnt
	//closing correctly.
	f := func(ctx context.Context, addr string) (net.Conn, error) {
		dialer := &net.Dialer{
			Deadline: time.Now().Add(1 * time.Minute),
		}
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			return nil, err
		}
		//http://www.serverframework.com/asynchronousevents/2011/01/
		//time-wait-and-its-design-implications-for-protocols-and-
		//scalable-servers.html
		//conn.(*net.TCPConn).SetLinger(0)
		err = conn.(*net.TCPConn).SetKeepAlive(true)
		if err != nil {
			log.Warn(err)
		}
		err = conn.(*net.TCPConn).SetKeepAlivePeriod(1 * time.Second)
		if err != nil {
			log.Warn(err)
		}
		return conn, err
	}

	opts := []grpc.DialOption{
		grpc.WithContextDialer(f),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(MaxMessageSize),
			grpc.MaxCallSendMsgSize(MaxMessageSize),
		),
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:            []string{connstr},
		DialTimeout:          3 * time.Second,
		DialKeepAliveTime:    -1 * time.Second,
		DialKeepAliveTimeout: -1 * time.Second,
		TLS:                  tlsc,
		DialOptions:          opts,
		MaxCallSendMsgSize:   MaxMessageSize,
		MaxCallRecvMsgSize:   MaxMessageSize,
	})

	log.Trace("client created")
	return cli, err

}
*/

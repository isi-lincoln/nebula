
Lighthouse:
```
ETCDCTL_API=3 etcdctl --cacert=/etc/etcd/ca.pem --cert=/etc/etcd/db.pem --key=/etc/etcd/db-key.pem --endpoints=192.168.0.1:2379 get --prefix ""
```

module github.com/slackhq/nebula

go 1.22.0

toolchain go1.22.2

require (
	dario.cat/mergo v1.0.0
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be
	github.com/armon/go-radix v1.0.0
	github.com/cyberdelia/go-metrics-graphite v0.0.0-20161219230853-39f87cc3b432
	github.com/flynn/noise v1.1.0
	github.com/gogo/protobuf v1.3.2
	github.com/google/gopacket v1.1.19
	github.com/kardianos/service v1.2.2
	github.com/miekg/dns v1.1.59
	github.com/prometheus/client_golang v1.11.0
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475
	github.com/sirupsen/logrus v1.9.3
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	github.com/songgao/water v0.0.0-20200317203138-2b4b6d7c09d8
	github.com/spf13/cobra v1.8.0
	github.com/stretchr/testify v1.9.0
	github.com/vishvananda/netlink v1.2.1-beta.2
	golang.org/x/crypto v0.23.0
	golang.org/x/exp v0.0.0-20230725093048-515e97ebf090
	golang.org/x/net v0.25.0
	golang.org/x/sync v0.7.0
	golang.org/x/sys v0.20.0
	golang.org/x/term v0.20.0
	golang.zx2c4.com/wintun v0.0.0-20230126152724-0fa3db229ce2
	golang.zx2c4.com/wireguard/windows v0.5.3
	google.golang.org/grpc v1.64.0
	google.golang.org/protobuf v1.34.1
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/nbrownus/go-metrics-prometheus v0.0.0-20210712211119-974a6260965f // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.26.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	golang.org/x/mod v0.16.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	golang.org/x/tools v0.19.0 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/slackhq/nebula => ./

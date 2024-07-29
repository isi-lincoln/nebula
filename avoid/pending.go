package avoid

import (
	"fmt"
)

var (
	RunnerPrefix       = "/runners"
	PendingPrefix      = "/pending"
	ActionPrefix       = "/actionrequest"
	ConnPrefix         = "/connections"
	FailPrefix         = "/failed"
	RegistrationPrefix = "/registration"

	MaxFailCount = 3
)

// TODO: TOTP
type Registration struct {
	UE          string
	EP          string
	Token       string
	IP          string
	DNS         string
	Certificate string
	Port        int64
	Version     int64
}

func (x *Registration) Key() string {
	return fmt.Sprintf("%s/%s", RegistrationPrefix, x.UE)
}
func (x *Registration) SetVersion(v int64) { x.Version = v }
func (x *Registration) GetVersion() int64  { return x.Version }
func (x *Registration) Value() interface{} { return x }

type Pending struct {
	ActionKey string // action key pointer
	Version   int64  // for stor to manage
	Owner     string // owner
	ErrCount  int    // failures
}

func (x *Pending) Key() string {
	return fmt.Sprintf("%s/%s", PendingPrefix, x.ActionKey)
}
func (x *Pending) SetVersion(v int64) { x.Version = v }
func (x *Pending) GetVersion() int64  { return x.Version }
func (x *Pending) Value() interface{} { return x }

// TODO: guard rails before calling Key()
func (x *ActionRequest) Key() string {
	return fmt.Sprintf("%s/%s", ActionPrefix, x.Action.Uuid)
}
func (x *ActionRequest) SetVersion(v int64) { x.Version = v }
func (x *ActionRequest) Value() interface{} { return x }

func (x *ConnectionInfo) Key() string {
	return fmt.Sprintf("%s/%s", ConnPrefix, x.Uuid)
}
func (x *ConnectionInfo) SetVersion(v int64) { x.Version = v }
func (x *ConnectionInfo) Value() interface{} { return x }

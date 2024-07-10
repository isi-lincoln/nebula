package tunnel

import (
	"fmt"
)

var (
	PendingPrefix = "/pending"
	ActionPrefix  = "/action"
)

type Pending struct {
	UE        string // unique identifier for UE
	NebulaEP  string // nebula endpoint connected to UE
	ActionKey string // action key pointer
}

func (x *Pending) Key() string {
	return fmt.Sprintf("%s/%s/%s", PendingPrefix, x.UE, x.ActionKey)
}
func (x *Pending) SetVersion(v int64) { x.Version = v }
func (x *Pending) Value() interface{} { return x }

func (x *ActionMessage) Key() string {
	return fmt.Sprintf("%s/%s", ActionPrefix, x.Uuid)
}
func (x *ActionMessage) SetVersion(v int64) { x.Version = v }
func (x *ActionMessage) Value() interface{} { return x }

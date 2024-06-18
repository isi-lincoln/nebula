package avoid

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/slackhq/nebula/config"
	"github.com/stretchr/testify/assert"
)

func TestLoading(t *testing.T) {
	l := logrus.New()
	l.Out = os.Stdout
	c := config.NewC(l)
	err := c.Load("configs/hybrid.yaml")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, c.Get("avoid.config"), "./configs/manager.yaml")

	// TODO
	av := NewAvoidFromConfig(l, c)
	assert.NotNil(t, av)
	assert.NotNil(t, av.primary)
	assert.Equal(t, av.primary.Address, "192.168.0.10")
	assert.Nil(t, av.backups)

	c.Settings["avoid"].(map[interface{}]interface{})["config"] = "./configs/client.yaml"
	av.reload(c, true)
	assert.NotNil(t, av)
	assert.NotNil(t, av.primary)
	assert.Equal(t, av.primary.Address, "192.168.0.10")
	assert.NotNil(t, av.backups)
	assert.Equal(t, av.backups[0].Address, "192.168.0.11")
	assert.Equal(t, av.identity, "12345")

	logrus.Infof("avoid: %#v\n", av)

}

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

	logrus.Infof("settings: %+v", c.Settings)
	assert.Equal(t, c.Get("avoid.config"), "./configs/manager.yaml")

	// TODO
	av := NewAvoidFromConfig(l, c)
	assert.NotNil(t, av)
	assert.NotNil(t, av.primary)

	c.Settings["avoid"].(map[interface{}]interface{})["config"] = "./configs/client.yaml"
	logrus.Infof("settings: %+v", c.Settings)
	av.reload(c, true)
	assert.NotNil(t, av)
	assert.NotNil(t, av.primary)

	logrus.Infof("avoid: %#v\n", av)

}

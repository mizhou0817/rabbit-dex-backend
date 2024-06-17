package settlement

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	config, err := ReadConfig()
	assert.NoError(t, err)
	logrus.Infof("config: %+v", config)

	// NewSettlementService(
	// 	config.Handlers,
	// 	"1888f5178cbffcb8750354c5bb081fd0efb560171a14a285b28b00b01ad90357",
	// )
}

package eureka

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func mockClient() *Client {
	appName := "eureka-test-app"
	localip := "192.168.210.32"
	port := 12345
	return &Client{
		ServerAddr:  "192.168.118.220",
		ServerPort:  8761,
		Address:     localip,
		Port:        uint64(port),
		ServiceName: appName,
		InstanceID:  fmt.Sprintf("%s:%s:%d", appName, localip, port),
	}
}

func TestRegister(t *testing.T) {
	c := mockClient()
	assert.NoError(t, c.Register())
	assert.NoError(t, c.Deregister())
	time.Sleep(10 * time.Second)
}

// func TestDeRegister(t *testing.T) {
// 	c := mockClient()
// 	assert.NoError(t, c.Deregister())
// }

func TestInstanceAddress(t *testing.T) {
	c := mockClient()
	assert.NoError(t, c.Register())
	defer c.Deregister()
	time.Sleep(time.Second * 2)

	assert.Equal(t, "192.168.210.32:12345", c.InstanceAddress("eureka-test-app"))
}

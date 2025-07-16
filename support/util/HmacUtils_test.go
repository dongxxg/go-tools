package util

import (
	"fmt"
	"testing"
)

func TestHmacSha256(t *testing.T) {
	mes := "{\"eventMs\":1560408533119,\"eventType\":10,\"noticeId\":\"4eb720f0-8da7-11e9-a43e-53f411c2761f\",\"notifyMs\":1560408533119,\"payload\":{\"a\":\"1\",\"b\":2},\"productId\":1}"
	res := HmacSha256("secret", mes)
	fmt.Println(res)
}

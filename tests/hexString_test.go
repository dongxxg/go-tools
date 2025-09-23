/**
  @Author: dongxx
  @Since: 2025/9/23 10:08
  @Desc: //TODO
**/

package tests

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestWriteHexString(t *testing.T) {
	info := strings.Replace("60 00", " ", "", -1)
	data, err := hex.DecodeString(info)
	if err != nil {
		t.Fatalf("hex DecodeString err: %v", err)
	} else {
		t.Log("hex DecodeString:", string(data))
	}

	info1 := strings.Replace("60 00 12 6a 7e", " ", "", -1)
	data1, err := hex.DecodeString(info1)
	if err != nil {
		t.Fatalf("hex DecodeString err: %v", err)
	} else {
		t.Log("hex DecodeString:", string(data1))
	}
}

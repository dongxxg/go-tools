package passwd

import (
	"fmt"
	"testing"
)

func TestGenPasswd(t *testing.T) {
	fmt.Println(GenPasswdNum(6))
	fmt.Println(GenPasswdMix(16))
}

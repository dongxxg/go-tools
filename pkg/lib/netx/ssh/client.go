package ssh

import (
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
	"unitechs.com/unios-dice/uni-base/lib/netx/ip"
)

func NewPasswordClient(address, username, password string, port int, conntimeout time.Duration) (*ssh.Client, error) {
	host := ip.NormalAddress(address, port)
	if host == "" {
		return nil, fmt.Errorf("invalid ip address: %s", address)
	}

	var sshconfig ssh.Config
	// 设置默认算法，包括秘钥交换、加密、校验
	sshconfig.SetDefaults()

	// 加入默认不支持的加密方法，默认支持的不需要再额外列出，合并后总共10个
	ciphers := []string{
		"aes128-cbc",
		"3des-cbc",
		"arcfour256", "arcfour128", "arcfour",
	}
	sshconfig.Ciphers = append(sshconfig.Ciphers, ciphers...)

	// 加入默认不支持的秘钥交换算法，合并后共8个
	keyExchanges := []string{
		"diffie-hellman-group1-sha1",
		"diffie-hellman-group-exchange-sha1",
		"diffie-hellman-group-exchange-sha256",
	}
	sshconfig.KeyExchanges = append(sshconfig.KeyExchanges, keyExchanges...)

	config := ConfigPassword(username, password)
	config.Config = sshconfig
	config.Timeout = conntimeout

	sshClient, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return nil, err
	}

	return sshClient, nil
}

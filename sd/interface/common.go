package sd

import (
	"fmt"

	"github.com/dongxxg/go-tools/common"
	"github.com/dongxxg/go-tools/config"
)

func CreateRegistion() (name, ip string, port uint64, err error) {
	err = nil
	ip = common.InitLocalIPbyUDP()
	if ipFromConf := config.GetString("SD.Host"); ipFromConf != "" {
		ip = ipFromConf
	}

	port = config.GetUint64("Server.Http.Port")
	if port == 0 {
		port = config.GetUint64("Server.Https.Port")
	}
	if port == 0 {
		port = config.GetUint64("Server.gRPC.Port")
	}
	if portFromConf := config.GetInt64("SD.Port"); portFromConf != 0 {
		port = uint64(portFromConf)
	}

	name = config.GetString("Service.Name")
	if nameFromConf := config.GetString("SD.Name"); nameFromConf != "" {
		name = nameFromConf
	}

	if name == "" || ip == "" || port == 0 {
		err = fmt.Errorf("invalid SD register parameters, name: %s, ip: %s, port: %d", name, ip, port)
	}

	return
}
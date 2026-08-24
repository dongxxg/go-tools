package db

import (
	"github.com/go-ini/ini"
	"log"
)

var Cfg *ini.File

func init() {
	var err error
	Cfg, err = ini.Load("app.ini")
	if err != nil {
		log.Fatalf("Fail to parse app.ini: %v", err)
	}
}

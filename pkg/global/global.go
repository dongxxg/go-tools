package global

import (
	"github.com/jinzhu/gorm"
	"go-tools/support/zap/config"
	"go.uber.org/zap"
)

var (
	Server    config.Server
	LOG       *zap.Logger
	JwtSecret string
	DB        *gorm.DB
)

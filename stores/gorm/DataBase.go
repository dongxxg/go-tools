package db

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	dameng "github.com/talentkai/gorm-dameng"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glog "gorm.io/gorm/logger"

	conf "github.com/dongxxg/go-tools/config"
	"github.com/dongxxg/go-tools/log"
)

// configKeyPrefix 数据库配置在 config 模块中的节名
const configKeyPrefix = "Database"

const (
	defaultSlowThresholdMs = 200             // 默认慢SQL阈值(毫秒)
	defaultMaxIdleConns    = 5               // 默认连接池空闲连接数
	defaultMaxOpenConns    = 10              // 默认连接池最大连接数
	poolMonitorInterval    = 5 * time.Minute // 连接池监控间隔
)

type DBType string // 读写分离
const (
	DBType_READ  DBType = "read"
	DBType_WRITE DBType = "write"
)

type DataBase struct {
	Dialect         string        `json:"dialect"`
	Host            string        `json:"host"`
	Port            string        `json:"port"`
	User            string        `json:"user"`
	Password        string        `json:"password"`
	DbName          string        `json:"db"`
	Pool            MysqlPool     `json:"pool"`            // 连接池配置
	SlowThresholdMs int           `json:"slowThresholdMs"` // 慢SQL阈值(毫秒), <=0 使用默认值
	GormLogLevel    glog.LogLevel `json:"gormLogLevel"`    // gorm日志级别 1静默 2错误 3警告 4信息, 未配置时根据 Logging.Level 推导
}

type MysqlPool struct {
	MaxIdleConns int `json:"maxIdleConns"`
	MaxOpenConns int `json:"maxOpenConns"`
}

var rdb *gorm.DB
var wdb *gorm.DB
var dateBaseConfig DataBase

// LoadFromConfig 从 config 模块读取数据库配置, 对应配置文件中的 [Database] 节:
//
//	[Database]
//	Dialect = "mysql"      # mysql / dameng
//	Host = "127.0.0.1"
//	Port = "3306"
//	User = "root"
//	Password = ""
//	DbName = "demo"
//	GormLogLevel = 3       # gorm日志级别 1静默 2错误 3警告 4信息
//	SlowThresholdMs = 200  # 慢SQL阈值(毫秒), <=0 使用默认值
//	[Database.Pool]
//	MaxIdleConns = 5
//	MaxOpenConns = 10
func LoadFromConfig() DataBase {
	var cfg DataBase
	cfg.Dialect = conf.GetString(configKeyPrefix + ".Dialect")
	cfg.Host = conf.GetString(configKeyPrefix + ".Host")
	cfg.Port = conf.GetString(configKeyPrefix + ".Port")
	cfg.User = conf.GetString(configKeyPrefix + ".User")
	cfg.Password = conf.GetString(configKeyPrefix + ".Password")
	cfg.DbName = conf.GetString(configKeyPrefix + ".DbName")
	cfg.Pool.MaxIdleConns = conf.GetInt(configKeyPrefix + ".Pool.MaxIdleConns")
	cfg.Pool.MaxOpenConns = conf.GetInt(configKeyPrefix + ".Pool.MaxOpenConns")
	cfg.GormLogLevel = glog.LogLevel(conf.GetInt(configKeyPrefix + ".GormLogLevel"))
	cfg.SlowThresholdMs = conf.GetInt(configKeyPrefix + ".SlowThresholdMs")

	// 未配置gorm日志级别时, 根据日志模块的级别推导
	if cfg.GormLogLevel == 0 {
		cfg.GormLogLevel = gormLevelFromLogConfig(conf.GetString("Logging.Level"))
	}
	return cfg
}

// InitFromConfig 从 config 模块加载配置并初始化数据库连接
func InitFromConfig() {
	cfg := LoadFromConfig()
	if cfg.Dialect == "" {
		panic(fmt.Sprintf("%s.Dialect is not configured", configKeyPrefix))
	}
	Init(cfg, "", "")
}

func Init(cfg DataBase, cloudProvider, serviceName string) {
	dateBaseConfig = cfg
	dbLogger := newGormLogger(
		normalizeGormLogLevel(cfg.GormLogLevel),
		slowThreshold(cfg),
	)

	var dsn string
	var dialector gorm.Dialector
	switch cfg.Dialect {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
			cfg.User,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.DbName)
		dialector = mysql.Open(dsn)
	case "dameng":
		options := map[string]string{
			"schema": cfg.DbName,
		}
		// dm://user:password@host:port?schema=SYSDBA[&...]
		port, _ := strconv.Atoi(cfg.Port)
		dsn = dameng.BuildUrl(
			cfg.User,
			cfg.Password,
			cfg.Host,
			port,
			options,
		)
		dialector = dameng.Open(dsn)
	default:
		panic(fmt.Sprintf("unsupported database type: %s", cfg.Dialect))
	}

	var err error
	wdb, err = gorm.Open(dialector, &gorm.Config{
		Logger:                 dbLogger,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		panic(fmt.Errorf("database connect error: %w", err))
	}

	//设置连接池
	setPool(wdb)

	// 启动连接池监控
	go monitorConnectionPool(wdb)

	// 不记录密码
	log.Infof("database[%s] connected: %s@%s:%s/%s", cfg.Dialect, cfg.User, cfg.Host, cfg.Port, cfg.DbName)
}

// normalizeGormLogLevel 将非法的gorm日志级别修正为默认值 Warn
func normalizeGormLogLevel(level glog.LogLevel) glog.LogLevel {
	if level < glog.Silent || level > glog.Info {
		return glog.Warn
	}
	return level
}

// gormLevelFromLogConfig 根据日志模块的级别字符串推导gorm日志级别
func gormLevelFromLogConfig(level string) glog.LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG", "TRACE":
		return glog.Info
	case "ERROR", "FATAL", "PANIC":
		return glog.Error
	default:
		// 包括空值和 WARN/WARNING/INFO
		return glog.Warn
	}
}

// slowThreshold 返回慢SQL阈值, 未配置时使用默认值
func slowThreshold(cfg DataBase) time.Duration {
	ms := cfg.SlowThresholdMs
	if ms <= 0 {
		ms = defaultSlowThresholdMs
	}
	return time.Duration(ms) * time.Millisecond
}

func GetConnection(dbType ...DBType) *gorm.DB {
	var db *gorm.DB
	if len(dbType) > 0 && dbType[0] == DBType_READ && rdb != nil {
		db = rdb
	} else {
		db = wdb
	}
	return db
}

func printConnectionID(db *gorm.DB) {
	var connectionID int
	var connectionIDSql string
	if dateBaseConfig.Dialect == "mysql" {
		connectionIDSql = "SELECT CONNECTION_ID();"
	} else if dateBaseConfig.Dialect == "dameng" {
		connectionIDSql = "SELECT SESSID();"
	}
	if connectionIDSql == "" {
		return
	}
	if err := db.Raw(connectionIDSql).Scan(&connectionID).Error; err != nil {
		log.Errorf("failed to get connection id: %v", err)
	} else {
		log.Infof("current connection ID: %d", connectionID)
	}
}

func monitorConnectionPool(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		// 运行在后台goroutine中, 记录错误而不是退出进程
		log.Errorf("failed to get sql.DB from gorm.DB: %v", err)
		return
	}
	ticker := time.NewTicker(poolMonitorInterval)
	defer ticker.Stop()
	for range ticker.C {
		stats := sqlDB.Stats()
		log.Infof(
			"connection pool stats: maxOpen=%d open=%d inUse=%d idle=%d",
			stats.MaxOpenConnections,
			stats.OpenConnections,
			stats.InUse,
			stats.Idle,
		)
	}
}

// getDefaultIfZero 返回一个int，如果提供的n值小于或等于0，则返回默认值d，否则返回n。
func getDefaultIfZero(n int, d int) int {
	if n <= 0 {
		return d
	}
	return n
}

func setPool(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxIdleConns(getDefaultIfZero(dateBaseConfig.Pool.MaxIdleConns, defaultMaxIdleConns))
	sqlDB.SetMaxOpenConns(getDefaultIfZero(dateBaseConfig.Pool.MaxOpenConns, defaultMaxOpenConns))
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)
}

/*
*
支持 外部系统自己定义rdb，然后赋值，使用ginmw下的统一函数功能
*/
func CustomizeExternalDB(rdbInput, wdbInput *gorm.DB) {
	rdb = rdbInput
	wdb = wdbInput
}

func GetSqlxConnection(dbType ...DBType) *sqlx.DB {
	db, _ := GetConnection(dbType...).DB()
	mysqlDB := sqlx.NewDb(db, "mysql")
	return mysqlDB
}

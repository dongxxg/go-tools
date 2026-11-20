package db

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	glog "gorm.io/gorm/logger"

	conf "github.com/dongxxg/go-tools/config"
	"github.com/dongxxg/go-tools/log"
)

func TestNormalizeGormLogLevel(t *testing.T) {
	tests := []struct {
		name  string
		level glog.LogLevel
		want  glog.LogLevel
	}{
		{"unset falls back to warn", 0, glog.Warn},
		{"negative falls back to warn", -1, glog.Warn},
		{"silent kept", glog.Silent, glog.Silent},
		{"error kept", glog.Error, glog.Error},
		{"warn kept", glog.Warn, glog.Warn},
		{"info kept", glog.Info, glog.Info},
		{"out of range falls back to warn", glog.Info + 1, glog.Warn},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeGormLogLevel(tt.level); got != tt.want {
				t.Errorf("normalizeGormLogLevel(%d) = %d, want %d", tt.level, got, tt.want)
			}
		})
	}
}

func TestGormLevelFromLogConfig(t *testing.T) {
	tests := []struct {
		name  string
		level string
		want  glog.LogLevel
	}{
		{"debug maps to info", "debug", glog.Info},
		{"DEBUG is case insensitive", "DEBUG", glog.Info},
		{"error maps to error", "Error", glog.Error},
		{"warn falls to default", "warn", glog.Warn},
		{"info falls to default", "info", glog.Warn},
		{"empty falls to default", "", glog.Warn},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := gormLevelFromLogConfig(tt.level); got != tt.want {
				t.Errorf("gormLevelFromLogConfig(%q) = %d, want %d", tt.level, got, tt.want)
			}
		})
	}
}

func TestSlowThreshold(t *testing.T) {
	tests := []struct {
		name string
		cfg  DataBase
		want time.Duration
	}{
		{"unset uses default", DataBase{}, 200 * time.Millisecond},
		{"negative uses default", DataBase{SlowThresholdMs: -5}, 200 * time.Millisecond},
		{"configured value kept", DataBase{SlowThresholdMs: 500}, 500 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := slowThreshold(tt.cfg); got != tt.want {
				t.Errorf("slowThreshold(%+v) = %v, want %v", tt.cfg, got, tt.want)
			}
		})
	}
}

func TestLoadFromConfig(t *testing.T) {
	conf.Set("Database.Dialect", "mysql")
	conf.Set("Database.Host", "127.0.0.1")
	conf.Set("Database.Port", "3306")
	conf.Set("Database.User", "root")
	conf.Set("Database.Password", "secret")
	conf.Set("Database.DbName", "demo")
	conf.Set("Database.Pool.MaxIdleConns", 20)
	conf.Set("Database.Pool.MaxOpenConns", 100)
	conf.Set("Database.GormLogLevel", int(glog.Info))
	conf.Set("Database.SlowThresholdMs", 1000)
	defer func() {
		conf.Set("Database.Dialect", nil)
		conf.Set("Database.Host", nil)
		conf.Set("Database.Port", nil)
		conf.Set("Database.User", nil)
		conf.Set("Database.Password", nil)
		conf.Set("Database.DbName", nil)
		conf.Set("Database.Pool.MaxIdleConns", nil)
		conf.Set("Database.Pool.MaxOpenConns", nil)
		conf.Set("Database.GormLogLevel", nil)
		conf.Set("Database.SlowThresholdMs", nil)
	}()

	cfg := LoadFromConfig()

	if cfg.Dialect != "mysql" || cfg.Host != "127.0.0.1" || cfg.Port != "3306" {
		t.Errorf("basic fields not loaded correctly: %+v", cfg)
	}
	if cfg.User != "root" || cfg.Password != "secret" || cfg.DbName != "demo" {
		t.Errorf("credential fields not loaded correctly: %+v", cfg)
	}
	if cfg.Pool.MaxIdleConns != 20 || cfg.Pool.MaxOpenConns != 100 {
		t.Errorf("pool config not loaded correctly: %+v", cfg.Pool)
	}
	if cfg.GormLogLevel != glog.Info {
		t.Errorf("gormLogLevel = %d, want %d", cfg.GormLogLevel, glog.Info)
	}
	if cfg.SlowThresholdMs != 1000 {
		t.Errorf("slowThresholdMs = %d, want 1000", cfg.SlowThresholdMs)
	}
}

func TestLoadFromConfigLogLevelFallback(t *testing.T) {
	conf.Set("Database.GormLogLevel", nil)
	conf.Set("Logging.Level", "debug")
	defer conf.Set("Logging.Level", nil)

	cfg := LoadFromConfig()

	if cfg.GormLogLevel != glog.Info {
		t.Errorf("gormLogLevel = %d, want %d (derived from Logging.Level=debug)", cfg.GormLogLevel, glog.Info)
	}
}

func captureLogOutput(t *testing.T, run func()) string {
	t.Helper()
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.Init() // 恢复默认logger
	run()
	return buf.String()
}

func TestGormLoggerTrace(t *testing.T) {
	fakeFC := func() (string, int64) { return "SELECT * FROM t_user WHERE id = 1", 1 }
	begin := time.Now()

	tests := []struct {
		name      string
		logLevel  glog.LogLevel
		slow      time.Duration
		err       error
		wantParts []string
	}{
		{
			name:      "error is logged at error level",
			logLevel:  glog.Error,
			err:       errors.New("connection refused"),
			wantParts: []string{"connection refused", "SELECT * FROM t_user"},
		},
		{
			name:      "record not found is not logged at error level",
			logLevel:  glog.Error,
			err:       glog.ErrRecordNotFound,
			wantParts: nil,
		},
		{
			name:      "slow sql is logged at warn level",
			logLevel:  glog.Warn,
			slow:      time.Nanosecond,
			wantParts: []string{"SLOW SQL", "SELECT * FROM t_user"},
		},
		{
			name:      "normal sql needs info level",
			logLevel:  glog.Info,
			wantParts: []string{"SELECT * FROM t_user"},
		},
		{
			name:      "normal sql is silent below info level",
			logLevel:  glog.Warn,
			wantParts: nil,
		},
		{
			name:      "silent level logs nothing",
			logLevel:  glog.Silent,
			err:       errors.New("connection refused"),
			wantParts: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureLogOutput(t, func() {
				logger := newGormLogger(tt.logLevel, tt.slow)
				logger.Trace(t.Context(), begin, fakeFC, tt.err)
			})

			for _, part := range tt.wantParts {
				if !strings.Contains(output, part) {
					t.Errorf("trace output should contain %q, got: %s", part, output)
				}
			}
			if len(tt.wantParts) == 0 && output != "" {
				t.Errorf("trace output should be empty, got: %s", output)
			}
		})
	}
}

func TestGormLoggerLogMode(t *testing.T) {
	logger := newGormLogger(glog.Silent, time.Second)

	switched := logger.LogMode(glog.Info)

	if switched.(*gormLogger).logLevel != glog.Info {
		t.Errorf("LogMode should switch log level to %d, got %d", glog.Info, switched.(*gormLogger).logLevel)
	}
	if logger.logLevel != glog.Silent {
		t.Errorf("LogMode should not mutate the original logger, got %d", logger.logLevel)
	}
}

func TestGormLoggerLevels(t *testing.T) {
	cases := []struct {
		name     string
		logLevel glog.LogLevel
		call     func(l *gormLogger)
		want     string
	}{
		{"info", glog.Info, func(l *gormLogger) { l.Info(t.Context(), "query %s", "users") }, "query users"},
		{"warn", glog.Warn, func(l *gormLogger) { l.Warn(t.Context(), "pool %d", 80) }, "pool 80"},
		{"error", glog.Error, func(l *gormLogger) { l.Error(t.Context(), "failed %s", "bad") }, "failed bad"},
		{"info suppressed when silent", glog.Silent, func(l *gormLogger) { l.Info(t.Context(), "query %s", "users") }, ""},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			output := captureLogOutput(t, func() {
				tt.call(newGormLogger(tt.logLevel, time.Second))
			})
			if tt.want == "" {
				if output != "" {
					t.Errorf("output should be empty, got: %s", output)
				}
				return
			}
			if !strings.Contains(output, tt.want) {
				t.Errorf("output should contain %q, got: %s", tt.want, output)
			}
		})
	}
}

// TestPostgresDSNMatchesConfigFormat 断言 postgres DSN 构造与后端现网
// internal/config.PostgresConfig.DSN() 的格式串/实参顺序逐字一致
// （pkg 模块不能 import 后端 internal，期望值用对齐格式串的字面量）。
func TestPostgresDSNMatchesConfigFormat(t *testing.T) {
	tests := []struct {
		name string
		cfg  DataBase
		want string
	}{
		{
			name: "full config matches config.PostgresConfig.DSN() literal",
			cfg: DataBase{
				Dialect: "postgres",
				Host:    "db.prod.internal", Port: "5432",
				User: "learnly", Password: "s3cret",
				DbName: "learnly", SSLMode: "disable",
			},
			want: "host=db.prod.internal port=5432 user=learnly password=s3cret dbname=learnly sslmode=disable TimeZone=Asia/Shanghai",
		},
		{
			name: "require sslmode kept verbatim",
			cfg: DataBase{
				Dialect: "postgres",
				Host:    "pg1", Port: "5433", User: "u", Password: "p",
				DbName: "d", SSLMode: "require",
			},
			want: "host=pg1 port=5433 user=u password=p dbname=d sslmode=require TimeZone=Asia/Shanghai",
		},
		{
			name: "empty sslmode formatted verbatim (disable fallback lives in store layer, not here)",
			cfg: DataBase{
				Dialect: "postgres",
				Host:    "pg1", Port: "5432", User: "u", Password: "p",
				DbName: "d", SSLMode: "",
			},
			want: "host=pg1 port=5432 user=u password=p dbname=d sslmode= TimeZone=Asia/Shanghai",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := postgresDSN(tt.cfg); got != tt.want {
				t.Errorf("postgresDSN() =\n%q\nwant\n%q", got, tt.want)
			}
		})
	}
}

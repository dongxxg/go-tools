package log

import (
	"os"
	"testing"
)

func TestLog(t *testing.T) {
	logger := New(os.Stdout, Config{
		Prefix:       "unit_test_log",
		LogLevel:     "debug",
		AddTimeStamp: true,
		JSONFormat:   true,
		StaticFields: Fields{
			"service": "logtest",
		},
		LogPath: "/tmp/logtest",
	})

	logger.Debug("log something balabala")
}

func TestLogWithFields(t *testing.T) {
	logger := New(os.Stdout, Config{
		Prefix:       "unit_test_log",
		LogLevel:     "debug",
		AddTimeStamp: true,
		JSONFormat:   true,
		StaticFields: Fields{
			"service": "logtest",
		},
		LogPath: "/tmp/logtest",
	})

	logger.Debug("log something balabala again")
	logger.WithFields(map[string]string{
		"a": "b",
	}).Infoln("log with  field")
}

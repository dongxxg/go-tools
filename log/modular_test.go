package log

import (
	"fmt"
	"strings"
	"testing"
)

type LogBuffer struct {
	data string
}

func (l *LogBuffer) Write(p []byte) (n int, err error) {
	l.data = fmt.Sprintf("%v%v", l.data, string(p))
	return len(p), nil
}

func TestModules(t *testing.T) {
	loggerConfig := NewConfig()
	loggerConfig.AddTimeStamp = false
	loggerConfig.JSONFormat = false
	loggerConfig.Prefix = "root"
	loggerConfig.LogLevel = "WARN"

	buf := LogBuffer{data: ""}

	logger := New(&buf, loggerConfig)
	logger.Warnln("Warning message root module")

	logger2 := logger.NewModule(".foo")
	logger2.Warnln("Warning message root.foo module")

	logger3 := logger.NewModule(".foo2")
	logger3.Warnln("Warning message root.foo2 module")

	logger4 := logger2.NewModule(".bar")
	logger4.Warnln("Warning message root.foo.bar module")

	expect1 := "Warning message root module"
	if strings.Index(buf.data, expect1) == -1 {
		t.Errorf("%v not exist in %v", expect1, buf.data)
	}

	expect2 := "Warning message root.foo module"
	if strings.Index(buf.data, expect1) == -1 {
		t.Errorf("%v not exist in %v", expect2, buf.data)
	}

	expect3 := "Warning message root.foo2 module"
	if strings.Index(buf.data, expect1) == -1 {
		t.Errorf("%v not exist in %v", expect3, buf.data)
	}

	expect4 := "Warning message root.foo.bar module"
	if strings.Index(buf.data, expect1) == -1 {
		t.Errorf("%v not exist in %v", expect4, buf.data)
	}
}

func TestStaticFields(t *testing.T) {
	loggerConfig := NewConfig()
	loggerConfig.AddTimeStamp = false
	loggerConfig.JSONFormat = true
	loggerConfig.Prefix = "root"
	loggerConfig.LogLevel = "WARN"
	loggerConfig.StaticFields = Fields{
		"@system": "foo",
	}

	buf := LogBuffer{data: ""}

	logger := New(&buf, loggerConfig)
	logger.Warnln("Warning message root module")

	logger2 := logger.NewModule(".bar")
	logger2.Warnln("Warning message root.bar module")

	expect1 := `"@module":"root.bar"`
	if strings.Index(buf.data, expect1) == -1 {
		t.Errorf("%v not exist in %v", expect1, buf.data)
	}

	expect2 := `"@system":"foo"`
	if strings.Index(buf.data, expect2) == -1 {
		t.Errorf("%v not exist in %v", expect2, buf.data)
	}
}

func TestStaticFieldsOverride(t *testing.T) {
	loggerConfig := NewConfig()
	loggerConfig.AddTimeStamp = false
	loggerConfig.JSONFormat = true
	loggerConfig.Prefix = "root"
	loggerConfig.LogLevel = "WARN"
	loggerConfig.StaticFields = Fields{
		"@service": "uni_base_service",
		"@system":  "foo",
	}

	buf := LogBuffer{data: ""}

	logger := New(&buf, loggerConfig)
	logger.Warnf("Warning message root module")

	logger2 := logger.WithFieldz(Fields{"foo": "bar", "@service": "fooserve"})
	logger2.Warnln("Warning message foo fields")

	expect1 := `"@service":"fooserve"`
	if strings.Index(buf.data, expect1) == -1 {
		t.Errorf("%v not exist in %v", expect1, buf.data)
	}
}

func TestFormattedLogging(t *testing.T) {
	loggerConfig := NewConfig()
	loggerConfig.AddTimeStamp = false
	loggerConfig.JSONFormat = false
	loggerConfig.Prefix = "test"
	loggerConfig.LogLevel = "TRACE"
	loggerConfig.JSONFormat = true

	buf := LogBuffer{data: ""}

	logger := New(&buf, loggerConfig)
	// logger.Fatalf("fatal test %v\n", 1)
	logger.Errorf("error test %v\n", 2)
	logger.Warnf("warn test %v\n", 3)
	logger.Infof("info test %v\n", 4)
	logger.Debugf("info test %v\n", 5)
	logger.Tracef("trace test %v\n", 6)

	expect1 := `{"@level":"error","msg":"error test 2\n"}`
	if strings.Index(buf.data, expect1) == -1 {
		t.Errorf("%v not exist in %v", expect1, buf.data)
	}

	expect2 := `{"@level":"error","msg":"error test 2\n"}`
	if strings.Index(buf.data, expect2) == -1 {
		t.Errorf("%v not exist in %v", expect2, buf.data)
	}
}

func TestLineLogging(t *testing.T) {
	loggerConfig := NewConfig()
	loggerConfig.AddTimeStamp = false
	loggerConfig.JSONFormat = false
	loggerConfig.Prefix = "test"
	loggerConfig.LogLevel = "WARN"
	loggerConfig.JSONFormat = true

	buf := LogBuffer{data: ""}

	logger := New(&buf, loggerConfig)
	// logger.Fatalln("fatal test")
	logger.Errorln("error test")
	logger.Warnln("warn test")
	logger.Infoln("info test")
	logger.Debugln("info test")
	logger.Traceln("trace test")

	expect1 := `{"@level":"error","msg":"error test"}`
	if strings.Index(buf.data, expect1) == -1 {
		t.Errorf("%v not exist in %v", expect1, buf.data)
	}

	expect2 := `{"@level":"error","msg":"error test"}`
	if strings.Index(buf.data, expect2) == -1 {
		t.Errorf("%v not exist in %v", expect2, buf.data)
	}
}

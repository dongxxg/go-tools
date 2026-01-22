package kctx

import (
	"testing"
)

func TestProvider(t *testing.T) {
	c := NewDefaultContext()
	cf, err := c.NewTraceProviderWith("http://192.168.5.136:14268/api/traces",
		"wiken_test")
	if err != nil {
		t.Fatal(err)
	}
	defer cf()
	//
	span := c.TraceAndStart("", "judge")
	defer span.End()

	span2 := c.TraceAndStart("", "input mysql")
	defer span2.End()
	// 传出的spanId
	t.Log(c.SpanId())
}

func TestSpanId(t *testing.T) {
	c := NewDefaultContext()
	cf, err := c.NewTraceProviderWith("http://192.168.5.136:14268/api/traces",
		"unios-alarm-subscribe")
	if err != nil {
		t.Fatal(err)
	}
	defer cf()
	id := "00-b91192f67a89120a5629e433f79b435c-291e3583191a5b14-01"
	span := c.TraceFromIdWith(id, "", "unios-alarm-subscribe")
	defer span.End()
}

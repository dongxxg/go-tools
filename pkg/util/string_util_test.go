package util

import (
	"fmt"
	"reflect"
	"testing"
)

func TestCutString(t *testing.T) {
	str := "{\"version\":1}#[\n[3,{\"addTime\":1"
	a, b, _ := CutString(str, "#")
	if actual, expected := a, "{\"version\":1}"; actual != expected {
		t.Errorf("expeced: %s, got: %s", expected, actual)
	}
	if actual, expected := b, "[\n[3,{\"addTime\":1"; actual != expected {
		t.Errorf("expeced: %s, got: %s", expected, actual)
	}
}

func TestCutBytes(t *testing.T) {
	str := []byte("{\"version\":1}#[\n[3,{\"addTime\":1")
	a, b, _ := CutBytes(str, []byte{'#'})
	if actual, expected := a, []byte("{\"version\":1}"); !reflect.DeepEqual(actual, expected) {
		t.Errorf("expeced: %s, got: %s", expected, actual)
	}
	if actual, expected := b, []byte("[\n[3,{\"addTime\":1"); !reflect.DeepEqual(actual, expected) {
		t.Errorf("expeced: %s, got: %s", expected, actual)
	}
}

func TestString2Cluster(t *testing.T) {
	var str string
	var clusters []string

	str = "Käse"
	clusters = String2Clusters(str)
	if actual, expected := len(clusters), 4; actual != expected {
		t.Errorf("expected: %d, got: %d", expected, actual)
	}

	str = "🏳️‍🌈"
	clusters = String2Clusters(str)
	if actual, expected := len(clusters), 1; actual != expected {
		t.Errorf("expected: %d, got: %d", expected, actual)
	}

	str = "🇩🇪"
	clusters = String2Clusters(str)
	if actual, expected := len(clusters), 1; actual != expected {
		t.Errorf("expected: %d, got: %d", expected, actual)
	}

	str = "🙏"
	clusters = String2Clusters(str)
	if actual, expected := len(clusters), 1; actual != expected {
		t.Errorf("expected: %d, got: %d", expected, actual)
	}

	str = "👋"
	clusters = String2Clusters(str)
	if actual, expected := len(clusters), 1; actual != expected {
		t.Errorf("expected: %d, got: %d", expected, actual)
	}

	str = "👨‍👩‍👧‍👦"
	clusters = String2Clusters(str)
	if actual, expected := len(clusters), 1; actual != expected {
		t.Errorf("expected: %d, got: %d", expected, actual)
	}

	str = "🤦🏻‍♂️"
	clusters = String2Clusters(str)
	if actual, expected := len(clusters), 1; actual != expected {
		t.Errorf("expected: %d, got: %d", expected, actual)
	}

	str = "\r\n"
	clusters = String2Clusters(str)
	if actual, expected := len(clusters), 2; actual != expected {
		t.Errorf("expected: %d, got: %d", expected, actual)
	}

	str = "\n\r"
	clusters = String2Clusters(str)
	if actual, expected := len(clusters), 2; actual != expected {
		t.Errorf("expected: %d, got: %d", expected, actual)
	}
}

func TestRand(t *testing.T) {
	fmt.Println(RandChar(5))
}

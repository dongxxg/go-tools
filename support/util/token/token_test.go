package token

import (
	"testing"
	"time"
)

func TestGenToken(t *testing.T) {
	_, err := GenToken("", "", time.Now(), 600*time.Minute)
	if err != ErrMissKey {
		t.Errorf("expected: %s, got %s\n", ErrMissKey, err)
	}

	token, err := GenToken("abc", "hqmin", time.Now(), 600*time.Minute)
	if err != nil {
		t.Error(err)
	}
	t.Log(token)
}

func TestCheckTokenSuccess(t *testing.T) {
	// check success
	vb3 := time.Now()
	vt3 := 600 * time.Minute
	token3, err := GenToken("abc", "hqmin", vb3, vt3)
	if err != nil {
		t.Error(err)
	}
	err4 := CheckToken("abc", token3)
	if err4 != nil {
		t.Errorf("expect: nil, got: %s\n", err4)
	}
}

func TestCheckTokenErr(t *testing.T) {
	// base64 decode error
	err := CheckToken("abc", "aaa")
	if err != ErrNotBase64 {
		t.Errorf("expect: %s, got: %s\n", ErrNotBase64, err)
	}
	// before ValidBegin error
	vb1 := time.Now().Add(time.Hour)
	vt1 := 600 * time.Minute
	token1, err := GenToken("abc", "hqmin", vb1, vt1)
	if err != nil {
		t.Error(err)
	}
	err1 := CheckToken("abc", token1)
	if err1 != ErrEarly {
		t.Errorf("expect: %s, got: %s\n", ErrEarly, err1)
	}
	// expired error
	vb2 := time.Now().Add(-time.Hour)
	vt2 := time.Minute
	token2, err := GenToken("abc", "hqmin", vb2, vt2)
	if err != nil {
		t.Error(err)
	}
	err2 := CheckToken("abc", token2)
	if err2 != ErrExpired {
		t.Errorf("expect: %s, got: %s\n", ErrExpired, err2)
	}
	// wrong sign key error
	vb3 := time.Now()
	vt3 := 600 * time.Minute
	token3, err := GenToken("abc", "hqmin", vb3, vt3)
	if err != nil {
		t.Error(err)
	}
	err3 := CheckToken("123", token3)
	if err3 != ErrWrongKey {
		t.Errorf("expect: %s, got: %s\n", ErrWrongKey, err3)
	}
}

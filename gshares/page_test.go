package gshares

import (
	"reflect"
	"testing"
	"time"
)

func TestPageToken(t *testing.T) {
	type token struct {
		UpdateTime int64 `json:"update_time"`
		Id         int64 `json:"id"`
	}
	to := &token{
		UpdateTime: time.Now().Unix(),
		Id:         1,
	}
	pageToken, err := EncodePageToken(to)
	if err != nil {
		t.Fatal(err)
	}
	// t.Logf("pageToken:%s", pageToken)
	to2 := &token{}
	err = DecodePageToken(pageToken, to2)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(to, to2) {
		t.Fatalf("not equal")
	}
}

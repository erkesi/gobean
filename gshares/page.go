package gshares

import (
	"encoding/base64"
	"encoding/json"
)

func EncodePageToken(v interface{}) (string, error) {
	bs, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bs), nil
}

func DecodePageToken(pageToken string, v interface{}) error {
	bs, err := base64.StdEncoding.DecodeString(pageToken)
	if err != nil {
		return err
	}
	return json.Unmarshal(bs, v)
}

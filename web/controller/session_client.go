package controller

import (
	b64 "encoding/base64"
	"encoding/json"
)

func EncodeBase64(value any) string {
	to_json, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return b64.StdEncoding.EncodeToString(to_json)
}

func DecodeBase64[T any](value string, to_struct *T) {
	if b, err := b64.StdEncoding.DecodeString(value); err == nil {
		if err := json.Unmarshal(b, to_struct); err != nil {
			panic(err)
		}
	} else {
		panic(err)
	}
}

package reMarshal

import "encoding/json"

func ReMarshal(v interface{}) string {
	// 重新 Marshal
	out, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(out)
}

package src

import (
	"encoding/json"
	"fmt"
)

func ParseData(data string) map[string]interface{} {
	var result map[string]interface{}
	json.Unmarshal([]byte(data), &result)
	fmt.Println(result)
	return result
}
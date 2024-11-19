package src

import (
	"encoding/json"
	"fmt"
)

func ParseData(data string) map[string]interface{} {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(data), &result)
	if err != nil {
			fmt.Println("Error parsing JSON:", err)
			return nil
	}

	return result
}

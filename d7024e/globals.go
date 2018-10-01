package d7024e

import (
    "encoding/json"
    "os"
)

type Configuration struct {
    Alpha    int
    K   int
}

func GetGlobals() Configuration {
	var configuration Configuration
	file, _ := os.Open("d7024e/config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.Decode(&configuration)
	return configuration
}



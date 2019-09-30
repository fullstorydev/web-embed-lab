package commands

import (
	"bufio"
	"log"
	"os"
	"strings"
)

var logger = log.New(os.Stdout, "[auto-formulate] ", 0)

// Set .env values to process env values if process env key doesn't already exist
func EnvOverrideDotEnv(filePath string) error {
	dotEnvVals, err := ParseDotEnv(filePath)
	if err != nil {
		return err
	}
	for key, value := range dotEnvVals {
		_, found := os.LookupEnv(key)
		if found {
			continue
		}
		os.Setenv(key, value)
	}
	return nil
}

func ParseDotEnv(filePath string) (map[string]string, error) {
	results := map[string]string{}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), " 	")
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		separatorIndex := strings.Index(line, "=")
		if separatorIndex == -1 || separatorIndex == len(line)-1 {
			logger.Println("Invalid dotenv line:", line)
			continue
		}
		key := strings.Trim(line[0:separatorIndex], " 	")
		value := strings.Trim(line[separatorIndex+1:], " 	")
		if strings.HasPrefix(value, "\"") {
			value = value[1:]
		}
		if strings.HasSuffix(value, "\"") {
			value = value[0 : len(value)-1]
		}
		results[key] = value
	}

	return results, nil
}

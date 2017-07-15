package cmd

import (
	"os"
	"strconv"
)

func strEnvDef(env string, def string) string {
	val := os.Getenv(env)
	if val == "" {
		return def
	}
	return val
}

func intEnvDef(env string, def int) int {
	val := os.Getenv(env)
	if val == "" {
		return def
	}
	ival, err := strconv.Atoi(val)
	if err != nil {
		return def
	}
	return ival
}

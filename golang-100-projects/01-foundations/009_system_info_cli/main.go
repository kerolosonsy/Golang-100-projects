package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
)

var Allowlist = map[string]bool{
	"HOME": true,
	"USER": true,
	"PATH": true,
}

func getEnvStatus(key string) string {
	if !Allowlist[key] {
		return "(not allowed)"
	}
	value, ok := os.LookupEnv(key)
	if !ok {
		return "(unset)"

	}
	if value == "" {
		return "(empty)"
	}
	return value
}

func main() {
	showOS := flag.Bool("os", false, "Show operating system")
	showArch := flag.Bool("arch", false, "Show architecture")
	showVersion := flag.Bool("version", false, "Show version")
	envName := flag.String("env", "", "Show a specific environment variable")

	flag.Parse()
	if *showVersion {
		fmt.Printf("version: %s\n", runtime.Version())
	}
	if *showOS {
		fmt.Printf("Operating System: %s\n", runtime.GOOS)
	}
	if *showArch {
		fmt.Printf("Architecture: %s\n", runtime.GOARCH)
	}
	if *envName != "" {
		fmt.Printf("Environment Variable: %s\n", getEnvStatus(*envName))
	}
}

package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	ctx := make(map[string]string)
	pluginEnv := os.Environ()
	pluginReg := regexp.MustCompile(`^PLUGIN_(.*)=(.*)`)
	droneReg := regexp.MustCompile(`^DRONE_(.*)=(.*)`)

	for _, value := range pluginEnv {
		if pluginReg.MatchString(value) {
			matches := pluginReg.FindStringSubmatch(value)
			key := strings.ToLower(matches[1])
			ctx[key] = matches[2]
		}

		if droneReg.MatchString(value) {
			matches := droneReg.FindStringSubmatch(value)
			key := strings.ToLower(matches[1])
			ctx[key] = matches[2]
		}
	}
	fmt.Println(ctx)
}

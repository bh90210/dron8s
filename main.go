package main

import (
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	body := strings.NewReader(
		os.Getenv("PLUGIN_BODY"),
	)

	req, err := http.NewRequest(
		os.Getenv("PLUGIN_METHOD"),
		os.Getenv("PLUGIN_URL"),
		body,
	)

	log.Println(req)

	if err != nil {
		os.Exit(1)
	}
}

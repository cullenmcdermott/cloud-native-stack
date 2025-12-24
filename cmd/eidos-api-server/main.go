package main

import (
	"log"

	"github.com/NVIDIA/cloud-native-stack/pkg/api"
)

func main() {
	if err := api.Serve(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

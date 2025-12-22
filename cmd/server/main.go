/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"log"

	"github.com/NVIDIA/cloud-native-stack/pkg/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}

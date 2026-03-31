package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"video-streaming/backend/shared/kafka"
)

type validationInput struct {
	Topics            []kafka.ActualTopicConfig `json:"topics"`
	ExpectedProducers map[string][]string       `json:"expectedProducers"`
	ExpectedConsumers map[string][]string       `json:"expectedConsumers"`
}

func main() {
	inputPath := flag.String("input", "", "path to topic validation json")
	flag.Parse()

	if *inputPath == "" {
		fmt.Fprintln(os.Stderr, "missing -input")
		os.Exit(1)
	}

	raw, err := os.ReadFile(*inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read input: %v\n", err)
		os.Exit(1)
	}

	var input validationInput
	if err := json.Unmarshal(raw, &input); err != nil {
		fmt.Fprintf(os.Stderr, "decode input: %v\n", err)
		os.Exit(1)
	}

	if err := kafka.ValidateTopicBaseline(input.Topics, input.ExpectedProducers, input.ExpectedConsumers); err != nil {
		fmt.Fprintf(os.Stderr, "validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("topic baseline validation passed")
}

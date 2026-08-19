package main

import (
	"context"
	"flag"
	"fmt"
	"os"
)

func main() {
	selfCheck := flag.Bool("self-check", false, "run offline self-check and exit")
	flag.Parse()

	if *selfCheck {
		if err := runSelfCheck(); err != nil {
			fmt.Fprintf(os.Stderr, "self-check failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("self-check: ok")
		return
	}

	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
		os.Exit(1)
	}
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("hwj-macgo-0019 retry budget engine starting (listen=%s, data=%s)\n", cfg.ListenAddr, cfg.DataDir)
}

func runSelfCheck() error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	return offlineSelfCheck(context.Background())
}

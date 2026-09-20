package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/SunilkumarT56/vun-gpu-agent/internal/cli"
	"github.com/SunilkumarT56/vun-gpu-agent/internal/config"
	"github.com/SunilkumarT56/vun-gpu-agent/internal/gpu"
	"github.com/SunilkumarT56/vun-gpu-agent/internal/logger"
	"github.com/SunilkumarT56/vun-gpu-agent/internal/version"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		printVersion()
		return
	}
	if len(os.Args) > 1 && (os.Args[1] == "help" || os.Args[1] == "--help" || os.Args[1] == "-h") {
		printHelp()
		return
	}
	if len(os.Args) == 1 {
		if err := cli.RunREPL(context.Background(), os.Stdin, os.Stdout); err != nil {
			os.Exit(1)
		}
		return
	}

	switch os.Args[1] {
	case "start":
		start(os.Args[2:])
	case "inventory", "discover":
		inventory(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "vun: unknown command %q\n\n", os.Args[1])
		printHelp()
		os.Exit(2)
	}
}

func start(args []string) {
	startFlags := flag.NewFlagSet("start", flag.ContinueOnError)
	startFlags.SetOutput(os.Stderr)
	configPath := startFlags.String("config", "configs/agent.yaml", "path to the agent configuration file")
	if err := startFlags.Parse(args); err != nil {
		os.Exit(2)
	}

	log := logger.New()
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	inventory, err := discoverInventory(context.Background())
	if err != nil {
		log.Error("failed to discover host inventory", "error", err)
		os.Exit(1)
	}
	cli.PrintInventory(os.Stdout, inventory)
	log.Info("gpu agent started", "listen_address", cfg.ListenAddress)
}

func inventory(args []string) {
	inventoryFlags := flag.NewFlagSet("inventory", flag.ContinueOnError)
	inventoryFlags.SetOutput(os.Stderr)
	jsonOutput := inventoryFlags.Bool("json", false, "print JSON output")
	mockOutput := inventoryFlags.Bool("mock", false, "use simulated GPUs")
	if err := inventoryFlags.Parse(args); err != nil {
		os.Exit(2)
	}

	var inventoryData gpu.HostInventory
	var err error
	if *mockOutput {
		inventoryData, err = gpu.DiscoverHostInventoryWith(context.Background(), gpu.NewMockGPUDiscoverer())
	} else {
		inventoryData, err = gpu.DiscoverHostInventory(context.Background())
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "vun: inventory discovery failed: %v\n", err)
		os.Exit(1)
	}
	if *jsonOutput {
		if err := cli.PrintInventoryJSON(os.Stdout, inventoryData); err != nil {
			fmt.Fprintf(os.Stderr, "vun: encode inventory: %v\n", err)
			os.Exit(1)
		}
		return
	}
	cli.PrintInventory(os.Stdout, inventoryData)
}

func printHelp() {
	fmt.Println("vun - GPU agent CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  vun <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  start      Start the agent and print host inventory")
	fmt.Println("  inventory  Discover and print host inventory")
	fmt.Println("             Use --mock for simulated GPUs or --json for JSON output")
	fmt.Println("  version    Show the VUN version")
	fmt.Println("  help       Show this help message")
	fmt.Println()
	fmt.Println("Options for start:")
	fmt.Println("  --config   Path to the agent configuration file")
	fmt.Println()
	fmt.Println("Run vun without arguments to open the interactive shell.")
}

func printVersion() {
	fmt.Printf("vun version %s\n", version.Value)
}

func discoverInventory(ctx context.Context) (gpu.HostInventory, error) {
	return gpu.DiscoverHostInventory(ctx)
}

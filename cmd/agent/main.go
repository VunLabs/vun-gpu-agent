package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/SunilkumarT56/vun-gpu-agent/internal/config"
	"github.com/SunilkumarT56/vun-gpu-agent/internal/gpu"
	"github.com/SunilkumarT56/vun-gpu-agent/internal/logger"
)

func main() {
	configPath := flag.String("config", "configs/agent.yaml", "path to the agent configuration file")
	flag.Parse()

	log := logger.New()
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	inventory, err := gpu.DiscoverHostInventory(context.Background())
	if err != nil {
		log.Error("failed to discover host inventory", "error", err)
		os.Exit(1)
	}
	printInventory(inventory)
	for _, device := range inventory.GPUs {
		printGPU(device)
	}
	fmt.Printf("\nAgent started\n  Listen address: %s\n", cfg.ListenAddress)
}

func printInventory(inventory gpu.HostInventory) {
	fmt.Println("Host Inventory")
	fmt.Println("==============")
	fmt.Printf("  Hostname:     %s\n", inventory.Hostname)
	fmt.Printf("  OS:           %s\n", inventory.OS)
	fmt.Printf("  Architecture: %s\n", inventory.Architecture)

	fmt.Println("\nCPU")
	fmt.Println("---")
	fmt.Printf("  Model:        %s\n", inventory.CPU.Model)
	fmt.Printf("  Cores:        %d\n", inventory.CPU.Cores)
	fmt.Printf("  Threads:      %d\n", inventory.CPU.Threads)
	fmt.Printf("  Architecture: %s\n", inventory.CPU.Architecture)

	fmt.Println("\nMemory")
	fmt.Println("------")
	fmt.Printf("  Total:        %d MB\n", inventory.Memory.TotalMB)
	fmt.Printf("  Used:         %d MB\n", inventory.Memory.UsedMB)
	fmt.Printf("  Free:         %d MB\n", inventory.Memory.FreeMB)

	fmt.Println("\nStorage")
	fmt.Println("-------")
	fmt.Printf("  Total:        %d GB\n", inventory.Storage.TotalGB)
	fmt.Printf("  Used:         %d GB\n", inventory.Storage.UsedGB)
	fmt.Printf("  Free:         %d GB\n", inventory.Storage.FreeGB)
	fmt.Println("\nNVIDIA GPUs")
	fmt.Println("-----------")
	if len(inventory.GPUs) > 0 {
		fmt.Printf("  Found:        %d\n", len(inventory.GPUs))
	} else {
		fmt.Println("  None found")
	}
}

func printGPU(device gpu.Device) {
	fmt.Printf("\nNVIDIA GPU %d\n", device.Index)
	fmt.Println("------------")
	fmt.Printf("  Name:         %s\n", device.Name)
	fmt.Printf("  UUID:         %s\n", device.UUID)
	fmt.Printf("  Memory:       %d MB total, %d MB used, %d MB free\n",
		device.MemoryTotalMB, device.MemoryUsedMB, device.MemoryFreeMB)
	fmt.Printf("  Utilization:  %d%%\n", device.UtilizationPct)
	fmt.Printf("  Temperature:  %d C\n", device.TemperatureC)
	fmt.Printf("  Power:        %.2f W / %.2f W\n", device.PowerDrawW, device.PowerLimitW)
	fmt.Printf("  Driver:       %s\n", device.DriverVersion)
	fmt.Printf("  PCI bus:      %s\n", device.PCIBusID)
}

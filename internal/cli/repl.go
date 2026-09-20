package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/SunilkumarT56/vun-gpu-agent/internal/gpu"
	"github.com/SunilkumarT56/vun-gpu-agent/internal/version"
)

// RunREPL starts the interactive command-line interface.
func RunREPL(ctx context.Context, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	fmt.Fprintln(out, "vun interactive shell")
	fmt.Fprintln(out, `Type "help" for available commands.`)

	for {
		fmt.Fprint(out, "vun> ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("read command: %w", err)
			}
			fmt.Fprintln(out)
			return nil
		}

		command := strings.TrimSpace(scanner.Text())
		if command == "" {
			continue
		}
		if command == "vun" {
			continue
		}
		if strings.HasPrefix(command, "vun ") {
			command = strings.TrimSpace(strings.TrimPrefix(command, "vun "))
		}
		if command == "" {
			continue
		}

		parts := strings.Fields(command)
		name := strings.ToLower(parts[0])
		jsonOutput := false
		mockOutput := false
		for _, part := range parts[1:] {
			switch part {
			case "--json":
				jsonOutput = true
			case "--mock":
				mockOutput = true
			default:
				fmt.Fprintf(out, "usage: %s [--json] [--mock]\n", name)
				jsonOutput = false
				mockOutput = false
				parts = nil
			}
		}

		switch name {
		case "help":
			printHelp(out)
		case "version":
			fmt.Fprintf(out, "vun version %s\n", version.Value)
		case "inventory", "discover":
			if parts == nil {
				continue
			}
			var discoverer gpu.Discoverer
			if mockOutput {
				discoverer = gpu.NewMockGPUDiscoverer()
			}
			if discoverer == nil {
				inventory, err := gpu.DiscoverHostInventory(ctx)
				if err != nil {
					fmt.Fprintf(out, "error: %v\n", err)
					continue
				}
				if jsonOutput {
					if err := PrintInventoryJSON(out, inventory); err != nil {
						fmt.Fprintf(out, "error: %v\n", err)
					}
				} else {
					PrintInventory(out, inventory)
				}
				continue
			}
			if len(parts) > 1 && !jsonOutput && !mockOutput {
				fmt.Fprintf(out, "usage: %s [--json]\n", name)
				continue
			}
			inventory, err := gpu.DiscoverHostInventoryWith(ctx, discoverer)
			if err != nil {
				fmt.Fprintf(out, "error: %v\n", err)
				continue
			}
			if jsonOutput {
				if err := PrintInventoryJSON(out, inventory); err != nil {
					fmt.Fprintf(out, "error: %v\n", err)
				}
			} else {
				PrintInventory(out, inventory)
			}
		case "exit", "quit":
			fmt.Fprintln(out, "Goodbye.")
			return nil
		default:
			fmt.Fprintf(out, "unknown command %q; type \"help\" for available commands\n", command)
		}
	}
}

func printHelp(out io.Writer) {
	fmt.Fprintln(out, "Commands:")
	fmt.Fprintln(out, "  help       Show this help message")
	fmt.Fprintln(out, "  version    Show the agent version")
	fmt.Fprintln(out, "  inventory  Discover and print host inventory")
	fmt.Fprintln(out, "  discover   Alias for inventory")
	fmt.Fprintln(out, "             Add --json for machine-readable output")
	fmt.Fprintln(out, "             Add --mock to use simulated GPUs")
	fmt.Fprintln(out, "  exit       Exit the agent")
}

// PrintInventoryJSON writes an indented JSON host inventory report.
func PrintInventoryJSON(out io.Writer, inventory gpu.HostInventory) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(inventory)
}

// PrintInventory writes a readable host inventory report.
func PrintInventory(out io.Writer, inventory gpu.HostInventory) {
	fmt.Fprintln(out, "Host Inventory")
	fmt.Fprintln(out, "==============")
	fmt.Fprintf(out, "  Hostname:     %s\n", inventory.Hostname)
	fmt.Fprintf(out, "  OS:           %s\n", inventory.OS)
	fmt.Fprintf(out, "  Architecture: %s\n", inventory.Architecture)

	fmt.Fprintln(out, "\nCPU")
	fmt.Fprintln(out, "---")
	fmt.Fprintf(out, "  Model:        %s\n", inventory.CPU.Model)
	fmt.Fprintf(out, "  Cores:        %d\n", inventory.CPU.Cores)
	fmt.Fprintf(out, "  Threads:      %d\n", inventory.CPU.Threads)
	fmt.Fprintf(out, "  Architecture: %s\n", inventory.CPU.Architecture)

	fmt.Fprintln(out, "\nMemory")
	fmt.Fprintln(out, "------")
	fmt.Fprintf(out, "  Total:        %d MB\n", inventory.Memory.TotalMB)
	fmt.Fprintf(out, "  Used:         %d MB\n", inventory.Memory.UsedMB)
	fmt.Fprintf(out, "  Free:         %d MB\n", inventory.Memory.FreeMB)

	fmt.Fprintln(out, "\nStorage")
	fmt.Fprintln(out, "-------")
	fmt.Fprintf(out, "  Total:        %d GB\n", inventory.Storage.TotalGB)
	fmt.Fprintf(out, "  Used:         %d GB\n", inventory.Storage.UsedGB)
	fmt.Fprintf(out, "  Free:         %d GB\n", inventory.Storage.FreeGB)

	fmt.Fprintln(out, "\nNVIDIA GPUs")
	fmt.Fprintln(out, "-----------")
	if len(inventory.GPUs) == 0 {
		fmt.Fprintln(out, "  None found")
		return
	}
	fmt.Fprintf(out, "  Found:        %d\n", len(inventory.GPUs))
	for _, device := range inventory.GPUs {
		printGPU(out, device)
	}
}

func printGPU(out io.Writer, device gpu.Device) {
	fmt.Fprintf(out, "\n  NVIDIA GPU %d\n", device.Index)
	fmt.Fprintf(out, "    Name:         %s\n", device.Name)
	fmt.Fprintf(out, "    UUID:         %s\n", device.UUID)
	fmt.Fprintf(out, "    Memory:       %d MB total, %d MB used, %d MB free\n",
		device.MemoryTotalMB, device.MemoryUsedMB, device.MemoryFreeMB)
	fmt.Fprintf(out, "    Utilization:  %d%%\n", device.UtilizationPct)
	fmt.Fprintf(out, "    Temperature:  %d C\n", device.TemperatureC)
	fmt.Fprintf(out, "    Power:        %.2f W / %.2f W\n", device.PowerDrawW, device.PowerLimitW)
	fmt.Fprintf(out, "    Driver:       %s\n", device.DriverVersion)
	fmt.Fprintf(out, "    PCI bus:      %s\n", device.PCIBusID)
}

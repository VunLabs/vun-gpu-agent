package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/SunilkumarT56/vun-gpu-agent/internal/config"
	"github.com/SunilkumarT56/vun-gpu-agent/internal/enrollment"
	"github.com/SunilkumarT56/vun-gpu-agent/internal/gpu"
	"github.com/SunilkumarT56/vun-gpu-agent/internal/state"
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
		if name == "inventory" || name == "discover" {
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
		}

		switch name {
		case "help":
			printHelp(out)
		case "version":
			fmt.Fprintf(out, "vun version %s\n", version.Value)
		case "enroll":
			if err := Enroll(ctx, out, parts[1:]); err != nil {
				fmt.Fprintf(out, "error: %v\n", err)
			}
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
	fmt.Fprintln(out, "  enroll     Enroll this host with the VUN API")
	fmt.Fprintln(out, "  inventory  Discover and print host inventory")
	fmt.Fprintln(out, "  discover   Alias for inventory")
	fmt.Fprintln(out, "             Add --json for machine-readable output")
	fmt.Fprintln(out, "             Add --mock to use simulated GPUs")
	fmt.Fprintln(out, "  exit       Exit the agent")
}

// Enroll discovers the host and registers it with the configured VUN API.
func Enroll(ctx context.Context, out io.Writer, args []string) error {
	token := os.Getenv("VUN_ENROLLMENT_TOKEN")
	configPath := "configs/agent.yaml"
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--token":
			if i+1 >= len(args) {
				return fmt.Errorf("usage: enroll [--token TOKEN] [--config PATH]")
			}
			token = args[i+1]
			i++
		case "--config":
			if i+1 >= len(args) {
				return fmt.Errorf("usage: enroll [--token TOKEN] [--config PATH]")
			}
			configPath = args[i+1]
			i++
		default:
			return fmt.Errorf("unknown enroll option %q", args[i])
		}
	}
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("enrollment token is required; set VUN_ENROLLMENT_TOKEN or use --token")
	}
	var cfg config.Config
	var err error
	if configPath == "configs/agent.yaml" {
		cfg, err = config.Load(configPath)
		if errors.Is(err, os.ErrNotExist) {
			cfg = config.Defaults()
		} else if err != nil {
			return err
		}
	} else {
		cfg, err = config.Load(configPath)
		if err != nil {
			return err
		}
	}
	if apiURL := strings.TrimSpace(os.Getenv("VUN_API_URL")); apiURL != "" {
		cfg.APIURL = strings.TrimRight(apiURL, "/")
	}
	inventory, err := gpu.DiscoverHostInventory(ctx)
	if err != nil {
		return fmt.Errorf("discover host inventory: %w", err)
	}
	result, err := (enrollment.Client{BaseURL: cfg.APIURL}).Enroll(ctx, token, version.Value, inventory)
	if err != nil {
		return err
	}
	if err := state.SaveEnrollment(state.Enrollment{HostID: result.HostID, AgentCredential: result.AgentCredential}); err != nil {
		return err
	}
	fmt.Fprintf(out, "Enrollment successful\n  Host ID: %s\n", result.HostID)
	fmt.Fprintln(out, "  Agent credential saved locally.")
	return nil
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

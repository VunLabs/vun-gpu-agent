package gpu

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
)

var defaultNvidiaSMI = "nvidia-smi"

const query = "index,uuid,name,memory.total,memory.used,memory.free,utilization.gpu,temperature.gpu,power.draw,power.limit,driver_version,pci.bus_id"

// Device describes a GPU reported by nvidia-smi.
type Device struct {
	Index          int     `json:"index"`
	UUID           string  `json:"uuid"`
	Name           string  `json:"name"`
	MemoryTotalMB  uint64  `json:"memoryTotalMB"`
	MemoryUsedMB   uint64  `json:"memoryUsedMB"`
	MemoryFreeMB   uint64  `json:"memoryFreeMB"`
	UtilizationPct uint64  `json:"utilizationPct"`
	TemperatureC   uint64  `json:"temperatureC"`
	PowerDrawW     float64 `json:"powerDrawW"`
	PowerLimitW    float64 `json:"powerLimitW"`
	DriverVersion  string  `json:"driverVersion"`
	PCIBusID       string  `json:"pciBusID"`
}

// Discover executes nvidia-smi and returns the installed NVIDIA GPUs.
func Discover(ctx context.Context) ([]Device, error) {
	output, err := exec.CommandContext(
		ctx,
		defaultNvidiaSMI,
		"--query-gpu="+query,
		"--format=csv,noheader,nounits",
	).Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return []Device{}, nil
		}

		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			stderr := strings.TrimSpace(string(exitError.Stderr))
			if strings.Contains(strings.ToLower(stderr), "no devices were found") {
				return []Device{}, nil
			}
			if stderr != "" {
				return nil, fmt.Errorf("nvidia-smi failed: %s: %w", stderr, err)
			}
		}
		return nil, fmt.Errorf("run nvidia-smi: %w", err)
	}

	devices, err := parse(string(output))
	if err != nil {
		return nil, fmt.Errorf("parse nvidia-smi output: %w", err)
	}
	return devices, nil
}

func parse(output string) ([]Device, error) {
	reader := csv.NewReader(strings.NewReader(output))
	reader.FieldsPerRecord = 12
	reader.TrimLeadingSpace = true

	var devices []Device
	for line := 1; ; line++ {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		if len(record) == 0 || (len(record) == 1 && strings.TrimSpace(record[0]) == "") {
			continue
		}

		device, err := parseRecord(record)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		devices = append(devices, device)
	}
	return devices, nil
}

func parseRecord(record []string) (Device, error) {
	for i := range record {
		record[i] = strings.TrimSpace(record[i])
	}

	index, err := strconv.Atoi(record[0])
	if err != nil || index < 0 {
		return Device{}, fmt.Errorf("invalid GPU index %q", record[0])
	}
	if record[1] == "" || record[2] == "" {
		return Device{}, errors.New("GPU UUID and name are required")
	}

	memoryTotal, err := parseUint(record[3], "memory total")
	if err != nil {
		return Device{}, err
	}
	memoryUsed, err := parseUint(record[4], "memory used")
	if err != nil {
		return Device{}, err
	}
	memoryFree, err := parseUint(record[5], "memory free")
	if err != nil {
		return Device{}, err
	}
	utilization, err := parseUint(record[6], "utilization")
	if err != nil || utilization > 100 {
		if err == nil {
			err = errors.New("must be between 0 and 100")
		}
		return Device{}, fmt.Errorf("invalid utilization %q: %w", record[6], err)
	}
	temperature, err := parseUint(record[7], "temperature")
	if err != nil {
		return Device{}, err
	}
	powerDraw, err := parseFloat(record[8], "power draw")
	if err != nil {
		return Device{}, err
	}
	powerLimit, err := parseFloat(record[9], "power limit")
	if err != nil {
		return Device{}, err
	}

	return Device{
		Index:          index,
		UUID:           record[1],
		Name:           record[2],
		MemoryTotalMB:  memoryTotal,
		MemoryUsedMB:   memoryUsed,
		MemoryFreeMB:   memoryFree,
		UtilizationPct: utilization,
		TemperatureC:   temperature,
		PowerDrawW:     powerDraw,
		PowerLimitW:    powerLimit,
		DriverVersion:  record[10],
		PCIBusID:       record[11],
	}, nil
}

func parseUint(value, field string) (uint64, error) {
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", field, value, err)
	}
	return parsed, nil
}

func parseFloat(value, field string) (float64, error) {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", field, value, err)
	}
	return parsed, nil
}

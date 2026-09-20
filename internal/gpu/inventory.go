package gpu

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

// HostInventory describes the compute resources available on the host.
type HostInventory struct {
	Hostname     string `json:"hostname"`
	OS           string `json:"os"`
	Architecture string `json:"architecture"`

	CPU     CPUInfo     `json:"cpu"`
	Memory  MemoryInfo  `json:"memory"`
	Storage StorageInfo `json:"storage"`

	GPUs []Device `json:"gpus"`
}

// CPUInfo describes the host CPU.
type CPUInfo struct {
	Model        string `json:"model"`
	Cores        int    `json:"cores"`
	Threads      int    `json:"threads"`
	Architecture string `json:"architecture"`
}

// MemoryInfo describes host memory in megabytes.
type MemoryInfo struct {
	TotalMB uint64 `json:"totalMB"`
	UsedMB  uint64 `json:"usedMB"`
	FreeMB  uint64 `json:"freeMB"`
}

// StorageInfo describes the storage available on the host root volume.
type StorageInfo struct {
	TotalGB uint64 `json:"totalGB"`
	UsedGB  uint64 `json:"usedGB"`
	FreeGB  uint64 `json:"freeGB"`
}

// Discoverer provides GPU discovery for host inventory collection.
type Discoverer interface {
	Discover(context.Context) ([]Device, error)
}

type nvidiaSMIDiscoverer struct{}

func (nvidiaSMIDiscoverer) Discover(ctx context.Context) ([]Device, error) {
	return Discover(ctx)
}

// DiscoverHostInventory collects host metadata and NVIDIA GPU information.
func DiscoverHostInventory(ctx context.Context) (HostInventory, error) {
	return DiscoverHostInventoryWith(ctx, nvidiaSMIDiscoverer{})
}

// DiscoverHostInventoryWith collects host metadata using the supplied GPU discoverer.
func DiscoverHostInventoryWith(ctx context.Context, discoverer Discoverer) (HostInventory, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return HostInventory{}, fmt.Errorf("get hostname: %w", err)
	}

	cpu, err := discoverCPU()
	if err != nil {
		return HostInventory{}, fmt.Errorf("discover CPU: %w", err)
	}

	memory, err := discoverMemory()
	if err != nil {
		return HostInventory{}, fmt.Errorf("discover memory: %w", err)
	}

	storage, err := discoverStorage()
	if err != nil {
		return HostInventory{}, fmt.Errorf("discover storage: %w", err)
	}

	gpus, err := discoverer.Discover(ctx)
	if err != nil {
		return HostInventory{}, fmt.Errorf("discover GPUs: %w", err)
	}

	return HostInventory{
		Hostname:     hostname,
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		CPU:          cpu,
		Memory:       memory,
		Storage:      storage,
		GPUs:         gpus,
	}, nil
}

func discoverStorage() (StorageInfo, error) {
	var stats syscall.Statfs_t
	if err := syscall.Statfs("/", &stats); err != nil {
		return StorageInfo{}, fmt.Errorf("stat root filesystem: %w", err)
	}

	blockSize := uint64(stats.Bsize)
	totalBytes := uint64(stats.Blocks) * blockSize
	freeBytes := uint64(stats.Bavail) * blockSize
	if freeBytes > totalBytes {
		freeBytes = totalBytes
	}

	const bytesPerGB = 1024 * 1024 * 1024
	return StorageInfo{
		TotalGB: totalBytes / bytesPerGB,
		UsedGB:  (totalBytes - freeBytes) / bytesPerGB,
		FreeGB:  freeBytes / bytesPerGB,
	}, nil
}

func discoverCPU() (CPUInfo, error) {
	threads := runtime.NumCPU()
	if threads < 1 {
		return CPUInfo{}, errors.New("runtime reported no CPU threads")
	}

	cpu := CPUInfo{
		Architecture: runtime.GOARCH,
		Threads:      threads,
		Cores:        threads,
		Model:        "unknown",
	}

	switch runtime.GOOS {
	case "linux":
		file, err := os.Open("/proc/cpuinfo")
		if err != nil {
			return CPUInfo{}, fmt.Errorf("open /proc/cpuinfo: %w", err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			key, value, found := strings.Cut(scanner.Text(), ":")
			if found && strings.TrimSpace(key) == "model name" {
				cpu.Model = strings.TrimSpace(value)
				break
			}
		}
		if err := scanner.Err(); err != nil {
			return CPUInfo{}, fmt.Errorf("read /proc/cpuinfo: %w", err)
		}
	case "darwin":
		output, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output()
		if err != nil {
			return CPUInfo{}, fmt.Errorf("read CPU model with sysctl: %w", err)
		}
		cpu.Model = strings.TrimSpace(string(output))
	default:
		return CPUInfo{}, fmt.Errorf("CPU discovery is unsupported on %s", runtime.GOOS)
	}

	if cpu.Model == "" {
		return CPUInfo{}, errors.New("CPU model is empty")
	}
	return cpu, nil
}

func discoverMemory() (MemoryInfo, error) {
	switch runtime.GOOS {
	case "linux":
		return discoverLinuxMemory()
	case "darwin":
		return discoverDarwinMemory()
	default:
		return MemoryInfo{}, fmt.Errorf("memory discovery is unsupported on %s", runtime.GOOS)
	}
}

func discoverLinuxMemory() (MemoryInfo, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return MemoryInfo{}, fmt.Errorf("open /proc/meminfo: %w", err)
	}
	defer file.Close()

	values := make(map[string]uint64)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key, value, found := strings.Cut(scanner.Text(), ":")
		if !found {
			continue
		}
		fields := strings.Fields(value)
		if len(fields) == 0 {
			continue
		}
		kb, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			return MemoryInfo{}, fmt.Errorf("parse %s: %w", strings.TrimSpace(key), err)
		}
		values[strings.TrimSpace(key)] = kb
	}
	if err := scanner.Err(); err != nil {
		return MemoryInfo{}, fmt.Errorf("read /proc/meminfo: %w", err)
	}

	total, ok := values["MemTotal"]
	if !ok {
		return MemoryInfo{}, errors.New("MemTotal is missing from /proc/meminfo")
	}
	free := values["MemAvailable"]
	if free == 0 {
		free = values["MemFree"]
	}
	return MemoryInfo{
		TotalMB: total / 1024,
		FreeMB:  free / 1024,
		UsedMB:  (total - minUint64(total, free)) / 1024,
	}, nil
}

func discoverDarwinMemory() (MemoryInfo, error) {
	totalOutput, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err != nil {
		return MemoryInfo{}, fmt.Errorf("read total memory with sysctl: %w", err)
	}
	total, err := strconv.ParseUint(strings.TrimSpace(string(totalOutput)), 10, 64)
	if err != nil {
		return MemoryInfo{}, fmt.Errorf("parse total memory: %w", err)
	}

	vmOutput, err := exec.Command("vm_stat").Output()
	if err != nil {
		return MemoryInfo{}, fmt.Errorf("read memory statistics with vm_stat: %w", err)
	}
	pageSize, err := darwinPageSize()
	if err != nil {
		return MemoryInfo{}, err
	}

	pages := parseVMStatPages(string(vmOutput))
	freePages := pages["Pages free"] +
		pages["Pages inactive"] +
		pages["Pages speculative"]
	freeBytes := freePages * pageSize
	if freeBytes > total {
		freeBytes = total
	}

	return MemoryInfo{
		TotalMB: total / (1024 * 1024),
		UsedMB:  (total - freeBytes) / (1024 * 1024),
		FreeMB:  freeBytes / (1024 * 1024),
	}, nil
}

func darwinPageSize() (uint64, error) {
	output, err := exec.Command("sysctl", "-n", "vm.pagesize").Output()
	if err != nil {
		return 0, fmt.Errorf("read VM page size with sysctl: %w", err)
	}
	pageSize, err := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse VM page size %q: %w", strings.TrimSpace(string(output)), err)
	}
	if pageSize == 0 {
		return 0, errors.New("VM page size is zero")
	}
	return pageSize, nil
}

func parseVMStatPages(output string) map[string]uint64 {
	pages := make(map[string]uint64)
	for _, line := range strings.Split(output, "\n") {
		key, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		value = strings.TrimSpace(strings.TrimSuffix(value, "."))
		parsed, err := strconv.ParseUint(value, 10, 64)
		if err == nil {
			pages[strings.TrimSpace(key)] = parsed
		}
	}
	return pages
}

func minUint64(left, right uint64) uint64 {
	if left < right {
		return left
	}
	return right
}

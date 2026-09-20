package gpu

import "context"

// MockGPUDiscoverer returns deterministic GPU data for demos and testing.
type MockGPUDiscoverer struct{}

// NewMockGPUDiscoverer creates a mock GPU discoverer.
func NewMockGPUDiscoverer() *MockGPUDiscoverer {
	return &MockGPUDiscoverer{}
}

// Discover returns two simulated NVIDIA RTX 4090 devices.
func (m *MockGPUDiscoverer) Discover(ctx context.Context) ([]Device, error) {
	return []Device{
		{
			Index:          0,
			UUID:           "GPU-MOCK-001",
			Name:           "NVIDIA RTX 4090",
			MemoryTotalMB:  24576,
			MemoryUsedMB:   4096,
			MemoryFreeMB:   20480,
			UtilizationPct: 15,
			TemperatureC:   55,
			PowerDrawW:     120,
			PowerLimitW:    450,
			DriverVersion:  "550.54.14",
			PCIBusID:       "0000:01:00.0",
		},
		{
			Index:          1,
			UUID:           "GPU-MOCK-002",
			Name:           "NVIDIA RTX 4090",
			MemoryTotalMB:  24576,
			MemoryUsedMB:   8192,
			MemoryFreeMB:   16384,
			UtilizationPct: 32,
			TemperatureC:   61,
			PowerDrawW:     180,
			PowerLimitW:    450,
			DriverVersion:  "550.54.14",
			PCIBusID:       "0000:02:00.0",
		},
	}, nil
}

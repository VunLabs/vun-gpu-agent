package gpu

import "testing"

func TestParse(t *testing.T) {
	devices, err := parse(`0, GPU-123, NVIDIA A100-SXM4-40GB, 40960, 1024, 39936, 25, 48, 125.50, 250.00, 550.54.15, 00000000:01:00.0
1, GPU-456, "NVIDIA RTX, Example", 8192, 0, 8192, 0, 35, 0.00, 150.00, 555.42.02, 00000000:02:00.0`)
	if err != nil {
		t.Fatalf("parse() error = %v", err)
	}
	if len(devices) != 2 {
		t.Fatalf("parse() returned %d devices, want 2", len(devices))
	}
	if devices[0].Name != "NVIDIA A100-SXM4-40GB" ||
		devices[0].MemoryTotalMB != 40960 ||
		devices[0].MemoryFreeMB != 39936 ||
		devices[0].UtilizationPct != 25 ||
		devices[0].PowerDrawW != 125.5 ||
		devices[0].DriverVersion != "550.54.15" {
		t.Fatalf("unexpected first device: %+v", devices[0])
	}
	if devices[1].Name != "NVIDIA RTX, Example" {
		t.Fatalf("quoted device name was not parsed: %q", devices[1].Name)
	}
}

func TestParseRejectsInvalidUtilization(t *testing.T) {
	_, err := parse("0,GPU-123,NVIDIA GPU,4096,0,4096,101,40,10.0,100.0,550.54.15,00000000:01:00.0")
	if err == nil {
		t.Fatal("parse() succeeded for utilization above 100")
	}
}

func TestMissingNvidiaSMIIsNonFatal(t *testing.T) {
	original := defaultNvidiaSMI
	defer func() {
		defaultNvidiaSMI = original
	}()
	defaultNvidiaSMI = "nvidia-smi-command-that-does-not-exist"

	devices, err := Discover(t.Context())
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	if len(devices) != 0 {
		t.Fatalf("Discover() returned %d devices, want 0", len(devices))
	}
}

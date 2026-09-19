package gpu

import "testing"

func TestStorageInfoShape(t *testing.T) {
	storage, err := discoverStorage()
	if err != nil {
		t.Fatalf("discoverStorage() error = %v", err)
	}
	if storage.TotalGB == 0 {
		t.Fatal("discoverStorage() returned zero total storage")
	}
	if storage.UsedGB+storage.FreeGB > storage.TotalGB {
		t.Fatalf("storage values exceed total: %+v", storage)
	}
}

func TestParseVMStatPages(t *testing.T) {
	pages := parseVMStatPages(`Mach Virtual Memory Statistics: (page size of 16384 bytes)
Pages free:                             100.
Pages active:                           200.
Pages inactive:                         300.
Pages speculative:                       50.
Pages wired down:                       75.
`)

	if pages["Pages free"] != 100 {
		t.Fatalf("free pages = %d, want 100", pages["Pages free"])
	}
	if pages["Pages inactive"] != 300 {
		t.Fatalf("inactive pages = %d, want 300", pages["Pages inactive"])
	}
	if pages["Pages speculative"] != 50 {
		t.Fatalf("speculative pages = %d, want 50", pages["Pages speculative"])
	}
}

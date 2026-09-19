# VUN GPU Agent

VUN GPU Agent is a Go service scaffold for monitoring and managing GPU
workloads.

## Project layout

```text
cmd/agent/       Application entry point
internal/config/ Configuration loading
internal/logger/ Structured logging
configs/         Example configuration
```

## Run

Create a local configuration from the example, then start the agent:

```sh
cp configs/agent.example.yaml configs/agent.yaml
go run ./cmd/agent -config configs/agent.yaml
```

The default configuration listens on `:8080`. The initial scaffold starts the
process, discovers host metadata (hostname, OS, architecture, CPU, memory, and storage),
discovers NVIDIA GPUs with `nvidia-smi`, and logs the inventory as structured
fields. GPU discovery uses:

```text
nvidia-smi --query-gpu=index,uuid,name,memory.total,memory.used,memory.free,utilization.gpu,temperature.gpu,power.draw,power.limit,driver_version,pci.bus_id --format=csv,noheader,nounits
```

The NVIDIA driver and `nvidia-smi` must be installed and available on `PATH`.
The inventory model is composed of `HostInventory`, `CPUInfo`, `MemoryInfo`,
and `Device` values in `internal/gpu`.
# vun-gpu-agent

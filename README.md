# VUN GPU Agent

VUN GPU Agent is a Go service for monitoring and managing GPU workloads.

## Run

Create a local configuration from the example:

```sh
cp configs/agent.example.yaml configs/agent.yaml
```

Build the CLI:

```sh
make build
```

The binary is `bin/vun`, so interact with the agent using `vun` commands:

```sh
./bin/vun version
./bin/vun start --config configs/agent.yaml
./bin/vun inventory --mock
./bin/vun inventory --mock --json
./bin/vun help
```

For local development, use `go run`:

```sh
go run ./cmd/agent start --config configs/agent.yaml
```

## Version

Show the VUN version:

```sh
go run ./cmd/agent version
```

Output:

```text
vun version dev
```

Release builds can set a version through linker flags:

```sh
go build \
  -ldflags "-X github.com/SunilkumarT56/vun-gpu-agent/internal/version.Value=v1.0.0" \
  -o bin/vun ./cmd/agent
./bin/vun version
```

Output:

```text
vun version v1.0.0
```

## Interactive shell

Run without arguments to open the VUN REPL:

```sh
go run ./cmd/agent
```

Available commands:

```text
help       Show available commands
version    Show the VUN version
inventory  Discover and print host inventory
discover   Alias for inventory
exit       Exit the shell
```

Use mock GPUs without NVIDIA hardware:

```text
vun> inventory --mock
vun> inventory --mock --json
```

The mock inventory returns two simulated NVIDIA RTX 4090 devices. Normal
`inventory` discovery continues to use the existing `nvidia-smi` integration.

The agent discovers host metadata (hostname, OS, architecture, CPU, memory,
and storage) and NVIDIA GPUs with `nvidia-smi`. The NVIDIA driver and
`nvidia-smi` must be installed and available on `PATH`.

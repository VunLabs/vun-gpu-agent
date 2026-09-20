package version

// Value is the agent version. Release builds can override it with:
//
//	go build -ldflags "-X github.com/SunilkumarT56/vun-gpu-agent/internal/version.Value=v1.0.0"
var Value = "dev"

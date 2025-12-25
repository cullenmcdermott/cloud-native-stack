# CLI Architecture

The `eidos` CLI provides command-line access to Cloud Native Stack configuration management capabilities.

## Overview

The CLI is built on the [urfave/cli/v3](https://github.com/urfave/cli) framework and provides two main commands:
- `snapshot` - Capture system configuration
- `recipe` - Generate configuration recommendations

## Architecture Diagram

```mermaid
flowchart TD
    A["eidos CLI<br/>cmd/eidos/main.go"] --> B["Root Command<br/>pkg/cli/root.go"]
    
    B --> B1["Version info (ldflags)<br/>Debug flag → Logging<br/>Shell completion"]
    
    B --> C["snapshot CMD<br/>pkg/cli/snapshot.go"]
    B --> D["recipe CMD<br/>pkg/cli/recipe.go"]
    
    C --> E[Shared Packages]
    D --> E
    
    E --> E1["Collector Factory"]
    E --> E2["Recipe Builder"]
    E --> E3["Serializer<br/>(JSON/YAML/Table)"]
    
    style A fill:#e1f5ff
    style B fill:#fff4e1
    style C fill:#e8f5e9
    style D fill:#e8f5e9
    style E fill:#f3e5f5
```

## Component Details

### Entry Point: `cmd/eidos/main.go`

Minimal entry point that delegates to the CLI package:

```go
package main

import "github.com/NVIDIA/cloud-native-stack/pkg/cli"

func main() {
    cli.Execute()
}
```

### Root Command: `pkg/cli/root.go`

**Responsibilities:**
- Command registration and routing
- Version information injection (via ldflags)
- Global flag handling (debug mode)
- Structured logging initialization

**Key Features:**
- Version info: `version`, `commit`, `date` (overridden at build time)
- Debug flag: `--debug` → Sets log level to debug
- Shell completion support
- Command listing for auto-completion

### Snapshot Command: `pkg/cli/snapshot.go`

Captures comprehensive system configuration snapshots.

#### Command Flow

```mermaid
flowchart TD
    A[User Invocation] --> B[Parse Flags<br/>format, output]
    B --> C[Create Collector Factory]
    C --> D[Initialize NodeSnapshotter]
    D --> E[Parallel Collection<br/>errgroup]
    E --> F[Aggregate Measurements]
    F --> G[Serialize Output]
    G --> H[Write to stdout/file]
    
    style E fill:#ffeb3b
```

#### Detailed Data Flow

```mermaid
flowchart TD
    A[Snapshot Command] --> B[collector.NewDefaultFactory]
    
    B --> B1["OSCollector<br/>(grub, kmod, sysctl)"]
    B --> B2["SystemDCollector<br/>(containerd, docker, kubelet)"]
    B --> B3["KubernetesCollector<br/>(server, images, policies)"]
    B --> B4["GPUCollector<br/>(nvidia-smi data)"]
    
    B1 & B2 & B3 & B4 --> C[NodeSnapshotter.Measure]
    
    C --> D["Parallel Collection<br/>(errgroup)"]
    
    D --> D1["Go Routine 1: Metadata<br/>• snapshot-version<br/>• source-node<br/>• timestamp"]
    D --> D2["Go Routine 2: Kubernetes<br/>• Server Version<br/>• Container Images<br/>• ClusterPolicies"]
    D --> D3["Go Routine 3: SystemD<br/>• containerd.service<br/>• docker.service<br/>• kubelet.service"]
    D --> D4["Go Routine 4: OS Config<br/>• GRUB parameters<br/>• Kernel modules<br/>• Sysctl parameters"]
    D --> D5["Go Routine 5: GPU<br/>• nvidia-smi properties<br/>• driver, CUDA, etc."]
    
    D1 & D2 & D3 & D4 & D5 --> E["All goroutines complete<br/>or first error returns"]
    
    E --> F["Snapshot Structure<br/>kind: Snapshot<br/>apiVersion: snapshot.dgxc.io/v1<br/>measurements: [k8s, systemd, os, gpu]"]
    
    F --> G[serializer.NewFileWriterOrStdout]
    
    G --> G1["Format: JSON/YAML/Table"]
    G --> G2["Output: stdout or file"]
    
    style D fill:#ffeb3b
    style D1 fill:#c8e6c9
    style D2 fill:#c8e6c9
    style D3 fill:#c8e6c9
    style D4 fill:#c8e6c9
    style D5 fill:#c8e6c9
```

#### Usage Examples

```bash
# Output to stdout in JSON format
eidos snapshot

# Save to file in YAML format
eidos snapshot --output system.yaml --format yaml

# Human-readable table format
eidos snapshot --format table
```

### Recipe Command: `pkg/cli/recipe.go`

Generates optimized configuration recipes based on environment parameters.

#### Command Flow

```mermaid
flowchart TD
    A[User Flags] --> B[Build Query from Flags]
    B --> C[Parse & Validate Versions]
    C --> D[recipe.BuildRecipe]
    D --> E["Load Recipe Store<br/>(embedded YAML)"]
    E --> F[Match Overlays]
    F --> G[Merge Measurements]
    G --> H[Serialize Output]
    H --> I[Write to stdout/file]
    
    style E fill:#fff9c4
    style F fill:#ffccbc
```

#### Detailed Data Flow

```mermaid
flowchart TD
    A[Recipe Command] --> B[buildQueryFromCmd]
    
    B --> B1["Parse CLI Flags:<br/>--os, --osv, --kernel<br/>--service, --k8s<br/>--gpu, --intent, --context"]
    B1 --> B2["Version Parsing:<br/>• ParseVersion for osv, kernel, k8s<br/>• Reject negative components<br/>• Support precision (1.2.3, 1.2, 1)"]
    
    B2 --> C[recipe.BuildRecipe]
    
    C --> C1["Step 1: Load Recipe Store<br/>(embedded YAML, cached)"]
    C1 --> C2["Step 2: Clone Base Measurements<br/>(deep copy: os, systemd, k8s, gpu)"]
    C2 --> C3["Step 3: Match Overlays<br/>• For each overlay: IsMatch(query)<br/>• Matching: empty=any, else equal<br/>• Version matching with precision"]
    C3 --> C4["Step 4: Merge Overlay Measurements<br/>• Index by measurement.Type<br/>• Merge subtypes by name<br/>• Overlay data takes precedence"]
    C4 --> C5["Step 5: Strip Context<br/>(if not requested)"]
    C5 --> C6["Recipe Structure:<br/>request, matchedRuleId<br/>payloadVersion, generatedAt<br/>measurements"]
    
    C6 --> D["serializer.NewFileWriterOrStdout<br/>(JSON/YAML/Table)"]
    
    style C1 fill:#fff9c4
    style C3 fill:#ffccbc
    style C4 fill:#c5cae9
```

#### Recipe Matching Algorithm

The recipe matching uses a **rule-based query system** where overlays specify keys that must match the user's query:

```yaml
overlays:
  - key:
      service: eks          # Rule: must have service=eks
      os: ubuntu           # Rule: must have os=ubuntu
    types:
      - type: os
        subtypes:
          - subtype: grub
            data:
              BOOT_IMAGE: /boot/vmlinuz-6.8.0-1028-aws
```

**Matching Rules:**
1. **All** fields in the overlay key must be satisfied
2. Empty overlay field → matches anything (wildcard)
3. Empty query field → matches nothing (no match)
4. Version fields use semantic version equality with precision awareness

#### Usage Examples

```bash
# Basic recipe for Ubuntu with H100 GPU
eidos recipe --os ubuntu --gpu h100

# Full specification with all parameters
eidos recipe \
  --os ubuntu \
  --osv 24.04 \
  --kernel 6.8.0 \
  --service eks \
  --k8s v1.33.0 \
  --gpu gb200 \
  --intent training \
  --context \
  --format yaml \
  --output recipe.yaml

# Inference workload on GKE
eidos recipe --service gke --gpu a100 --intent inference
```

## Shared Infrastructure

### Collector Factory Pattern

The CLI uses the **Factory Pattern** for collector instantiation, enabling:
- **Testability**: Inject mock collectors for unit tests
- **Flexibility**: Easy to add new collector types
- **Encapsulation**: Hide collector creation complexity

```go
type Factory interface {
    CreateSystemDCollector() Collector
    CreateOSCollector() Collector
    CreateKubernetesCollector() Collector
    CreateGPUCollector() Collector
}
```

### Serializer Abstraction

Output formatting is abstracted through the `serializer.Serializer` interface:

```go
type Serializer interface {
    Serialize(data interface{}) error
}
```

Implementations:
- **JSON**: `encoding/json` with 2-space indent
- **YAML**: `gopkg.in/yaml.v3` 
- **Table**: `text/tabwriter` for columnar display

### Measurement Data Model

All collected data uses a unified `measurement.Measurement` structure:

```go
type Measurement struct {
    Type     Type      // os, k8s, systemd, gpu
    Subtypes []Subtype // Named collections of readings
}

type Subtype struct {
    Name    string                // grub, kmod, sysctl, server, image, etc.
    Data    map[string]Reading    // Key-value readings
    Context map[string]string     // Human-readable descriptions
}

type Reading struct {
    Value interface{}  // Actual value (int, string, bool, float64)
}
```

## Error Handling

### CLI Error Strategy

1. **Flag Validation**: User-friendly error messages for invalid flags
2. **Version Parsing**: Specific error types (ErrNegativeComponent, etc.)
3. **Collector Failures**: Log errors, continue with partial data where possible
4. **Serialization Errors**: Fatal - abort and report
5. **Exit Codes**: Non-zero exit code on any failure

### Example Error Messages

```bash
# Invalid version format
$ eidos recipe --osv -1.0
Error: error parsing recipe input parameter: os version cannot contain negative numbers: -1.0

# Unknown output format
$ eidos snapshot --format xml
Error: unknown output format: "xml"

# Missing required parameters
$ eidos recipe
# Still succeeds - generates base recipe with no overlays
```

## Performance Characteristics

### Snapshot Command

- **Parallel Collection**: All collectors run concurrently via `errgroup`
- **Typical Duration**: 100-500ms depending on cluster size
- **Memory Usage**: ~10-50MB for typical workloads
- **Scalability**: O(n) with number of pods/nodes for K8s collector

### Recipe Command

- **Store Loading**: Once per process (cached via `sync.Once`)
- **Typical Duration**: <10ms after initial load
- **Memory Usage**: ~5-10MB (embedded YAML + parsed structure)
- **Scalability**: O(m) with number of overlays (typically <100)

## Build Configuration

### Version Injection via ldflags

Build-time version information injection:

```makefile
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT ?= $(shell git rev-parse --short HEAD)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -X github.com/NVIDIA/cloud-native-stack/pkg/cli.version=$(VERSION)
LDFLAGS += -X github.com/NVIDIA/cloud-native-stack/pkg/cli.commit=$(COMMIT)
LDFLAGS += -X github.com/NVIDIA/cloud-native-stack/pkg/cli.date=$(DATE)

go build -ldflags="$(LDFLAGS)" -o bin/eidos ./cmd/eidos
```

## Testing Strategy

### Unit Tests
- Flag parsing and validation
- Version parsing and error handling
- Query building from command flags
- Serializer format selection

### Integration Tests
- Mock collectors for deterministic output
- Full command execution with fake factory
- Output format validation

### Example Test Structure

```go
func TestSnapshotCommand(t *testing.T) {
    // Create mock factory
    mockFactory := &MockFactory{
        k8s:     mockK8sCollector,
        systemd: mockSystemDCollector,
        os:      mockOSCollector,
        gpu:     mockGPUCollector,
    }
    
    // Execute snapshot with mock
    snapshotter := NodeSnapshotter{
        Factory: mockFactory,
        Serializer: &bytes.Buffer{},
    }
    
    err := snapshotter.Measure(ctx)
    assert.NoError(t, err)
}
```

## Dependencies

### External Libraries
- `github.com/urfave/cli/v3` - CLI framework
- `golang.org/x/sync/errgroup` - Concurrent error handling
- `gopkg.in/yaml.v3` - YAML parsing
- `log/slog` - Structured logging

### Internal Packages
- `pkg/collector` - System data collection
- `pkg/measurement` - Data model
- `pkg/recipe` - Recipe building
- `pkg/version` - Semantic versioning
- `pkg/serializer` - Output formatting
- `pkg/logging` - Logging configuration
- `pkg/snapshotter` - Snapshot orchestration

## Future Enhancements

### Potential Improvements
1. **Caching**: Cache snapshot results with TTL
2. **Differential Snapshots**: Compare two snapshots and show diff
3. **Filtering**: Allow filtering measurements by type/subtype
4. **Compression**: Compress large snapshots (gzip)
5. **Streaming**: Stream measurements for large datasets
6. **Validation**: Validate snapshots against expected schemas
7. **Import/Export**: Convert between different snapshot formats
8. **Plugins**: External collector plugins via RPC

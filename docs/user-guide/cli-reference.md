# CLI Reference

Complete reference for the `eidos` command-line interface.

## Overview

Eidos provides a three-step workflow for optimizing GPU infrastructure:

```
┌──────────────┐      ┌──────────────┐      ┌──────────────┐
│   Snapshot   │─────▶│    Recipe    │─────▶│    Bundle    │
└──────────────┘      └──────────────┘      └──────────────┘
```

**Step 1**: Capture system configuration  
**Step 2**: Generate optimization recipes  
**Step 3**: Create deployment bundles  

## Global Flags

Available for all commands:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--debug` | `-d` | bool | false | Enable debug logging |
| `--help` | `-h` | bool | false | Show help |
| `--version` | `-v` | bool | false | Show version |

## Commands

### eidos snapshot

Capture comprehensive system configuration including OS, GPU, Kubernetes, and SystemD settings.

**Synopsis:**
```shell
eidos snapshot [flags]
```

**Flags:**
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--output` | `-o` | string | stdout | Output file path |
| `--format` | `-f` | string | yaml | Output format: json, yaml, table |

**What it captures:**
- **SystemD Services**: containerd, docker, kubelet configurations
- **OS Configuration**: grub, kmod, sysctl, release info
- **Kubernetes**: server version, images, ClusterPolicy
- **GPU**: driver version, CUDA, MIG settings, hardware info

**Examples:**

```shell
# Output to stdout (YAML)
eidos snapshot

# Save to file (JSON)
eidos snapshot --output system.json --format json

# Debug mode
eidos --debug snapshot

# Table format (human-readable)
eidos snapshot --format table
```

**Output structure:**
```yaml
apiVersion: snapshot.dgxc.io/v1
kind: Snapshot
metadata:
  created: "2025-12-31T10:30:00Z"
  hostname: gpu-node-1
measurements:
  - type: SystemD
    subtypes: [...]
  - type: OS
    subtypes: [...]
  - type: K8s
    subtypes: [...]
  - type: GPU
    subtypes: [...]
```

---

### eidos recipe

Generate optimized configuration recipes from query parameters or captured snapshots.

**Synopsis:**
```shell
eidos recipe [flags]
```

**Modes:**

#### Query Mode
Generate recipes using direct system parameters:

**Flags:**
| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--os` | | string | OS family: ubuntu, rhel, cos |
| `--osv` | | string | OS version: 24.04, 22.04 |
| `--kernel` | | string | Kernel version: 6.8, 5.15 |
| `--service` | | string | K8s service: eks, gke, aks, self-managed |
| `--k8s` | | string | Kubernetes version: v1.33, 1.32 |
| `--gpu` | | string | GPU type: h100, gb200, a100, l40 |
| `--intent` | | string | Workload intent: training, inference |
| `--context` | | bool | Include context metadata in response |
| `--output` | `-o` | string | Output file (default: stdout) |
| `--format` | `-f` | string | Format: json, yaml, table (default: json) |

**Examples:**
```shell
# Basic recipe for Ubuntu on EKS with H100
eidos recipe --os ubuntu --service eks --gpu h100

# Full specification with context
eidos recipe \
  --os ubuntu \
  --osv 24.04 \
  --kernel 6.8 \
  --service eks \
  --k8s 1.33 \
  --gpu gb200 \
  --intent training \
  --context \
  --format yaml

# Save to file
eidos recipe --os ubuntu --gpu h100 --output recipe.yaml
```

#### Snapshot Mode
Generate recipes from captured snapshots:

**Flags:**
| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--snapshot` | `-f` | string | Path to snapshot file (required) |
| `--intent` | `-i` | string | Workload intent: training, inference |
| `--output` | `-o` | string | Output file (default: stdout) |
| `--format` | | string | Format: json, yaml, table (default: json) |
| `--context` | | bool | Include context metadata |

**Examples:**
```shell
# Generate recipe from snapshot
eidos recipe --snapshot system.yaml --intent training

# With custom output
eidos recipe -f system.yaml -i inference -o recipe.yaml --format yaml
```

**Output structure:**
```yaml
apiVersion: recipe.dgxc.io/v1
kind: Recipe
metadata:
  created: "2025-12-31T10:30:00Z"
request:
  os: ubuntu
  gpu: h100
  service: eks
matchedRules:
  - "OS: ubuntu, GPU: h100, Service: eks"
measurements:
  - type: K8s
    subtypes: [...]
  - type: GPU
    subtypes: [...]
```

---

### eidos bundle

Generate deployment-ready bundles from recipes containing Helm values, manifests, scripts, and documentation.

**Synopsis:**
```shell
eidos bundle [flags]
```

**Flags:**
| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--recipe` | `-f` | string | Path to recipe file (required) |
| `--bundlers` | `-b` | string[] | Bundler types to execute (repeatable) |
| `--output` | `-o` | string | Output directory (default: current dir) |
| `--format` | | string | Summary format: json, yaml, table (default: json) |

**Available bundlers:**
- `gpu-operator` - NVIDIA GPU Operator deployment bundle
- `network-operator` - NVIDIA Network Operator deployment bundle

**Behavior:**
- If `--bundlers` is omitted, **all registered bundlers** execute
- Bundlers run in **parallel** by default
- Each bundler creates a subdirectory in the output directory

**Examples:**
```shell
# Generate all bundles
eidos bundle --recipe recipe.yaml --output ./bundles

# Generate specific bundler only
eidos bundle -f recipe.yaml -b gpu-operator -o ./deployment

# Multiple specific bundlers
eidos bundle -f recipe.yaml \
  -b gpu-operator \
  -b network-operator \
  -o ./bundles
```

**Bundle structure** (GPU Operator example):
```
gpu-operator/
├── values.yaml                    # Helm chart configuration
├── manifests/
│   └── clusterpolicy.yaml        # ClusterPolicy CR
├── scripts/
│   ├── install.sh                # Installation script
│   └── uninstall.sh              # Cleanup script
├── README.md                      # Deployment guide
└── checksums.txt                  # SHA256 checksums
```

**Deploying a bundle:**
```shell
# Navigate to bundle
cd bundles/gpu-operator

# Review configuration
cat values.yaml
cat README.md

# Verify integrity
sha256sum -c checksums.txt

# Deploy to cluster
chmod +x scripts/install.sh
./scripts/install.sh
```

---

## Complete Workflow Example

```shell
# Step 1: Capture system configuration
eidos snapshot --output snapshot.yaml

# Step 2: Generate optimized recipe for training workloads
eidos recipe \
  --snapshot snapshot.yaml \
  --intent training \
  --output recipe.yaml

# Step 3: Create deployment bundle
eidos bundle \
  --recipe recipe.yaml \
  --bundlers gpu-operator \
  --output ./deployment

# Step 4: Deploy to cluster
cd deployment/gpu-operator
./scripts/install.sh

# Step 5: Verify deployment
kubectl get pods -n gpu-operator
kubectl logs -n gpu-operator -l app=nvidia-operator-validator
```

## Shell Completion

Generate shell completion scripts:

```shell
# Bash
eidos completion bash

# Zsh
eidos completion zsh

# Fish
eidos completion fish

# PowerShell
eidos completion powershell
```

**Installation:**

**Bash:**
```shell
source <(eidos completion bash)
# Or add to ~/.bashrc for persistence
echo 'source <(eidos completion bash)' >> ~/.bashrc
```

**Zsh:**
```shell
source <(eidos completion zsh)
# Or add to ~/.zshrc
echo 'source <(eidos completion zsh)' >> ~/.zshrc
```

## Environment Variables

Eidos respects standard environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `KUBECONFIG` | Path to Kubernetes config file | `~/.kube/config` |
| `LOG_LEVEL` | Logging level: debug, info, warn, error | info |
| `NO_COLOR` | Disable colored output | false |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |
| 3 | File I/O error |
| 4 | Kubernetes connection error |
| 5 | Recipe generation error |

## Common Usage Patterns

### Quick Recipe Generation
```shell
eidos recipe --os ubuntu --gpu h100 | jq '.measurements[]'
```

### Save All Steps
```shell
eidos snapshot -o snapshot.yaml
eidos recipe -f snapshot.yaml -i training -o recipe.yaml
eidos bundle -f recipe.yaml -o ./bundles
```

### JSON Processing
```shell
# Extract GPU driver version from recipe
eidos recipe --os ubuntu --gpu h100 --format json | \
  jq -r '.measurements[] | select(.type=="GPU") | 
         .subtypes[] | select(.subtype=="driver") | 
         .data.version'
```

### Multiple Environments
```shell
# Generate recipes for different cloud providers
for service in eks gke aks; do
  eidos recipe --os ubuntu --service $service --gpu h100 \
    --output recipe-${service}.yaml
done
```

## Troubleshooting

### Snapshot Fails
```shell
# Check GPU drivers
nvidia-smi

# Check Kubernetes access
kubectl cluster-info

# Run with debug
eidos --debug snapshot
```

### Recipe Not Found
```shell
# Query parameters may not match any overlay
# Try broader query:
eidos recipe --os ubuntu --gpu h100
```

### Bundle Generation Fails
```shell
# Verify recipe file
cat recipe.yaml

# Check bundler is valid
eidos bundle --help  # Shows available bundlers

# Run with debug
eidos --debug bundle -f recipe.yaml -b gpu-operator
```

## See Also

- [Installation Guide](installation.md) - Install Eidos
- [Agent Deployment](agent-deployment.md) - Kubernetes agent setup
- [API Reference](../integration/api-reference.md) - Programmatic access
- [Architecture Docs](../architecture/README.md) - Internal architecture

# Cloud Native Stack

Cloud Native Stack (CNS) provides tooling and comprehensive documentation to help you deploy, validate, and operate optimized AI workloads in your GPU-accelerated Kubernetes clusters. The project includes:

- **Documentation** – Installation guides, playbooks, optimizations, and troubleshooting for GPU infrastructure
- **`eidos` CLI** – Command-line tool for system snapshots and recipe generation
- **Eidos Agent** – Kubernetes job for automated cluster configuration and optimization

## Quick Start

### Install the `eidos` CLI

The `eidos` CLI are built with each release. You can find the latest release [here](https://github.com/mchmarny/cloud-native-stack/releases/latest). Download and install the latest version:

```shell
curl -sfL https://raw.githubusercontent.com/mchmarny/cloud-native-stack/refs/heads/main/installer | bash -s --
```

Verify installation:

```shell
eidos --version
```

### CLI Commands

#### Snapshot System Configuration

Capture a comprehensive snapshot of your system including CPU/GPU settings, kernel parameters, systemd services, and Kubernetes configuration:

```shell
# Output to stdout (JSON)
eidos snapshot

# Save to file (YAML format)
eidos snapshot --output system.yaml --format yaml

# Table format for human readability
eidos snapshot --format table
```

The snapshot includes:
- CPU and GPU hardware details
- GRUB boot parameters
- Kubernetes cluster configuration
- Loaded kernel modules
- Sysctl kernel parameters
- SystemD service configurations

#### Generate Configuration Recipe

Generate optimized configuration recipes based on your environment:

```shell
# Basic recipe for Ubuntu on EKS with H100 GPUs
eidos recipe --os ubuntu --service eks --gpu h100

# Full specification with context
eidos recipe \
  --os ubuntu \
  --osv 24.04 \
  --kernel 5.15.0 \
  --service eks \
  --k8s v1.28.0 \
  --gpu gb200 \
  --intent training \
  --context \
  --format yaml
```

**Available flags:**
- `--os` – Operating system (ubuntu, cos, etc.)
- `--osv` – OS version (e.g., 24.04)
- `--kernel` – Kernel version
- `--service` – Kubernetes service (eks, gke, aks, self-managed)
- `--k8s` – Kubernetes version
- `--gpu` – GPU type (h100, gb200, etc.)
- `--intent` – Workload intent (training, inference)
- `--context` – Include metadata in response
- `--format` – Output format (json, yaml, table)
- `--output` – Save to file (default: stdout)

### Deploy the Eidos Agent

The Eidos Agent runs as a Kubernetes Job to automatically capture cluster configuration snapshots. This is useful for auditing, troubleshooting, and configuration management.

#### Prerequisites

- Kubernetes cluster with GPU nodes
- `kubectl` configured with cluster access
- GPU Operator installed (agent runs in `gpu-operator` namespace)

#### Installation

1. Apply the required RBAC permissions and service account:

```shell
kubectl apply -f https://raw.githubusercontent.com/mchmarny/cloud-native-stack/main/deployments/eidos-agent/1-deps.yaml
```

2. Deploy the agent job:

```shell
kubectl apply -f https://raw.githubusercontent.com/mchmarny/cloud-native-stack/main/deployments/eidos-agent/2-job.yaml
```

#### Customization

Before deploying, you may need to customize the Job manifest:

**Node Selection** – Update `nodeSelector` to target specific GPU nodes:
```yaml
nodeSelector:
  nodeGroup: your-gpu-node-group
```

**Tolerations** – Adjust tolerations for your node taints:
```yaml
tolerations:
  - key: nvidia.com/gpu
    operator: Exists
    effect: NoSchedule
```

**Image Version** – Use a specific version:
```yaml
image: ghcr.io/mchmarny/eidos-api-server:v0.5.16
```

#### View Agent Output

Check job status:
```shell
kubectl get jobs -n gpu-operator
```

View snapshot output:
```shell
kubectl logs -n gpu-operator job/eidos
```

The agent outputs a YAML snapshot of the cluster node configuration to stdout.

## Documentation

Comprehensive deployment and operations guides:

- **[Installation Guides](docs/install-guides)** – Step-by-step setup for various platforms
- **[Playbooks](docs/playbooks)** – Ansible automation for CNS deployment
- **[Optimizations](docs/optimizations)** – Hardware-specific performance tuning
- **[Troubleshooting](docs/troubleshooting)** – Common issues and solutions
- **[Full Documentation](docs/README.md)** – Complete reference

## Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Development setup and workflow
- Code quality standards
- Pull request process
- Building and testing locally

## Support

- **Releases**: [GitHub Releases](https://github.com/NVIDIA/cloud-native-stack/releases)
- **Issues**: [GitHub Issues](https://github.com/NVIDIA/cloud-native-stack/issues)
- **Questions**: Open a discussion or issue on GitHub

# Examples

This directory contains example snapshots, recipes, and bundles demonstrating the Cloud Native Stack workflow.

## Directory Structure

```
examples/
├── bundles/          # Generated deployment bundles
├── recipes/          # Optimized configuration recipes  
├── snapshots/        # System configuration snapshots
├── gb200-h100-comp.md
├── gb200.yaml
├── h100.yaml
└── recipe.yaml
```

## Snapshots

Example system configuration snapshots captured from GPU clusters:

### GB200 System ([gb200.yaml](gb200.yaml))

Complete snapshot from a GB200 NVL72 system showing:
- OS configuration (Ubuntu 24.04)
- GPU hardware details (GB200 with NVLink)
- Kubernetes cluster configuration (EKS 1.33)
- SystemD service states
- Deployed container images

**Use case:** Generate recipes optimized for GB200 training workloads

```bash
eidos recipe --snapshot examples/gb200.yaml --intent training
```

### H100 System ([h100.yaml](h100.yaml))

Snapshot from an H100 GPU cluster with:
- OS configuration (Ubuntu 22.04)
- H100 GPU specifications
- Kubernetes configuration (GKE 1.32)
- GPU Operator ClusterPolicy settings

**Use case:** Generate recipes optimized for H100 inference workloads

```bash
eidos recipe --snapshot examples/h100.yaml --intent inference
```

## Recipes

Optimized configuration recipes generated from snapshots or query parameters:

### General Recipe ([recipe.yaml](recipe.yaml))

Example recipe showing the recipe structure:
- Matched overlay rules
- System measurements (OS, K8s, GPU, SystemD)
- Optimized configuration values
- Context metadata

**Generated from:**
```bash
eidos recipe --os ubuntu --service eks --gpu h100 --intent training --context
```

**Use case:** Understand recipe structure or generate bundles

```bash
eidos bundle --recipe examples/recipe.yaml --output ./my-bundles
```

## Bundles

Generated deployment packages ready for production use:

### GPU Operator Bundle ([bundles/gpu-operator/](bundles/gpu-operator/))

Complete GPU Operator deployment bundle containing:
- `values.yaml` – Helm chart configuration
- `manifests/clusterpolicy.yaml` – ClusterPolicy CR
- `scripts/install.sh` – Automated installation script
- `scripts/uninstall.sh` – Cleanup script
- `README.md` – Deployment instructions
- `checksums.txt` – SHA256 checksums for verification

**Generated from:**
```bash
eidos bundle --recipe examples/recipe.yaml --bundlers gpu-operator --output examples/bundles
```

**Deploy:**
```bash
cd examples/bundles/gpu-operator
chmod +x scripts/install.sh
./scripts/install.sh
```

## Comparisons

### GB200 vs H100 Comparison ([gb200-h100-comp.md](gb200-h100-comp.md))

Detailed comparison document showing configuration differences between GB200 and H100 systems:
- Hardware specifications
- Driver and CUDA versions
- Network configuration (NVLink vs PCIe)
- Memory topology
- Recommended settings per GPU type

**Use case:** Understand platform-specific optimizations

## Complete Workflow Example

End-to-end example using the provided files:

```bash
# 1. Review example snapshot
cat examples/gb200.yaml

# 2. Generate optimized recipe for training
eidos recipe \
  --snapshot examples/gb200.yaml \
  --intent training \
  --format yaml \
  --output my-recipe.yaml

# 3. Compare with provided recipe
diff my-recipe.yaml examples/recipe.yaml

# 4. Generate deployment bundle
eidos bundle \
  --recipe my-recipe.yaml \
  --bundlers gpu-operator \
  --output ./my-deployment

# 5. Review generated bundle
tree my-deployment/gpu-operator/
cat my-deployment/gpu-operator/README.md

# 6. Verify checksums
cd my-deployment/gpu-operator
sha256sum -c checksums.txt

# 7. Deploy to cluster
./scripts/install.sh
```

## Generate Your Own Examples

### Capture Snapshot

From your GPU cluster:

```bash
# Capture snapshot to file
eidos snapshot --output my-snapshot.yaml

# Or deploy agent to Kubernetes
kubectl apply -f https://raw.githubusercontent.com/mchmarny/cloud-native-stack/main/deployments/eidos-agent/1-deps.yaml
kubectl apply -f https://raw.githubusercontent.com/mchmarny/cloud-native-stack/main/deployments/eidos-agent/2-job.yaml
kubectl logs -n gpu-operator job/eidos > my-snapshot.yaml
```

### Generate Recipe

From snapshot or query:

```bash
# From snapshot
eidos recipe --snapshot my-snapshot.yaml --intent training --output my-recipe.yaml

# From query parameters
eidos recipe \
  --os ubuntu \
  --osv 24.04 \
  --service eks \
  --k8s 1.33 \
  --gpu gb200 \
  --intent inference \
  --output my-recipe.yaml
```

### Create Bundle

From recipe:

```bash
# Generate all bundlers
eidos bundle --recipe my-recipe.yaml --output ./bundles

# Generate specific bundler
eidos bundle \
  --recipe my-recipe.yaml \
  --bundlers gpu-operator \
  --output ./bundles
```

## See Also

- **[CLI Reference](../docs/user-guide/cli-reference.md)** – Complete CLI documentation
- **[Data Flow](../docs/integration/data-flow.md)** – Understanding snapshots, recipes, and bundles
- **[Bundler Development](../docs/architecture/bundler-development.md)** – How to create custom bundlers
- **[API Reference](../docs/integration/api-reference.md)** – Using the API for recipe generation

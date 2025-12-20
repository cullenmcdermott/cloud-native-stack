# H100 vs GB200 Node Snapshot Comparison Report

## Files Compared

| System | Source | Cluster      | Node             |
|--------|--------|--------------|------------------|
| H100   | AWS    | `a8c39f8a`   | `ip-10-0-158-18` |
| GB200  | AWS    | `validation` | `ip-10-0-157-11` |

Both snapshots use `snapshot.dgxc.io/v1`, version `v0.4.1`.

> Meaningful config and capability diffs only. Ignores order, timestamps, and other expected runtime noise.

⸻

## 1. High-Level Summary

| Category | Classification | Notes |
|----------|----------------|-------|
| Kernel & Boot | Different | Same kernel family (6.8 AWS), different patch level and flags |
| CPU Architecture | Different | H100 is x86_64; GB200 is ARM64 |
| Crypto Acceleration | Different | Architecture-specific crypto modules |
| NUMA / Memory Policy | Different | Explicit NUMA tuning only on GB200 |
| Kubernetes | Missing in GB200 snapshot | Version reported only on H100 |
| Container Runtime | Equivalent | containerd configuration aligned |
| GPU Stack | Equivalent | NVIDIA + GDR present on both |
| Networking / RDMA | Equivalent | EFA + RDMA stacks aligned |
| Docker | Equivalent (disabled) | Inactive on both |
| Kubelet systemd unit | Equivalent (inactive) | Inactive on both |


⸻

## 2. Kernel & Boot Configuration (Grub)

### Kernel Version

| System | Kernel Version |
|--------|----------------|
| H100 | 6.8.0-1024-aws |
| GB200 | 6.8.0-1028-aws |

**Classification:** Patch-level skew only; same kernel line.

### Boot Flags – Real Differences

| Flag | H100 | GB200 |
|------|------|-------|
| init_on_alloc | not set | 0 |
| numa_balancing | default | disable |
| hugepages | 5128 | 5128 |
| hugepagesz | 2M | 2M |
| nokaslr | enabled | enabled |

**Interpretation:** GB200 explicitly disables NUMA auto-balancing and init-on-alloc, indicating tighter control over memory placement and determinism. H100 relies on kernel defaults.

⸻

## 3. CPU Architecture & Crypto Stack

### Architecture Evidence

**H100 (x86_64-oriented modules):**
- aesni_intel
- sha256_ssse3
- ghash_clmulni_intel

**GB200 (ARM64-oriented modules):**
- aes_ce, sha*_ce
- sm3, sm4
- polyval_ce

**Classification:** Fundamental architectural difference. Expected and correct for GB200-class systems.

⸻

## 4. Kernel Module Inventory (KMod)

### GPU / NVIDIA Stack

| Module | H100 | GB200 |
|--------|------|-------|
| nvidia | ✓ | ✓ |
| nvidia_uvm | ✓ | ✓ |
| nvidia_modeset | ✓ | ✓ |
| gdrdrv | ✓ | ✓ |
| ecc | ✓ | ✓ |

**Assessment:** No gap. GPU driver and GDR plumbing are aligned.

### Networking & RDMA

Both snapshots include:
- efa
- ib_core, ib_uverbs
- rdma_cm, iw_cm
- rpcrdma, sunrpc

**Assessment:** No meaningful difference. RDMA and EFA parity is good.

### Filesystem / Storage Stack

- **H100:** Includes full Lustre client stack (lustre, lmv, mdc, osc, ptlrpc, etc.)
- **GB200:** Lustre modules not present

**Classification:** True functional gap. H100 nodes are Lustre-capable; GB200 nodes are not configured with Lustre support.

⸻

## 5. Kubernetes Presence

| Aspect | H100 | GB200 |
|--------|------|-------|
| Kubernetes metadata | Present | Missing |
| Reported version | v1.30.14-eks | N/A |

**Classification:** Observability gap, not necessarily a system gap. Either kubelet is not installed/running on GB200 at snapshot time, or Kubernetes metadata collection was skipped.

This is the single largest structural difference between the two snapshots.

⸻

## 6. systemd: containerd

### containerd.service

- Active and enabled on both nodes
- Identical drop-ins (999-tuning-tuning.conf)
- Same cgroup delegation, limits, restart policy

Observed differences are limited to runtime counters:
- CPUUsageNSec
- MemoryCurrent
- TasksCurrent

**Classification:** Runtime variance only. No configuration drift.

⸻

## 7. Docker & Kubelet Units

| Unit | H100 | GB200 |
|------|------|-------|
| docker.service | inactive / not-found | inactive / not-found |
| kubelet.service | inactive | inactive |

**Assessment:** No difference.



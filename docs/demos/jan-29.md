# CNS Jan 29th Updates

Setup

```shell
rm -fr ~/Downloads/demo1 ~/Downloads/demo2
mkdir ~/Downloads/demo1 ~/Downloads/demo2
cp -r ~/dev/cloud-native-stack/examples/data ~/Downloads/demo2
```

Demo `v0.31.1` (latest)

## For contributor (works in the repo)

Fully Declarative Recipe Support  

https://github.com/mchmarny/cloud-native-stack/blob/main/docs/architecture/component.md#quick-start

Yuan: GKE/COS Image
Val/Patrick: Kicking off later today component/recipe/values/manifest review
Focus: Functional capabilities > data (optimized/validated/reproducible)

## For User (uses released artifacts)

`cd ~/Downloads/demo1`

### Recipe Criteria

Basic, using flags:

```shell
cnsctl recipe \
  --service eks \
  --accelerator gb200 \
  --os ubuntu \
  --intent training \
  --output recipe.yaml
```

From criteria file:

```shell
cat > /tmp/criteria.yaml << 'EOF'
kind: recipeCriteria
apiVersion: cns.nvidia.com/v1alpha1
metadata:
  name: gb200-eks-training
spec:
  service: eks
  accelerator: gb200
  os: ubuntu
  intent: training
EOF
```

> May evolve. Nathan, defining the MVP scope of criteria we need to support

Generate recipe from criteria file:

```shell
cnsctl recipe --criteria /tmp/criteria.yaml | yq .
```

Flags still override criteria file values:

```shell
cnsctl recipe --criteria /tmp/criteria.yaml --service gke --os cos | yq .criteria
```

### Self-hosted API

Same criteria support
Configurable allowed list support on each criterion:

```shell
curl -s -X POST "https://cns.dgxc.io/v1/recipe" \
  -H "Content-Type: application/x-yaml" \
  -d 'kind: recipeCriteria
apiVersion: cns.nvidia.com/v1alpha1
metadata:
  name: my-training-recipe
spec:
  service: eks
  accelerator: gb200
  intent: training' | jq .criteria
```

Error on `h100`

```shell
curl -s -X POST "https://cns.dgxc.io/v1/recipe" \
  -H "Content-Type: application/x-yaml" \
  -d 'kind: recipeCriteria
apiVersion: cns.nvidia.com/v1alpha1
metadata:
  name: my-training-recipe
spec:
  service: eks
  accelerator: h100
  intent: training' | jq .
```

### Reproducible Bundler    

Bundle from Recipe:

```shell
cnsctl bundle \
  --recipe recipe.yaml \
  --output ./bundle \
  --system-node-selector nodeGroup=system-pool \
  --accelerated-node-selector nodeGroup=customer-gpu \
  --accelerated-node-toleration nvidia.com/gpu=present:NoSchedule
```

Bundle content: 

```shell
cd ./bundle && tree .
```

Each file byte-level reproducible: 

```shell
shasum -a 256 -c checksums.txt
```

Check the umbrella chart: 

```shell
yq . Chart.yaml
```

Prep the deployment: 

```shell
helm dependency update
tree .
```

> NVS == OCI Repo Type == Digest

Validate Bundle: 

```shell
helm lint . -n cns-system
```

### OCI Image (Open Container Initiative, NOT Oracle Cloud Infrastructure)

Back to recipe: 

```shell
cd ../
```

Bundle as an OCI image:

```shell
cnsctl bundle \
  --recipe recipe.yaml \
  --output oci://ghcr.io/mchmarny/cns-bundle \
  --image-refs .digest
```

Review manifest: 

```shell
crane manifest "ghcr.io/mchmarny/cns-bundle@$(cat .digest)" | jq .
```

Unpack the image: 

```shell
skopeo copy "docker://ghcr.io/mchmarny/cns-bundle@$(cat .digest)" oci:image-oci
mkdir -p ./cns-unpacked
oras pull --oci-layout "image-oci@$(cat .digest)" -o ./cns-unpacked
tree ./cns-unpacked
```

### Runtime Data Support

In `~/Downloads/demo2`

Generate recipe with external data:

```shell
cnsctl recipe \
  --criteria /tmp/criteria.yaml \
  --data ./data \
  --output recipe.yaml
```

`merged component registries: embedded_components=7 external_components=1 merged_components=8`

Documented data struct (same as in the repo):

```shell
tree ./data 
```

Check registry: 

```shell
yq . ./data/registry.yaml
```

Check Recipe: 

Now combine it all together to generate bundles:

```shell
cnsctl bundle \
  --recipe recipe.yaml \
  --output ./bundle \
  --data ./data \
  --deployer argocd \
  --output oci://ghcr.io/mchmarny/cns-argo-bundle \
  --system-node-selector nodeGroup=system-pool \
  --accelerated-node-selector nodeGroup=customer-gpu \
  --accelerated-node-toleration nvidia.com/gpu=present:NoSchedule
```

Local files (App of Apps, Sub-Apps)

```shell
tree ./bundle
```

> To help debug overlay issues the `--debug` flag gives all the composition details

## Summary (What we'd seen)

* Fully declarative recipe development (No Go code required)
* Criteria resource as input param (Extendability options)
* Allowed lists supported in Self-hosted API (Enterprise friendly)
* External data support on recipe and bundle gen (internal-only components)
* Bundle output to OCI image (digest-level reproducibility)

## Links

For User: 
* [Installation Guide](https://github.com/mchmarny/cloud-native-stack/blob/main/docs/user-guide/installation.md)
* [CLI Reference](https://github.com/mchmarny/cloud-native-stack/blob/main/docs/user-guide/cli-reference.md)
* [API Reference](https://github.com/mchmarny/cloud-native-stack/blob/main/docs/user-guide/api-reference.md)

For Contributor: 
* [Component/Recipe](https://github.com/mchmarny/cloud-native-stack/blob/main/docs/architecture/component.md)
* [Data Reference](https://github.com/mchmarny/cloud-native-stack/blob/main/pkg/recipe/data/README.md)

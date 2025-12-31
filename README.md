# Cloud Native Stack

Cloud Native Stack (CNS) provides tooling and comprehensive documentation to help you deploy, validate, and operate optimized AI workloads in your GPU-accelerated Kubernetes clusters:

- **CLI (Eidos)** – Three-step workflow: capture system snapshots, generate optimization recipes, and create deployment bundles
- **API** – REST API for recipe generation and integration with automation pipelines
- **Agent** – Kubernetes job for automated cluster snapshot collection

**Note**: The documentation related to the previous version of the Cloud Native Stack project (manual installation guides, playbooks, and optimizations for GPU infrastructure) are all located in [docs/v1](docs/v1).

## Documentation

### For Users

Get started with installing and using Cloud Native Stack:

- **[Installation Guide](docs/user-guide/installation.md)** – Install the eidos CLI (automated script, manual, or build from source)
- **[CLI Reference](docs/user-guide/cli-reference.md)** – Complete command reference with examples
- **[Agent Deployment](docs/user-guide/agent-deployment.md)** – Deploy the Kubernetes agent for automated snapshots

### For Developers

Learn how to contribute and understand the architecture:

- **[Contributing Guide](CONTRIBUTING.md)** – Development setup, testing, and PR process
- **[Architecture Overview](docs/architecture/README.md)** – System design and components
- **[Bundler Development](docs/architecture/bundler-development.md)** – How to create new bundlers
- **[Data Architecture](docs/architecture/data.md)** – Recipe data model and query matching

### For Integrators

Integrate Cloud Native Stack into your infrastructure automation:

- **[API Reference](docs/integration/api-reference.md)** – REST API endpoints and usage examples
- **[Data Flow](docs/integration/data-flow.md)** – Understanding snapshots, recipes, and bundles
- **[Automation Guide](docs/integration/automation.md)** – CI/CD integration patterns
- **[Kubernetes Deployment](docs/integration/kubernetes-deployment.md)** – Self-hosted API server setup

### Additional Resources

Platform-specific deployment and optimization guides:

- **[Installation Guides](docs/v1/install-guides)** – Step-by-step setup for various platforms
- **[Playbooks](docs/v1/playbooks)** – Ansible automation for CNS deployment
- **[Optimizations](docs/v1/optimizations)** – Hardware-specific performance tuning
- **[Troubleshooting](docs/v1/troubleshooting)** – Common issues and solutions
- **[Full Documentation](docs/v1/README.md)** – Complete legacy documentation

## Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Development setup and workflow
- Code quality standards
- Pull request process
- Building and testing locally

## Support

- **Security**: [Project and Artifact Security](docs/SECURITY.md)
- **Releases**: [GitHub Releases](https://github.com/NVIDIA/cloud-native-stack/releases)
- **Issues**: [GitHub Issues](https://github.com/NVIDIA/cloud-native-stack/issues)
- **Questions**: Open a discussion or issue on GitHub

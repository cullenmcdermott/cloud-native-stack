package collector

import (
	"github.com/NVIDIA/cloud-native-stack/pkg/collector/gpu"
	"github.com/NVIDIA/cloud-native-stack/pkg/collector/k8s"
	"github.com/NVIDIA/cloud-native-stack/pkg/collector/os"
	"github.com/NVIDIA/cloud-native-stack/pkg/collector/systemd"
)

// Factory creates collectors with their dependencies.
// This interface enables dependency injection for testing.
type Factory interface {
	GetVersion() string
	CreateSystemDCollector() Collector
	CreateOSCollector() Collector
	CreateKubernetesCollector() Collector
	CreateGPUCollector() Collector
}

// Option defines a configuration option for DefaultFactory.
type Option func(*DefaultFactory)

// WithSystemDServices configures the systemd services to monitor.
func WithSystemDServices(services []string) Option {
	{
		return func(f *DefaultFactory) {
			f.SystemDServices = services
		}
	}
}

// WithVersion sets the version for the factory.
func WithVersion(version string) Option {
	return func(f *DefaultFactory) {
		f.Version = version
	}
}

// DefaultFactory creates collectors with production dependencies.
type DefaultFactory struct {
	SystemDServices []string
	Version         string
}

// NewDefaultFactory creates a factory with default settings.
func NewDefaultFactory(opts ...Option) *DefaultFactory {
	f := &DefaultFactory{
		SystemDServices: []string{
			"containerd.service",
			"docker.service",
			"kubelet.service",
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(f)
	}

	return f
}

// GetVersion returns the factory version.
func (f *DefaultFactory) GetVersion() string {
	return f.Version
}

// CreateSMICollector creates an GPU collector.
func (f *DefaultFactory) CreateGPUCollector() Collector {
	return &gpu.Collector{}
}

// CreateSystemDCollector creates a systemd collector.
func (f *DefaultFactory) CreateSystemDCollector() Collector {
	return &systemd.Collector{
		Services: f.SystemDServices,
	}
}

// CreateGrubCollector creates a GRUB collector.
func (f *DefaultFactory) CreateOSCollector() Collector {
	return &os.Collector{}
}

// CreateKubernetesCollector creates a Kubernetes API collector.
func (f *DefaultFactory) CreateKubernetesCollector() Collector {
	return &k8s.Collector{}
}

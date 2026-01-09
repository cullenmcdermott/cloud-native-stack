// Package recipe provides recipe building and matching functionality.
package recipe

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// criteriaAnyValue is the wildcard value for criteria matching.
const criteriaAnyValue = "any"

// CriteriaServiceType represents the Kubernetes service/platform type for criteria.
type CriteriaServiceType string

// CriteriaServiceType constants for supported Kubernetes services.
const (
	CriteriaServiceAny CriteriaServiceType = "any"
	CriteriaServiceEKS CriteriaServiceType = "eks"
	CriteriaServiceGKE CriteriaServiceType = "gke"
	CriteriaServiceAKS CriteriaServiceType = "aks"
	CriteriaServiceOKE CriteriaServiceType = "oke"
)

// ParseCriteriaServiceType parses a string into a CriteriaServiceType.
func ParseCriteriaServiceType(s string) (CriteriaServiceType, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", criteriaAnyValue, "self-managed", "self", "vanilla":
		return CriteriaServiceAny, nil
	case "eks":
		return CriteriaServiceEKS, nil
	case "gke":
		return CriteriaServiceGKE, nil
	case "aks":
		return CriteriaServiceAKS, nil
	case "oke":
		return CriteriaServiceOKE, nil
	default:
		return CriteriaServiceAny, fmt.Errorf("invalid service type: %s", s)
	}
}

// GetCriteriaServiceTypes returns all supported service types sorted alphabetically.
func GetCriteriaServiceTypes() []string {
	return []string{"aks", "eks", "gke", "oke"}
}

// CriteriaFabricType represents the network fabric type.
type CriteriaFabricType string

// CriteriaFabricType constants for supported network fabrics.
const (
	CriteriaFabricAny CriteriaFabricType = "any"
	CriteriaFabricEFA CriteriaFabricType = "efa"
	CriteriaFabricIB  CriteriaFabricType = "ib"
)

// ParseCriteriaFabricType parses a string into a CriteriaFabricType.
func ParseCriteriaFabricType(s string) (CriteriaFabricType, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", criteriaAnyValue:
		return CriteriaFabricAny, nil
	case "efa":
		return CriteriaFabricEFA, nil
	case "ib", "infiniband":
		return CriteriaFabricIB, nil
	default:
		return CriteriaFabricAny, fmt.Errorf("invalid fabric type: %s", s)
	}
}

// GetCriteriaFabricTypes returns all supported fabric types sorted alphabetically.
func GetCriteriaFabricTypes() []string {
	return []string{"efa", "ib"}
}

// CriteriaAcceleratorType represents the GPU/accelerator type.
type CriteriaAcceleratorType string

// CriteriaAcceleratorType constants for supported accelerators.
const (
	CriteriaAcceleratorAny   CriteriaAcceleratorType = "any"
	CriteriaAcceleratorH100  CriteriaAcceleratorType = "h100"
	CriteriaAcceleratorGB200 CriteriaAcceleratorType = "gb200"
	CriteriaAcceleratorA100  CriteriaAcceleratorType = "a100"
	CriteriaAcceleratorL40   CriteriaAcceleratorType = "l40"
)

// ParseCriteriaAcceleratorType parses a string into a CriteriaAcceleratorType.
func ParseCriteriaAcceleratorType(s string) (CriteriaAcceleratorType, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", criteriaAnyValue:
		return CriteriaAcceleratorAny, nil
	case "h100":
		return CriteriaAcceleratorH100, nil
	case "gb200":
		return CriteriaAcceleratorGB200, nil
	case "a100":
		return CriteriaAcceleratorA100, nil
	case "l40":
		return CriteriaAcceleratorL40, nil
	default:
		return CriteriaAcceleratorAny, fmt.Errorf("invalid accelerator type: %s", s)
	}
}

// GetCriteriaAcceleratorTypes returns all supported accelerator types sorted alphabetically.
func GetCriteriaAcceleratorTypes() []string {
	return []string{"a100", "gb200", "h100", "l40"}
}

// CriteriaIntentType represents the workload intent.
type CriteriaIntentType string

// CriteriaIntentType constants for supported workload intents.
const (
	CriteriaIntentAny       CriteriaIntentType = "any"
	CriteriaIntentTraining  CriteriaIntentType = "training"
	CriteriaIntentInference CriteriaIntentType = "inference"
)

// ParseCriteriaIntentType parses a string into a CriteriaIntentType.
func ParseCriteriaIntentType(s string) (CriteriaIntentType, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", criteriaAnyValue:
		return CriteriaIntentAny, nil
	case "training":
		return CriteriaIntentTraining, nil
	case "inference":
		return CriteriaIntentInference, nil
	default:
		return CriteriaIntentAny, fmt.Errorf("invalid intent type: %s", s)
	}
}

// GetCriteriaIntentTypes returns all supported intent types sorted alphabetically.
func GetCriteriaIntentTypes() []string {
	return []string{"inference", "training"}
}

// CriteriaOSType represents an operating system type.
type CriteriaOSType string

// CriteriaOSType constants for supported operating systems.
const (
	CriteriaOSAny         CriteriaOSType = "any"
	CriteriaOSUbuntu      CriteriaOSType = "ubuntu"
	CriteriaOSRHEL        CriteriaOSType = "rhel"
	CriteriaOSCOS         CriteriaOSType = "cos"
	CriteriaOSAmazonLinux CriteriaOSType = "amazonlinux"
)

// ParseCriteriaOSType parses a string into a CriteriaOSType.
func ParseCriteriaOSType(s string) (CriteriaOSType, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", criteriaAnyValue:
		return CriteriaOSAny, nil
	case "ubuntu":
		return CriteriaOSUbuntu, nil
	case "rhel":
		return CriteriaOSRHEL, nil
	case "cos":
		return CriteriaOSCOS, nil
	case "amazonlinux", "al2", "al2023":
		return CriteriaOSAmazonLinux, nil
	default:
		return CriteriaOSAny, fmt.Errorf("invalid os type: %s", s)
	}
}

// GetCriteriaOSTypes returns all supported OS types sorted alphabetically.
func GetCriteriaOSTypes() []string {
	return []string{"amazonlinux", "cos", "rhel", "ubuntu"}
}

// Criteria represents the input parameters for recipe matching.
// All fields are optional and default to "any" if not specified.
type Criteria struct {
	// Service is the Kubernetes service type (eks, gke, aks, oke, self-managed).
	Service CriteriaServiceType `json:"service,omitempty" yaml:"service,omitempty"`

	// Fabric is the network fabric type (efa, ib).
	Fabric CriteriaFabricType `json:"fabric,omitempty" yaml:"fabric,omitempty"`

	// Accelerator is the GPU/accelerator type (h100, gb200, a100, l40).
	Accelerator CriteriaAcceleratorType `json:"accelerator,omitempty" yaml:"accelerator,omitempty"`

	// Intent is the workload intent (training, inference).
	Intent CriteriaIntentType `json:"intent,omitempty" yaml:"intent,omitempty"`

	// Worker is the worker node OS type.
	Worker CriteriaOSType `json:"worker,omitempty" yaml:"worker,omitempty"`

	// System is the system/control-plane node OS type.
	System CriteriaOSType `json:"system,omitempty" yaml:"system,omitempty"`

	// Nodes is the number of worker nodes (0 means any/unspecified).
	Nodes int `json:"nodes,omitempty" yaml:"nodes,omitempty"`
}

// NewCriteria creates a new Criteria with all fields set to "any".
func NewCriteria() *Criteria {
	return &Criteria{
		Service:     CriteriaServiceAny,
		Fabric:      CriteriaFabricAny,
		Accelerator: CriteriaAcceleratorAny,
		Intent:      CriteriaIntentAny,
		Worker:      CriteriaOSAny,
		System:      CriteriaOSAny,
		Nodes:       0,
	}
}

// Matches checks if the given criteria matches this criteria.
// A criteria matches if all non-"any" fields match.
// "any" acts as a wildcard and matches everything.
func (c *Criteria) Matches(other *Criteria) bool {
	if other == nil {
		return true
	}

	// Check each field - "any" matches everything
	if c.Service != CriteriaServiceAny && other.Service != CriteriaServiceAny && c.Service != other.Service {
		return false
	}
	if c.Fabric != CriteriaFabricAny && other.Fabric != CriteriaFabricAny && c.Fabric != other.Fabric {
		return false
	}
	if c.Accelerator != CriteriaAcceleratorAny && other.Accelerator != CriteriaAcceleratorAny && c.Accelerator != other.Accelerator {
		return false
	}
	if c.Intent != CriteriaIntentAny && other.Intent != CriteriaIntentAny && c.Intent != other.Intent {
		return false
	}
	if c.Worker != CriteriaOSAny && other.Worker != CriteriaOSAny && c.Worker != other.Worker {
		return false
	}
	if c.System != CriteriaOSAny && other.System != CriteriaOSAny && c.System != other.System {
		return false
	}
	// Nodes: 0 means any, otherwise must match exactly
	if c.Nodes != 0 && other.Nodes != 0 && c.Nodes != other.Nodes {
		return false
	}

	return true
}

// Specificity returns a score indicating how specific this criteria is.
// Higher scores mean more specific criteria (fewer "any" fields).
// Used for ordering overlay application - more specific overlays are applied later.
func (c *Criteria) Specificity() int {
	score := 0
	if c.Service != CriteriaServiceAny {
		score++
	}
	if c.Fabric != CriteriaFabricAny {
		score++
	}
	if c.Accelerator != CriteriaAcceleratorAny {
		score++
	}
	if c.Intent != CriteriaIntentAny {
		score++
	}
	if c.Worker != CriteriaOSAny {
		score++
	}
	if c.System != CriteriaOSAny {
		score++
	}
	if c.Nodes != 0 {
		score++
	}
	return score
}

// String returns a human-readable representation of the criteria.
func (c *Criteria) String() string {
	parts := []string{}
	if c.Service != CriteriaServiceAny {
		parts = append(parts, fmt.Sprintf("service=%s", c.Service))
	}
	if c.Fabric != CriteriaFabricAny {
		parts = append(parts, fmt.Sprintf("fabric=%s", c.Fabric))
	}
	if c.Accelerator != CriteriaAcceleratorAny {
		parts = append(parts, fmt.Sprintf("accelerator=%s", c.Accelerator))
	}
	if c.Intent != CriteriaIntentAny {
		parts = append(parts, fmt.Sprintf("intent=%s", c.Intent))
	}
	if c.Worker != CriteriaOSAny {
		parts = append(parts, fmt.Sprintf("worker=%s", c.Worker))
	}
	if c.System != CriteriaOSAny {
		parts = append(parts, fmt.Sprintf("system=%s", c.System))
	}
	if c.Nodes != 0 {
		parts = append(parts, fmt.Sprintf("nodes=%d", c.Nodes))
	}
	if len(parts) == 0 {
		return "criteria(any)"
	}
	return fmt.Sprintf("criteria(%s)", strings.Join(parts, ", "))
}

// CriteriaOption is a functional option for building Criteria.
type CriteriaOption func(*Criteria) error

// WithCriteriaService sets the service type.
func WithCriteriaService(s string) CriteriaOption {
	return func(c *Criteria) error {
		st, err := ParseCriteriaServiceType(s)
		if err != nil {
			return err
		}
		c.Service = st
		return nil
	}
}

// WithCriteriaFabric sets the fabric type.
func WithCriteriaFabric(s string) CriteriaOption {
	return func(c *Criteria) error {
		ft, err := ParseCriteriaFabricType(s)
		if err != nil {
			return err
		}
		c.Fabric = ft
		return nil
	}
}

// WithCriteriaAccelerator sets the accelerator type.
func WithCriteriaAccelerator(s string) CriteriaOption {
	return func(c *Criteria) error {
		at, err := ParseCriteriaAcceleratorType(s)
		if err != nil {
			return err
		}
		c.Accelerator = at
		return nil
	}
}

// WithCriteriaIntent sets the intent type.
func WithCriteriaIntent(s string) CriteriaOption {
	return func(c *Criteria) error {
		it, err := ParseCriteriaIntentType(s)
		if err != nil {
			return err
		}
		c.Intent = it
		return nil
	}
}

// WithCriteriaWorker sets the worker OS type.
func WithCriteriaWorker(s string) CriteriaOption {
	return func(c *Criteria) error {
		ot, err := ParseCriteriaOSType(s)
		if err != nil {
			return err
		}
		c.Worker = ot
		return nil
	}
}

// WithCriteriaSystem sets the system OS type.
func WithCriteriaSystem(s string) CriteriaOption {
	return func(c *Criteria) error {
		ot, err := ParseCriteriaOSType(s)
		if err != nil {
			return err
		}
		c.System = ot
		return nil
	}
}

// WithCriteriaNodes sets the number of nodes.
func WithCriteriaNodes(n int) CriteriaOption {
	return func(c *Criteria) error {
		if n < 0 {
			return fmt.Errorf("invalid nodes count: %d (must be >= 0)", n)
		}
		c.Nodes = n
		return nil
	}
}

// BuildCriteria creates a Criteria from functional options.
func BuildCriteria(opts ...CriteriaOption) (*Criteria, error) {
	c := NewCriteria()
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// ParseCriteriaFromRequest parses recipe criteria from HTTP query parameters.
// All parameters are optional and default to "any" if not specified.
// Supported parameters: service, fabric, accelerator (alias: gpu), intent, worker, system, nodes.
func ParseCriteriaFromRequest(r *http.Request) (*Criteria, error) {
	if r == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	q := r.URL.Query()
	return ParseCriteriaFromValues(q)
}

// ParseCriteriaFromValues parses recipe criteria from URL values.
// All parameters are optional and default to "any" if not specified.
// Supported parameters: service, fabric, accelerator (alias: gpu), intent, worker, system, nodes.
func ParseCriteriaFromValues(values url.Values) (*Criteria, error) {
	c := NewCriteria()

	// Parse service
	if s := values.Get("service"); s != "" {
		st, err := ParseCriteriaServiceType(s)
		if err != nil {
			return nil, err
		}
		c.Service = st
	}

	// Parse fabric
	if s := values.Get("fabric"); s != "" {
		ft, err := ParseCriteriaFabricType(s)
		if err != nil {
			return nil, err
		}
		c.Fabric = ft
	}

	// Parse accelerator (also accept "gpu" as alias for backwards compatibility)
	accelParam := values.Get("accelerator")
	if accelParam == "" {
		accelParam = values.Get("gpu")
	}
	if accelParam != "" {
		at, err := ParseCriteriaAcceleratorType(accelParam)
		if err != nil {
			return nil, err
		}
		c.Accelerator = at
	}

	// Parse intent
	if s := values.Get("intent"); s != "" {
		it, err := ParseCriteriaIntentType(s)
		if err != nil {
			return nil, err
		}
		c.Intent = it
	}

	// Parse worker OS
	if s := values.Get("worker"); s != "" {
		ot, err := ParseCriteriaOSType(s)
		if err != nil {
			return nil, err
		}
		c.Worker = ot
	}

	// Parse system OS
	if s := values.Get("system"); s != "" {
		ot, err := ParseCriteriaOSType(s)
		if err != nil {
			return nil, err
		}
		c.System = ot
	}

	// Parse nodes count
	if s := values.Get("nodes"); s != "" {
		var n int
		if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
			return nil, fmt.Errorf("invalid nodes value: %s", s)
		}
		if n < 0 {
			return nil, fmt.Errorf("invalid nodes count: %d (must be >= 0)", n)
		}
		c.Nodes = n
	}

	return c, nil
}

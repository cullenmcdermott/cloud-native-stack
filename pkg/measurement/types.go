package measurement

type Type string

func (mt Type) String() string {
	return string(mt)
}

const (
	TypeGrub    Type = "Grub"
	TypeImage   Type = "Image"
	TypeKMod    Type = "KMod"
	TypeK8s     Type = "K8s"
	TypeSMI     Type = "SMI"
	TypeSysctl  Type = "Sysctl"
	TypeSystemD Type = "SystemD"
)

// Types is the list of all supported measurement types.
var Types = []Type{
	TypeGrub,
	TypeImage,
	TypeKMod,
	TypeK8s,
	TypeSMI,
	TypeSysctl,
	TypeSystemD,
}

// ParseMeasurementType parses a string into a Type.
func ParseType(s string) (Type, bool) {
	for _, mt := range Types {
		if string(mt) == s {
			return mt, true
		}
	}
	return "", false
}

// Measurement represents a single collector configuration measurement.
type Measurement struct {
	Type Type `json:"type" yaml:"type"`
	Data any  `json:"data" yaml:"data"`
}

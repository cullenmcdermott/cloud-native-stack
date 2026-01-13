package result

import "time"

// Artifacts represents the result of deployer artifact generation.
type Artifacts struct {
	// Type is the deployer type that generated these artifacts
	Type string

	// Files is the list of file paths generated
	Files []string

	// ReadmeContent is the generated README content
	ReadmeContent string

	// Duration is how long artifact generation took
	Duration time.Duration

	// Success indicates if generation succeeded
	Success bool

	// Error contains any error message if Success is false
	Error string
}

// New creates a new Artifacts instance for the given deployer type.
func New(deployerType string) *Artifacts {
	return &Artifacts{
		Type:    deployerType,
		Files:   []string{},
		Success: true,
	}
}

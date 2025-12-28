package bundler

// MakeOptions configures how bundles are generated.
type MakeOptions struct {
	// BundlerTypes specifies which bundlers to execute.
	// If empty, all registered bundlers are executed.
	BundlerTypes []BundleType

	// Parallel enables concurrent execution of bundlers.
	// Default is false (sequential execution).
	Parallel bool

	// FailFast stops execution on first bundler error.
	// Default is false (continues and collects all errors).
	FailFast bool

	// Config provides bundler-specific configuration.
	Config *BundlerConfig

	// DryRun simulates bundle generation without writing files.
	DryRun bool
}

// MakeOption is a functional option for configuring MakeOptions.
type MakeOption func(*MakeOptions)

// WithBundlers specifies which bundlers to execute.
// If not set, all registered bundlers are executed.
func WithBundlers(types ...BundleType) MakeOption {
	return func(opts *MakeOptions) {
		opts.BundlerTypes = types
	}
}

// WithParallel enables concurrent execution of bundlers.
// This can significantly speed up bundle generation when multiple bundlers are used.
func WithParallel() MakeOption {
	return func(opts *MakeOptions) {
		opts.Parallel = true
	}
}

// WithFailFast stops execution on the first bundler error.
// Useful when you want to abort early rather than collecting all errors.
func WithFailFast() MakeOption {
	return func(opts *MakeOptions) {
		opts.FailFast = true
	}
}

// WithConfig provides bundler-specific configuration.
func WithConfig(config *BundlerConfig) MakeOption {
	return func(opts *MakeOptions) {
		opts.Config = config
	}
}

// WithDryRun simulates bundle generation without writing files.
// Useful for validation and testing.
func WithDryRun() MakeOption {
	return func(opts *MakeOptions) {
		opts.DryRun = true
	}
}

// applyDefaults applies default values to options.
func (opts *MakeOptions) applyDefaults() {
	if opts.Config == nil {
		opts.Config = DefaultBundlerConfig()
	}
}

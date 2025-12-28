// Package bundler provides functionality to create application bundles
// based on predefined recipes. It supports reading recipe files, assembling
// application components, and outputting the final bundle to a specified
// directory.
//
// # Architecture
//
// The bundler package uses a registry-based architecture where bundlers
// can be dynamically registered and executed. Each bundler is responsible
// for generating deployment artifacts for a specific component (e.g., GPU Operator).
//
// # Bundler Interface
//
// Bundlers implement the Bundler interface:
//
//	type Bundler interface {
//	    Make(ctx context.Context, recipe *recipe.Recipe, dir string) (*BundleResult, error)
//	}
//
// Optional interfaces:
//
//	ConfigurableBundler - Supports configuration
//	Validator - Validates recipes before bundling
//
// # Usage
//
// Basic usage:
//
//	recipe, _ := recipe.NewBuilder().BuildFromQuery(ctx, query)
//	output, err := bundler.Make(ctx, recipe, "./output")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(output.Summary())
//
// With options:
//
//	output, err := bundler.Make(ctx, recipe, "./output",
//	    bundler.WithBundlers(bundler.BundleTypeGpuOperator),
//	    bundler.WithParallel(),
//	    bundler.WithConfig(config),
//	)
//
// # Adding New Bundlers
//
// To add a new bundler:
//
// 1. Create a package under pkg/bundler/yourcomponent/
// 2. Implement the Bundler interface
// 3. Register in init():
//
//	func init() {
//	    bundler.Register(bundler.BundleTypeYourComponent, NewBundler())
//	}
//
// # Bundle Structure
//
// Generated bundles follow this structure:
//
//	output_dir/
//	  component-name/
//	    values.yaml          # Configuration values
//	    manifests/           # Kubernetes manifests
//	      namespace.yaml
//	      clusterpolicy.yaml
//	    scripts/             # Installation scripts
//	      install.sh
//	      uninstall.sh
//	    README.md            # Documentation
//	    checksums.txt        # File checksums
//	  metadata.json          # Bundle metadata
//
// # Error Handling
//
// Bundlers use structured errors from pkg/errors for consistent error handling.
// Errors include error codes, context, and are suitable for programmatic handling.
//
// # Observability
//
// The bundler package exposes Prometheus metrics:
//
//	eidos_bundles_generated_total      - Total bundles generated
//	eidos_bundle_duration_seconds      - Bundle generation duration
//	eidos_bundle_size_bytes            - Bundle size
//	eidos_bundle_files_total           - Files per bundle
//	eidos_bundle_errors_total          - Generation errors
//
// # Validation
//
// Recipes are validated before bundling:
//
//	if err := bundler.ValidateRecipe(ctx, recipe); err != nil {
//	    // Handle validation error
//	}
//
// # Configuration
//
// Bundlers can be configured using BundlerConfig:
//
//	config := bundler.DefaultBundlerConfig()
//	config.OutputFormat = "yaml"
//	config.IncludeScripts = true
//	config.HelmChartVersion = "v1.0.0"
package bundler

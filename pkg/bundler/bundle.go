package bundler

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/NVIDIA/cloud-native-stack/pkg/errors"
	"github.com/NVIDIA/cloud-native-stack/pkg/recipe"
	"golang.org/x/sync/errgroup"
)

// Make generates bundles from the given recipe into the specified directory.
// It accepts various options to customize the bundling process.
// Returns a BundleOutput summarizing the results of the bundling operation.
// Errors encountered during the process are returned as well.
func Make(ctx context.Context, recipe *recipe.Recipe, dir string, opts ...MakeOption) (*BundleOutput, error) {
	start := time.Now()

	// Apply options
	options := &MakeOptions{}
	for _, opt := range opts {
		opt(options)
	}
	options.applyDefaults()

	// Validate input
	if recipe == nil {
		return nil, errors.New(errors.ErrCodeInvalidRequest, "recipe cannot be nil")
	}

	if err := ValidateRecipeStructure(recipe); err != nil {
		return nil, errors.Wrap(errors.ErrCodeInvalidRequest, "recipe validation failed", err)
	}

	if dir == "" {
		dir = "."
	}

	// Create output directory
	if !options.DryRun {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, errors.Wrap(errors.ErrCodeInternal,
				fmt.Sprintf("failed to create directory %s", dir), err)
		}
	}

	// Select bundlers to execute
	bundlers := selectBundlers(options.BundlerTypes)
	if len(bundlers) == 0 {
		return nil, errors.New(errors.ErrCodeInvalidRequest, "no bundlers selected")
	}

	slog.Info("starting bundle generation",
		"bundler_count", len(bundlers),
		"output_dir", dir,
		"parallel", options.Parallel,
		"dry_run", options.DryRun,
	)

	// Generate bundles
	var output *BundleOutput
	var err error

	if options.Parallel {
		output, err = makeParallel(ctx, recipe, dir, bundlers, options)
	} else {
		output, err = makeSequential(ctx, recipe, dir, bundlers, options)
	}

	if err != nil {
		return output, err
	}

	output.TotalDuration = time.Since(start)
	output.OutputDir = dir

	slog.Info("bundle generation complete", "summary", output.Summary())

	return output, nil
}

// makeSequential executes bundlers sequentially.
func makeSequential(ctx context.Context, recipe *recipe.Recipe, dir string,
	bundlers map[BundleType]Bundler, options *MakeOptions) (*BundleOutput, error) {

	output := &BundleOutput{
		Results: make([]*BundleResult, 0, len(bundlers)),
		Errors:  make([]BundleError, 0),
	}

	for bundlerType, b := range bundlers {
		result, err := executeBundler(ctx, bundlerType, b, recipe, dir, options)
		output.Results = append(output.Results, result)

		if result.Success {
			output.TotalSize += result.Size
			output.TotalFiles += len(result.Files)
		}

		if err != nil {
			bundleErr := BundleError{
				BundlerType: bundlerType,
				Error:       err.Error(),
			}
			output.Errors = append(output.Errors, bundleErr)

			slog.Error("bundler failed",
				"bundler_type", bundlerType,
				"error", err,
			)

			if options.FailFast {
				return output, errors.Wrap(errors.ErrCodeInternal,
					fmt.Sprintf("bundler %s failed", bundlerType), err)
			}
		}
	}

	if len(output.Errors) > 0 && options.FailFast {
		return output, errors.New(errors.ErrCodeInternal,
			fmt.Sprintf("%d bundler(s) failed", len(output.Errors)))
	}

	return output, nil
}

// makeParallel executes bundlers concurrently.
func makeParallel(ctx context.Context, recipe *recipe.Recipe, dir string,
	bundlers map[BundleType]Bundler, options *MakeOptions) (*BundleOutput, error) {

	output := &BundleOutput{
		Results: make([]*BundleResult, 0, len(bundlers)),
		Errors:  make([]BundleError, 0),
	}

	g, gctx := errgroup.WithContext(ctx)
	resultChan := make(chan *BundleResult, len(bundlers))
	errorChan := make(chan BundleError, len(bundlers))

	for bundlerType, b := range bundlers {
		bundlerType := bundlerType // capture loop variable
		b := b

		g.Go(func() error {
			result, err := executeBundler(gctx, bundlerType, b, recipe, dir, options)
			resultChan <- result

			if err != nil {
				errorChan <- BundleError{
					BundlerType: bundlerType,
					Error:       err.Error(),
				}

				if options.FailFast {
					return err
				}
			}
			return nil
		})
	}

	// Wait for all bundlers
	err := g.Wait()
	close(resultChan)
	close(errorChan)

	// Collect results
	for result := range resultChan {
		output.Results = append(output.Results, result)
		if result.Success {
			output.TotalSize += result.Size
			output.TotalFiles += len(result.Files)
		}
	}

	// Collect errors
	for bundleErr := range errorChan {
		output.Errors = append(output.Errors, bundleErr)
	}

	if err != nil && options.FailFast {
		return output, errors.Wrap(errors.ErrCodeInternal, "bundler execution failed", err)
	}

	return output, nil
}

// executeBundler executes a single bundler and records metrics.
func executeBundler(ctx context.Context, bundlerType BundleType, b Bundler,
	recipe *recipe.Recipe, dir string, options *MakeOptions) (*BundleResult, error) {

	start := time.Now()
	result := NewBundleResult(bundlerType)

	slog.Debug("executing bundler",
		"bundler_type", bundlerType,
		"output_dir", dir,
	)

	// Configure bundler if it supports configuration
	if cb, ok := b.(ConfigurableBundler); ok && options.Config != nil {
		if err := cb.Configure(options.Config); err != nil {
			recordBundleError(bundlerType, "configuration")
			return result, err
		}
	}

	// Validate if bundler supports validation
	if v, ok := b.(Validator); ok {
		if err := v.Validate(ctx, recipe); err != nil {
			recordValidationFailure(bundlerType)
			recordBundleError(bundlerType, "validation")
			return result, err
		}
	}

	// Execute bundler
	if !options.DryRun {
		bundlerResult, err := b.Make(ctx, recipe, dir)
		if err != nil {
			result.Duration = time.Since(start)
			recordBundleGenerated(bundlerType, false)
			recordBundleDuration(bundlerType, result.Duration.Seconds())
			recordBundleError(bundlerType, "execution")
			return result, err
		}
		result = bundlerResult
	}

	result.Duration = time.Since(start)
	result.MarkSuccess()

	// Record metrics
	recordBundleGenerated(bundlerType, true)
	recordBundleDuration(bundlerType, result.Duration.Seconds())
	recordBundleSize(bundlerType, result.Size)
	recordBundleFiles(bundlerType, len(result.Files))

	slog.Info("bundler completed",
		"bundler_type", bundlerType,
		"files", len(result.Files),
		"size_bytes", result.Size,
		"duration", result.Duration.Round(time.Millisecond),
	)

	return result, nil
}

// selectBundlers selects which bundlers to execute based on options.
func selectBundlers(types []BundleType) map[BundleType]Bundler {
	if len(types) == 0 {
		// Return all registered bundlers
		return defaultRegistry.GetAll()
	}

	// Return only specified bundlers
	selected := make(map[BundleType]Bundler)
	for _, t := range types {
		if b, ok := defaultRegistry.Get(t); ok {
			selected[t] = b
		}
	}
	return selected
}

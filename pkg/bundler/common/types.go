package common

import (
	"context"

	"github.com/NVIDIA/cloud-native-stack/pkg/bundler/result"
	"github.com/NVIDIA/cloud-native-stack/pkg/recipe"
)

// Bundler defines the interface for creating application bundles.
// Implementations generate deployment artifacts from recipes.
type Bundler interface {
	Make(ctx context.Context, recipe *recipe.Recipe, dir string) (*result.Result, error)
}

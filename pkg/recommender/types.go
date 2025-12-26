package recommender

import (
	"context"

	"github.com/NVIDIA/cloud-native-stack/pkg/recipe"
	"github.com/NVIDIA/cloud-native-stack/pkg/snapshotter"
)

// Recommender defines the interface for generating recommendations based on snapshots and intent.
type Recommender interface {
	Recommend(ctx context.Context, intent recipe.IntentType, snap *snapshotter.Snapshot) (*recipe.Recipe, error)
}

// Recommendation is the structure representing recommended configuration for a given set of
// environment configurations and intent. The environment configurations are derived from the
// provided snapshot.
type Recommendation struct {
	*recipe.Recipe
}

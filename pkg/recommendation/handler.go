package recommendation

import (
	"fmt"
	"net/http"

	"github.com/NVIDIA/cloud-native-stack/pkg/serializers"
	"github.com/NVIDIA/cloud-native-stack/pkg/server"
)

var (
	recommendCacheTTLInSec = 600 * 1000 // 10 minutes
)

// HandleRecommendations processes recommendation requests and returns recommendations.
// It supports GET requests with query parameters to specify recommendation criteria.
// The response is returned in JSON format with appropriate caching headers.
// Errors are handled and returned in a structured format.
func (b *Builder) HandleRecommendations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		server.WriteError(w, r, http.StatusMethodNotAllowed, server.ErrCodeMethodNotAllowed,
			"Method not allowed", false, map[string]interface{}{
				"method": r.Method,
			})
		return
	}

	q, err := ParseQuery(r)
	if err != nil {
		server.WriteError(w, r, http.StatusBadRequest, server.ErrCodeInvalidRequest,
			"Invalid recommendation query", false, map[string]interface{}{
				"error": err.Error(),
			})
		return
	}

	resp, err := b.Build(q)
	if err != nil {
		server.WriteError(w, r, http.StatusInternalServerError, server.ErrCodeInternalError,
			"Failed to build recommendation", true, map[string]interface{}{
				"error": err.Error(),
			})
		return
	}

	// Set caching headers
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", recommendCacheTTLInSec))

	serializers.RespondJSON(w, http.StatusOK, resp)
}

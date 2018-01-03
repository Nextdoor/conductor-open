package core

import (
	"fmt"
	"net/http"

	"github.com/Nextdoor/conductor/services/data"
	"github.com/Nextdoor/conductor/shared/types"
)

func searchEndpoints() []endpoint {
	return []endpoint{
		newEp("/api/search", get, search),
	}
}

func search(r *http.Request) response {
	dataClient := data.NewClient()

	query := r.URL.Query()
	params := make(map[string]string)
	for key, values := range query {
		params[key] = values[0]
	}
	sha, ok := params["commit"]
	if !ok {
		return errorResponse(
			"Search only supports commit",
			http.StatusBadRequest)
	}
	commit := &types.Commit{SHA: sha}

	trains, err := dataClient.TrainsByCommit(commit)
	if err != nil {
		return errorResponse(
			fmt.Sprintf("Error getting trains for commit: %v", err),
			http.StatusInternalServerError)
	}
	if len(trains) == 0 {
		return errorResponse(
			fmt.Sprintf("Could not find any trains for commit %s", sha),
			http.StatusNotFound)
	}

	return dataResponse(&types.Search{
		Params:  params,
		Results: trains,
	})
}

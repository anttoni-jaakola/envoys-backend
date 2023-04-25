package future

import (
	"context"
	// "fmt"
	// "time"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbfuture"
)

func (a *Service) GetFutures(_ context.Context, req *pbfuture.GetRequestFutures) (*pbfuture.ResponseFutures, error) {
	var (
		response pbfuture.ResponseFutures
		// exist    bool
	)
	return &response, nil
}

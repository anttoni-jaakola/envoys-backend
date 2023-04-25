package future

import (
	"context"
	// "fmt"
	// "time"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbfuture"
)

func (a *Service) GetFutures(_ context.Context, req *pbfuture.GetRequest) *pbfuture.Response {
	var (
		response pbprovider.Response
		exist    bool
	)
	return &response, nil
}

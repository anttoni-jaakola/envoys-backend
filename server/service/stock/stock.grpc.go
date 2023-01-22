package stock

import (
	"github.com/cryptogateway/backend-envoys/server/proto/pbstock"
	"golang.org/x/net/context"
)

func (e *Service) GetSymbol(ctx context.Context, symbol *pbstock.GetRequestSymbol) (*pbstock.ResponseSymbol, error) {

	var (
		response pbstock.ResponseSymbol
	)

	response.Success = true

	return &response, nil
}

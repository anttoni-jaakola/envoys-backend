package ohlcv

import (
	"github.com/cryptogateway/backend-envoys/assets"
)

// Service - The purpose of this code is to create a "Service" struct that contains a pointer to an assets.Context. This allows the
// service to access the context and any of the assets within the context.
type Service struct {
	Context *assets.Context
}

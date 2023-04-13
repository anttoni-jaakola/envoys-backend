package admin_spot

import (
	"github.com/cryptogateway/backend-envoys/assets"
)

// Service - The purpose of the Service struct is to store data related to a service, such as the Context, run and wait maps, and
// the block map. The Context is a pointer to an assets Context, which contains information about the service. The run
// and wait maps are booleans that indicate whether the service is running or waiting for an action. The block map is an
// integer that stores the block number associated with a particular service.
type Service struct {
	Context *assets.Context
}

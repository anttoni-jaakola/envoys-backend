package kyc

import "github.com/cryptogateway/backend-envoys/assets"

// Service - The Service struct is used to create a structure that holds a pointer to an assets.Context. This allows the Service
// struct to access the assets.Context and all of its data, and to use that data to carry out tasks.
type Service struct {
	Context *assets.Context
}

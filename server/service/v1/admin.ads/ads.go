package admin_ads

import (
	"github.com/cryptogateway/backend-envoys/assets"
)

// Service - The type Service struct is used to store a pointer to an assets.Context object. This type is used to provide access to
// application-level assets such as templates, images, and configuration files that are used throughout the application.
type Service struct {
	Context *assets.Context
}

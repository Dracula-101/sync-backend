package moderatordto

// PermissionCheckResponse is the response for checking a permission
type PermissionCheckResponse struct {
	HasPermission bool `json:"hasPermission"`
}

// NewPermissionCheckResponse creates a new response for checking a permission
func NewPermissionCheckResponse(hasPermission bool) *PermissionCheckResponse {
	return &PermissionCheckResponse{
		HasPermission: hasPermission,
	}
}

package models

type ListVisibleWorkspacesResponse struct {
	WorkspaceID string `json:"workspaceId"`
	Alias       string `json:"alias"`
}

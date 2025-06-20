package services

import (
	"context"

	ghippov1alpha1 "ghippo.io/api/wssdk/v1alpha1/types"
	wsypes "ghippo.io/api/wssdk/v1alpha1/types"
	"github.com/auth-engine/config"
	"github.com/auth-engine/pkg/clients"
	apiv1alpha1 "kpanda.io/api/apiextensions/v1alpha1"
	"kpanda.io/api/types"
)

type IClientService interface {
	ListCR(ctx context.Context, cluster, namespace, name, group, resource, version, workspace string, page *types.Pagination) (*apiv1alpha1.ListCustomResourcesResponse, error)
	ListVisibleWorkspaces(ctx context.Context) ([]ghippov1alpha1.Workspace, error)
}
type ClientService struct {
	DceClient *clients.DCEClient
	GhippoSdk wsypes.Interface
}

func NewClientService(appConfig *config.AppConfig, ghippoSdk wsypes.Interface) IClientService {
	return &ClientService{
		DceClient: clients.NewDCEClient(appConfig.DCE.URL, appConfig.InsecureSkipVerify),
		GhippoSdk: ghippoSdk,
	}
}

func (s *ClientService) ListCR(ctx context.Context, cluster, namespace, name, group, resource, version, workspace string, page *types.Pagination) (*apiv1alpha1.ListCustomResourcesResponse, error) {
	return s.DceClient.ListCR(ctx, cluster, namespace, name, group, resource, version, workspace, page)
}

func (s *ClientService) ListVisibleWorkspaces(ctx context.Context) ([]ghippov1alpha1.Workspace, error) {
	return s.GhippoSdk.ListVisibleWorkspaces(ctx)
}

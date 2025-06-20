package clients

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/go-resty/resty/v2"
	"k8s.io/apimachinery/pkg/labels"
	apiv1alpha1 "kpanda.io/api/apiextensions/v1alpha1"
	ghippov1alpha1 "kpanda.io/api/ghippo/v1alpha1"
	"kpanda.io/api/types"
)

type DCEClient struct {
	client *resty.Client
}

func NewDCEClient(dceURL string, skipVerify bool) *DCEClient {
	c := resty.New().
		SetRetryCount(1).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(3 * time.Second).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: skipVerify}).
		SetBaseURL(dceURL)
	return &DCEClient{client: c}
}

func getDceToken(ctx context.Context) string {
	return ctx.Value("DceToken").(string)
}

func (c *DCEClient) ListCluster(ctx context.Context) ([]byte, error) {
	res, err := c.client.R().
		SetAuthToken(getDceToken(ctx)).
		Get("/apis/kpanda.io/v1alpha1/clusters")
	if !res.IsSuccess() {
		hlog.Errorf("list cluster failed: %s, err: %v, code: %d", string(res.Body()), err, res.StatusCode())
		return nil, fmt.Errorf("list cluster failed: %s, err: %v, code: %d", string(res.Body()), err, res.StatusCode())
	}
	return res.Body(), nil
}

func (c *DCEClient) ListCR(ctx context.Context, cluster, namespace, name, group, resource, version, workspace string, page *types.Pagination) (*apiv1alpha1.ListCustomResourcesResponse, error) {
	response := &apiv1alpha1.ListCustomResourcesResponse{}

	// Prepare the path parameters
	pathParams := map[string]string{
		"cluster":   cluster,
		"group":     group,
		"version":   version,
		"namespace": namespace,
		"resource":  resource,
	}

	// Initialize the query parameters map
	queryParams := map[string]string{
		"showDetail": "true",
		"page":       fmt.Sprintf("%d", page.Page),
		"pageSize":   fmt.Sprintf("%d", page.PageSize),
	}
	// Conditionally add the labelSelector query parameter
	if workspace != "" && resource != "llms" {
		label := labels.FormatLabels(map[string]string{
			"workspace": workspace,
		})
		queryParams["labelSelector"] = label
	}

	if name != "" {
		queryParams["name"] = name
	}

	// Make the request with conditional query parameters
	resp, err := c.client.R().
		SetAuthToken(getDceToken(ctx)).
		SetPathParams(pathParams).
		SetQueryParams(queryParams).
		SetResult(response).
		SetHeader("Accept", "application/json").
		Get("/apis/kpanda.io/v1alpha1/clusters/{cluster}/gvr/{group}/{version}/namespaces/{namespace}/{resource}")

	// Check for errors in the response
	if err != nil || !resp.IsSuccess() {
		hlog.Errorf("list resource %s failed, err: %v, response: %s", resource, err, string(resp.Body()))
		return nil, fmt.Errorf("list resource %s failed, err: %v, response: %s", resource, err, string(resp.Body()))
	}
	return response, nil
}

func (c *DCEClient) ListWorkspace(ctx context.Context) (*ghippov1alpha1.ListWorkspacesResponse, error) {
	response := &ghippov1alpha1.ListWorkspacesResponse{}
	res, err := c.client.R().
		SetResult(&response).
		SetAuthToken(getDceToken(ctx)).
		Get("/apis/kpanda.io/v1alpha1/workspaces")
	if err != nil || !res.IsSuccess() {
		hlog.Errorf("list workspace %s  failed, err: %v", err, string(res.Body()))
		return nil, fmt.Errorf("list workspace failed, err: %v, response: %s", err, string(res.Body()))
	}
	return response, nil
}

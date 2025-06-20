package clients

import (
	"context"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"insight.io/api/insight/v1alpha1"
)

//counterfeiter:generate -o fake/insight_insight_client.go --fake-name FakeInsightInsightClient insight.io/api/insight/v1alpha1.InsightClient
//counterfeiter:generate -o fake/insight_metric_client.go --fake-name FakeInsightMetricClient insight.io/api/insight/v1alpha1.MetricClient

type InsightClients struct {
	Client       v1alpha1.InsightClient
	MetricClient v1alpha1.MetricClient
}

func NewInsightClient(addr string) (*InsightClients, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	client := v1alpha1.NewInsightClient(conn)
	_, _ = client.GetVersion(context.TODO(), &v1alpha1.Empty{})

	metricClient := v1alpha1.NewMetricClient(conn)

	return &InsightClients{
		Client:       client,
		MetricClient: metricClient,
	}, nil
}

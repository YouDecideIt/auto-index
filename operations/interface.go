package operations

import "context"

type Cluster struct {
	SQLEndpoint string
	NgmEndpoint string
}

type BCluster interface {
	StartBCluster(ctx context.Context) error
	WaitBClusterStartedAndMirrored(ctx context.Context) (*Cluster, error)
	DestroyBCluster(ctx context.Context) error
}

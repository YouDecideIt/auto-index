package operations

import "context"

type BClusterEndpoint struct {
	SQLEndpoint string
	NgmEndpoint string
}

type BCluster interface {
	StartBCluster(ctx context.Context) error
	WaitBClusterStartedAndMirrored(ctx context.Context) (*BClusterEndpoint, error)
	DestroyBCluster(ctx context.Context) error
}

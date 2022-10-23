package b_cluster

import "github.com/YouDecideIt/auto-index/context"

var _ = BCluster(MockCluster{})

type BClusterEndpoint struct {
	SQLEndpoint string
	NgmEndpoint string
}

type BCluster interface {
	StartBCluster()
	WaitBClusterStartedAndMirrored(ctx context.Context) (*BClusterEndpoint, error)
	DestroyBCluster()
}

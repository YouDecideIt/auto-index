package b_cluster

import "github.com/YouDecideIt/auto-index/context"

type MockCluster struct {
}

func (m MockCluster) StartBCluster() {

}

func (m MockCluster) WaitBClusterStartedAndMirrored(_ context.Context) (*BClusterEndpoint, error) {
	return &BClusterEndpoint{NgmEndpoint: "172.16.4.42:23309", SQLEndpoint: "172.16.4.42:23300"}, nil

}

func (m MockCluster) DestroyBCluster() {

}

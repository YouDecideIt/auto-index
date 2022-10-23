package operations

import "github.com/YouDecideIt/auto-index/context"

type MockCluster struct {
}

func (m MockCluster) StartBCluster() {

}

func (m MockCluster) WaitBClusterStartedAndMirrored(_ context.Context) (*BClusterEndpoint, error) {
	return &BClusterEndpoint{NgmEndpoint: "172.16.4.42:23709", SQLEndpoint: "172.16.4.42:23700"}, nil

}

func (m MockCluster) DestroyBCluster() {

}

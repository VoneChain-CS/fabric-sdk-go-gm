/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	selectopts "github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/client/common/selection/options"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/options"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/providers/context"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/providers/fab"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/fab/mocks"
)

// MockSelectionService implements mock selection service
type MockSelectionService struct {
	Error          error
	Peers          []fab.Peer
	ChannelContext context.Channel
}

// NewMockSelectionService returns mock selection service
func NewMockSelectionService(err error, peers ...fab.Peer) *MockSelectionService {
	return &MockSelectionService{Error: err, Peers: peers}
}

// GetEndorsersForChaincode mockcore retrieving endorsing peers
func (ds *MockSelectionService) GetEndorsersForChaincode(chaincodes []*fab.ChaincodeCall, opts ...options.Opt) ([]fab.Peer, error) {

	if ds.Error != nil {
		return nil, ds.Error
	}

	params := selectopts.NewParams(opts)

	var peers []fab.Peer
	if ds.ChannelContext != nil {
		var err error
		discovery, err := ds.ChannelContext.ChannelService().Discovery()
		if err != nil {
			return nil, err
		}
		peers, err = discovery.GetPeers()
		if err != nil {
			return nil, err
		}
	} else if ds.Peers == nil {
		mockPeer := mocks.NewMockPeer("Peer1", "http://peer1.com")
		peers = append(peers, mockPeer)
	}

	if params.PeerFilter != nil {
		for _, p := range ds.Peers {
			if params.PeerFilter(p) {
				peers = append(peers, p)
			}
		}
	} else {
		peers = ds.Peers
	}

	if params.PeerSorter != nil {
		sortedPeers := make([]fab.Peer, len(peers))
		copy(sortedPeers, peers)
		peers = params.PeerSorter(sortedPeers)
	}

	return peers, nil

}

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peerresolver

import (
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/options"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/providers/context"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/providers/fab"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/fab/events/service"
)

// Resolver decided which peer to connect to and when to disconnect from that peer
type Resolver interface {
	// Resolve chooses a peer from the given set of peers
	Resolve(peers []fab.Peer) (fab.Peer, error)
	// ShouldDisconnect returns true to disconnect from the connected peer
	ShouldDisconnect(peers []fab.Peer, connectedPeer fab.Peer) bool
}

// Provider creates a peer Resolver
type Provider func(ed service.Dispatcher, context context.Client, channelID string, opts ...options.Opt) Resolver

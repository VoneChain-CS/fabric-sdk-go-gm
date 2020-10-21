/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defsvc

import (
	discovery "github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/client/common/discovery/staticdiscovery"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/options"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/providers/fab"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/fabsdk/provider/chpvdr"
)

// ProviderFactory represents the default SDK provider factory for services.
type ProviderFactory struct{}

// NewProviderFactory returns the default SDK provider factory for services.
func NewProviderFactory() *ProviderFactory {
	f := ProviderFactory{}
	return &f
}

// CreateLocalDiscoveryProvider returns a static local discovery provider. This should be changed
// to use the dynamic provider when Fabric 1.1 is no longer supported
func (f *ProviderFactory) CreateLocalDiscoveryProvider(config fab.EndpointConfig) (fab.LocalDiscoveryProvider, error) {
	return discovery.NewLocalProvider(config)
}

// CreateChannelProvider returns a new default implementation of channel provider
func (f *ProviderFactory) CreateChannelProvider(config fab.EndpointConfig, opts ...options.Opt) (fab.ChannelProvider, error) {
	return chpvdr.New(config, opts...)
}

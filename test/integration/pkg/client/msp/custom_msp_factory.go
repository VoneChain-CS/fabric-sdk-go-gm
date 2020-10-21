/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/providers/core"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/providers/fab"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/providers/msp"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/fabsdk/factory/defmsp"
)

// ========== MSP Provider Factory with custom user store ============= //

// CustomMSPFactory is a custom factory for tests.
type CustomMSPFactory struct {
	defaultFactory  *defmsp.ProviderFactory
	customUserStore msp.UserStore
}

// NewCustomMSPFactory creates a custom MSPFactory
func NewCustomMSPFactory(customUserStore msp.UserStore) *CustomMSPFactory {
	return &CustomMSPFactory{defaultFactory: defmsp.NewProviderFactory(), customUserStore: customUserStore}
}

// CreateUserStore creates UserStore
func (f *CustomMSPFactory) CreateUserStore(config msp.IdentityConfig) (msp.UserStore, error) {
	return f.customUserStore, nil
}

// CreateIdentityManagerProvider creates an IdentityManager provider
func (f *CustomMSPFactory) CreateIdentityManagerProvider(endpointConfig fab.EndpointConfig, cryptoProvider core.CryptoSuite, userStore msp.UserStore) (msp.IdentityManagerProvider, error) {
	return f.defaultFactory.CreateIdentityManagerProvider(endpointConfig, cryptoProvider, f.customUserStore)
}

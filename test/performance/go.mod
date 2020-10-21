// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/VoneChain-CS/fabric-sdk-go-gm/test/performance

replace github.com/VoneChain-CS/fabric-sdk-go-gm => ../../

require (
	github.com/VoneChain-CS/fabric-sdk-go-gm v0.0.0-00010101000000-000000000000
	github.com/golang/protobuf v1.3.3
	github.com/hyperledger/fabric-protos-go v0.0.0-20200707132912-fee30f3ccd23
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.5.1
	golang.org/x/net v0.0.0-20200421231249-e086a090c8fd
	google.golang.org/grpc v1.29.1
)

go 1.14

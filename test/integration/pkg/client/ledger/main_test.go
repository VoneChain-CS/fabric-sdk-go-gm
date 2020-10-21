/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package ledger

import (
	"testing"

	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/fabsdk"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/test/integration"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/test/integration/util/runner"
)

const (
	org1Name      = "Org1"
	org1AdminUser = "Admin"
)

var mainSDK *fabsdk.FabricSDK
var mainTestSetup *integration.BaseSetupImpl
var mainChaincodeID string

func TestMain(m *testing.M) {
	r := runner.NewWithExampleCC()
	r.Initialize()
	mainSDK = r.SDK()
	mainTestSetup = r.TestSetup()
	mainChaincodeID = r.ExampleChaincodeID()

	r.Run(m)
}

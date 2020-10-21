/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resource

import (
	"testing"

	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/providers/fab"
	contextImpl "github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/context"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/fab/txn"

	"time"

	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/fab/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCreateChaincodeInstallProposal(t *testing.T) {
	ctx := setupContext()
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "peer1.example.com", MockRoles: []string{}, MockCert: nil, Payload: []byte("A"), Status: 200}

	request := ChaincodeInstallRequest{
		Name:    "examplecc",
		Path:    "github.com/examplecc",
		Version: "1",
		Package: &ChaincodePackage{},
	}

	txid, err := txn.NewHeader(ctx, fab.SystemChannel)
	assert.Nil(t, err, "create transaction ID failed")

	prop, err := CreateChaincodeInstallProposal(txid, request)
	assert.Nil(t, err, "CreateChaincodeInstallProposal failed")

	reqCtx, cancel := contextImpl.NewRequest(ctx, contextImpl.WithTimeout(10*time.Second))
	defer cancel()

	_, err = txn.SendProposal(reqCtx, prop, []fab.ProposalProcessor{&peer})
	assert.Nil(t, err, "sending mock proposal failed")
}

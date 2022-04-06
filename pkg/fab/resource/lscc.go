/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resource

import (
	"github.com/VoneChain-CS/fabric-sdk-go-gm/internal/github.com/hyperledger/fabric/protoutil"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/providers/fab"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/fab/txn"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
)

const (
	lscc                    = "lscc"
	lsccInstall             = "install"
	lsccInstalledChaincodes = "getinstalledchaincodes"
	lsccBathCall            = "batchcall"
	lsccUninsatll           = "uninstall"
)

// ChaincodeInstallRequest requests chaincode installation on the network
type ChaincodeInstallRequest struct {
	Name    string
	Path    string
	Version string
	Package *ChaincodePackage
}

// ChaincodePackage contains package type and bytes required to create CDS
type ChaincodePackage struct {
	Type pb.ChaincodeSpec_Type
	Code []byte
}

// CreateChaincodeInstallProposal creates an install chaincode proposal.
func CreateChaincodeInstallProposal(txh fab.TransactionHeader, request ChaincodeInstallRequest) (*fab.TransactionProposal, error) {
	cir, err := createInstallInvokeRequest(request)
	if err != nil {
		return nil, errors.WithMessage(err, "creating lscc install invocation request failed")
	}

	return txn.CreateChaincodeInvokeProposal(txh, cir)
}

func createInstallInvokeRequest(request ChaincodeInstallRequest) (fab.ChaincodeInvokeRequest, error) {
	// Generate arguments for install
	args := [][]byte{}

	ccds := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: request.Package.Type, ChaincodeId: &pb.ChaincodeID{Name: request.Name, Path: request.Path, Version: request.Version}},
		CodePackage: request.Package.Code}

	ccdsBytes, err := protoutil.Marshal(ccds)
	if err != nil {
		return fab.ChaincodeInvokeRequest{}, errors.WithMessage(err, "marshal of chaincode deployment spec failed")
	}
	args = append(args, ccdsBytes)

	cir := fab.ChaincodeInvokeRequest{
		ChaincodeID: lscc,
		Fcn:         lsccInstall,
		Args:        args,
	}
	return cir, nil
}

func createUninstallInvokeRequest(packageID string) fab.ChaincodeInvokeRequest {
	args := [][]byte{[]byte(packageID)}
	cir := fab.ChaincodeInvokeRequest{
		ChaincodeID: lscc,
		Fcn:         lsccUninsatll,
		Args:        args,
	}
	return cir
}

func createBathCallInvokeRequest(args [][]byte) fab.ChaincodeInvokeRequest {
	cir := fab.ChaincodeInvokeRequest{
		ChaincodeID: lscc,
		Fcn:         lsccBathCall,
		Args:        args,
	}
	return cir
}

func createInstalledChaincodesInvokeRequest() fab.ChaincodeInvokeRequest {
	cir := fab.ChaincodeInvokeRequest{
		ChaincodeID: lscc,
		Fcn:         lsccInstalledChaincodes,
	}
	return cir
}

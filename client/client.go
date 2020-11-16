package client

import (
	"bytes"
	"fmt"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/client/resmgmt"
	contextApi "github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/providers/context"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/core/config"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/fab/resource"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/fabsdk"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/vendor/github.com/hyperledger/fabric-config/protolator"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/vendor/github.com/hyperledger/fabric-protos-go/common"
	"github.com/golang/protobuf/proto"
	"time"
)

type dsClientCtx struct {
	org   string
	sdk   *fabsdk.FabricSDK
	clCtx contextApi.ClientProvider
	rsCl  *resmgmt.Client
}

func CreateDSClientCtx(configPath string, org, adminUser string) {
	c := config.FromFile(configPath)
	d := &dsClientCtx{org: org}
	// create SDK with dynamic discovery enabled
	d.sdk, _ = fabsdk.New(c)
	d.clCtx = d.sdk.Context(fabsdk.WithUser(adminUser), fabsdk.WithOrg(org))
	d.rsCl, _ = resmgmt.New(d.clCtx)
}

func GetCurrentChannelConfig(ctx *dsClientCtx, orderer, channelID string) (*common.Config, error) {
	block, err := ctx.rsCl.QueryConfigBlockFromOrderer(channelID, resmgmt.WithOrdererEndpoint(orderer))
	if err != nil {
		return nil, err
	}
	return resource.ExtractConfigFromBlock(block)
}

func SignConfigUpdate(ctx *dsClientCtx, channelID string, proposedConfigJSON string) (*common.ConfigSignature, error) {
	configUpdate, err := GetConfigUpdate(ctx, channelID, proposedConfigJSON)
	if err != nil {
		fmt.Errorf("getConfigUpdate returned error: %s", err)
	}
	configUpdate.ChannelId = channelID

	configUpdateBytes, err := proto.Marshal(configUpdate)
	if err != nil {
		fmt.Errorf("ConfigUpdate marshal returned error: %s", err)
	}

	org1Client, err := ctx.clCtx()
	if err != nil {
		fmt.Errorf("Client provider returned error: %s", err)
	}
	return resource.CreateConfigSignature(org1Client, configUpdateBytes)
}

func GetConfigUpdate(ctx *dsClientCtx, channelID string, proposedConfigJSON string) (*common.ConfigUpdate, error) {

	proposedConfig := &common.Config{}
	err := protolator.DeepUnmarshalJSON(bytes.NewReader([]byte(proposedConfigJSON)), proposedConfig)
	if err != nil {
		return nil, err
	}
	channelConfig, err := GetCurrentChannelConfig(ctx, "", channelID)
	if err != nil {
		return nil, err
	}
	configUpdate, err := resmgmt.CalculateConfigUpdate(channelID, channelConfig, proposedConfig)
	if err != nil {
		return nil, err
	}
	configUpdate.ChannelId = channelID

	return configUpdate, nil
}

func getConfigEnvelopeBytes(configUpdate *common.ConfigUpdate) ([]byte, error) {

	var buf bytes.Buffer
	if err := protolator.DeepMarshalJSON(&buf, configUpdate); err != nil {
		return nil, err
	}

	channelConfigBytes, err := proto.Marshal(configUpdate)
	if err != nil {
		return nil, err
	}
	configUpdateEnvelope := &common.ConfigUpdateEnvelope{
		ConfigUpdate: channelConfigBytes,
		Signatures:   nil,
	}
	configUpdateEnvelopeBytes, err := proto.Marshal(configUpdateEnvelope)
	if err != nil {
		return nil, err
	}
	payload := &common.Payload{
		Data: configUpdateEnvelopeBytes,
	}
	payloadBytes, err := proto.Marshal(payload)
	if err != nil {
		return nil, err
	}
	configEnvelope := &common.Envelope{
		Payload: payloadBytes,
	}

	return proto.Marshal(configEnvelope)
}

func E2eModifyChannel(ordererClCtx *dsClientCtx, org1ClCtx *dsClientCtx, channelID string) error {

	// retrieve channel config
	channelConfig, err := GetCurrentChannelConfig(org1ClCtx, "", channelID)
	if err != nil {
		return err
	}

	// channel config is modified by adding a new application policy.
	// This change must be signed by the majority of org admins.
	// The modified config becomes the proposed channel config.

	// proposed config is distributed to other orgs as JSON string for signing
	var buf bytes.Buffer
	if err := protolator.DeepMarshalJSON(&buf, channelConfig); err != nil {
		fmt.Errorf("DeepMarshalJSON returned error: %s", err)
	}
	proposedChannelConfigJSON := buf.String()

	// org1 calculates and signs config update tx
	signedConfigOrg1, err := SignConfigUpdate(org1ClCtx, channelID, proposedChannelConfigJSON)
	if err != nil {
		fmt.Errorf("error getting signed configuration: %s", err)
	}
	// build config update envelope for constructing channel update request
	configUpdate, err := GetConfigUpdate(org1ClCtx, channelID, proposedChannelConfigJSON)
	if err != nil {
		fmt.Errorf("getConfigUpdate returned error: %s", err)
	}
	configUpdate.ChannelId = channelID
	configEnvelopeBytes, err := getConfigEnvelopeBytes(configUpdate)
	if err != nil {
		fmt.Errorf("error marshaling channel configuration: %s", err)
	}

	// Vefiry that org1 alone cannot sign the change
	configReader := bytes.NewReader(configEnvelopeBytes)
	req := resmgmt.SaveChannelRequest{ChannelID: channelID, ChannelConfig: configReader}
	txID, err := org1ClCtx.rsCl.SaveChannel(req, resmgmt.WithConfigSignatures(signedConfigOrg1), resmgmt.WithOrdererEndpoint("orderer.example.com"))

	fmt.Print(txID)

	// Sign by both orgs and submit tx by the orderer org
	configReader = bytes.NewReader(configEnvelopeBytes)
	req = resmgmt.SaveChannelRequest{ChannelID: channelID, ChannelConfig: configReader}
	txID, err = ordererClCtx.rsCl.SaveChannel(req, resmgmt.WithConfigSignatures(signedConfigOrg1), resmgmt.WithOrdererEndpoint("orderer.example.com"))

	time.Sleep(time.Second * 3)

	// verify updated channel config
	_, err = getCurrentChannelConfig(ordererClCtx, ordererClCtx.org, channelID)
	if err != nil {
		fmt.Errorf("get updated channel config returned error: %s", err)
	}
	return nil
}

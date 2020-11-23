package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	mspclient "github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/client/msp"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/client/resmgmt"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/errors/retry"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/errors/status"
	contextAPI "github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/providers/context"
	contextApi "github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/providers/context"
	fabAPI "github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/providers/fab"
	pmsp "github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/providers/msp"
	contextImpl "github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/context"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/core/config"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/fab/resource"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/fabsdk"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-config/protolator"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type dsClientCtx struct {
	org   string
	sdk   *fabsdk.FabricSDK
	clCtx contextApi.ClientProvider
	rsCl  *resmgmt.Client
}

func Create(orgAdmin, newChannelID, channelConfigPath, configPath, ordererUrl, orderer string) error {

	sdk, _ := fabsdk.New(config.FromFile(configPath))
	rcp := sdk.Context(fabsdk.WithUser(orgAdmin), fabsdk.WithOrg(orderer))
	resMgmtClient, err := resmgmt.New(rcp)

	mspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg(orderer))
	mspClient2, err := mspclient.New(sdk.Context(), mspclient.WithOrg("gm10Org2"))

	if err != nil {
		log.Print(err)
		return err
	}
	//随机选择orderer

	//获取orderer admin签名身份
	adminIdentity, err := mspClient.GetSigningIdentity(orgAdmin)
	adminIdentity2, err := mspClient2.GetSigningIdentity(orgAdmin)

	fmt.Printf("----admin2,%v", adminIdentity2)
	//	fmt.Printf("----admin,%v",adminIdentity)
	req := resmgmt.SaveChannelRequest{
		ChannelID:         newChannelID,                                          //新的通道ID
		ChannelConfigPath: channelConfigPath,                                     //通道配置文件路径 e.g. ./channel-artifacts/channel.tx
		SigningIdentities: []pmsp.SigningIdentity{adminIdentity, adminIdentity2}, //已经弃用
	}

	res, err := resMgmtClient.SaveChannel(
		req,
		resmgmt.WithOrdererEndpoint(ordererUrl), //e.g. orderer.example.com
	)
	if err != nil || "" == res.TransactionID {
		log.Print(err)
		return err
	}
	byteData, _ := json.MarshalIndent(res, "", "\t")
	log.Printf("====== Create Channel ======\n %s\n", string(byteData))
	log.Printf("Create channel successfully.")
	return nil

}

func CreateDSClientCtx(configPath string, org, adminUser string) *dsClientCtx {
	c := config.FromFile(configPath)
	d := &dsClientCtx{org: org}
	// create SDK with dynamic discovery enabled
	d.sdk, _ = fabsdk.New(c)
	d.clCtx = d.sdk.Context(fabsdk.WithUser(adminUser), fabsdk.WithOrg(org))
	d.rsCl, _ = resmgmt.New(d.clCtx)
	return d
}

func Update(configPath string, org, ordererName, output, channelID string) {

	ordererClCtx := CreateDSClientCtx(configPath, org, "Admin")

	channelConfig, err := GetCurrentChannelConfig(ordererClCtx, ordererName, channelID)
	if err != nil {
		fmt.Print(err)
	}

	// channel config is modified by adding a new application policy.
	// This change must be signed by the majority of org admins.
	// The modified config becomes the proposed channel config.

	// proposed config is distributed to other orgs as JSON string for signing
	var buf bytes.Buffer
	if err := protolator.DeepMarshalJSON(&buf, channelConfig); err != nil {
		fmt.Errorf("DeepMarshalJSON returned error: %s", err)
	}

	keyFile := filepath.Join(output, "genesis.json")
	err = ioutil.WriteFile(keyFile, buf.Bytes(), 0600)

}

func GetCurrentChannelConfig(ctx *dsClientCtx, ordererName, channelID string) (*common.Config, error) {
	block, err := ctx.rsCl.QueryConfigBlockFromOrderer(channelID, resmgmt.WithOrdererEndpoint(ordererName))
	if err != nil {
		return nil, err
	}
	return resource.ExtractConfigFromBlock(block)
}

func SignConfigUpdate(ctx *dsClientCtx, orderer, channelID string, proposedConfigJSON string) (*common.ConfigSignature, error) {
	configUpdate, err := GetConfigUpdate(ctx, orderer, channelID, proposedConfigJSON)
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

func GetConfigUpdate(ctx *dsClientCtx, ordererName, channelID string, proposedConfigJSONByte string) (*common.ConfigUpdate, error) {

	proposedConfig := &common.Config{}
	err := protolator.DeepUnmarshalJSON(bytes.NewReader([]byte(proposedConfigJSONByte)), proposedConfig)
	if err != nil {
		return nil, err
	}
	channelConfig, err := GetCurrentChannelConfig(ctx, ordererName, channelID)
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

func E2eModifyChannel(channelID string, configPath string, orderer, ordererName, txPath string) error {

	sdk, _ := fabsdk.New(config.FromFile(configPath))

	mspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg("gm10Org2"))

	adminIdentity, err := mspClient.GetSigningIdentity("Admin")

	ordererClCtx := CreateDSClientCtx(configPath, orderer, "Admin")

	channelConfig, _ := GetCurrentChannelConfig(ordererClCtx, ordererName, channelID)

	// channel config is modified by adding a new application policy.
	// This change must be signed by the majority of org admins.
	// The modified config becomes the proposed channel config.

	// proposed config is distributed to other orgs as JSON string for signing
	var buf bytes.Buffer
	if err := protolator.DeepMarshalJSON(&buf, channelConfig); err != nil {
		fmt.Errorf("DeepMarshalJSON returned error: %s", err)
	}
	proposedChannelConfigJSONS, _ := ioutil.ReadFile(txPath)
	proposedChannelConfigJSON := string(proposedChannelConfigJSONS)
	// org1 calculates and signs config update tx
	/*signedConfigOrg1, err := SignConfigUpdate(ordererClCtx, channelID, proposedChannelConfigJSON)
	if err != nil {
		fmt.Errorf("error getting signed configuration: %s", err)
	}
	fmt.Print(signedConfigOrg1)*/
	// build config update envelope for constructing channel update request
	configUpdate, err := GetConfigUpdate(ordererClCtx, ordererName, channelID, proposedChannelConfigJSON)
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
	req := resmgmt.SaveChannelRequest{ChannelID: channelID, ChannelConfig: configReader, SigningIdentities: []pmsp.SigningIdentity{adminIdentity}}
	txID, err := ordererClCtx.rsCl.SaveChannel(req, resmgmt.WithOrdererEndpoint(ordererName))

	if err != nil {
		fmt.Print(err)
	}
	fmt.Print(txID)

	// Sign by both orgs and submit tx by the orderer org
	configReader = bytes.NewReader(configEnvelopeBytes)
	req = resmgmt.SaveChannelRequest{ChannelID: channelID, ChannelConfig: configReader}

	return nil
}

func IsJoin(channelID, configPath, orgName, peer string) {
	// 判断是否已经加入过channel
	// 一个组织内如果有一个peer加入过channel，则被认为已经加入过了
	sdk, _ := fabsdk.New(config.FromFile(configPath))

	rcp := sdk.Context(fabsdk.WithUser("Admin"), fabsdk.WithOrg(orgName))
	client, err := resmgmt.New(rcp)
	if err != nil {
		log.Panicf("Failed to create resource client: %s", err)

	}
	log.Println("Resmgmt client created successfully.")

	res, err := client.QueryChannels(resmgmt.WithTargetEndpoints(peer))
	if err != nil {
		fmt.Print(err)
	} else {
		fmt.Print(res)
	}

}

//发现本地peers
// DiscoverLocalPeers queries the local peers for the given MSP context and returns all of the peers. If
// the number of peers does not match the expected number then an error is returned.
func DiscoverLocalPeers(ctxProvider contextAPI.ClientProvider, expectedPeers int) ([]fabAPI.Peer, error) {
	ctx, err := contextImpl.NewLocal(ctxProvider)
	if err != nil {
		return nil, errors.Wrap(err, "error creating local context")
	}

	discoveredPeers, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			peers, serviceErr := ctx.LocalDiscoveryService().GetPeers()
			if serviceErr != nil {
				return nil, errors.Wrapf(serviceErr, "error getting peers for MSP [%s]", ctx.Identifier().MSPID)
			}
			if len(peers) < expectedPeers {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("Expecting %d peers but got %d", expectedPeers, len(peers)), nil)
			}
			return peers, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return discoveredPeers.([]fabAPI.Peer), nil
}

/*func createOrderDsClientCtx(ordererOrgName ,adminUser string) *dsClientCtx {
	sdk, _ := fabsdk.WithOrg()


	ordererCtx := sdk.Context(fabsdk.WithUser(adminUser), fabsdk.WithOrg(ordererOrgName))

	// create Channel management client for OrdererOrg
	chMgmtClient, _ := resmgmt.New(ordererCtx)

	return &dsClientCtx{
		org:   ordererOrgName,
		sdk:   sdk,
		clCtx: ordererCtx,
		rsCl:  chMgmtClient,
	}
}*/

func JoinChannel(newChannelID, orderer, orgName, configPath string) {

	sdk, err := fabsdk.New(config.FromFile(configPath))
	if err != nil {
		fmt.Printf("Failed to create new SDK: %s\n", err)
		os.Exit(1)
	}
	defer sdk.Close()

	adminContext := sdk.Context(fabsdk.WithUser("Admin"), fabsdk.WithOrg(orgName))

	// Org resource management client
	orgResMgmt, err := resmgmt.New(adminContext)

	if err != nil {
		fmt.Print(err)
	}

	if err := orgResMgmt.JoinChannel(newChannelID, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint(orderer)); err != nil {
		fmt.Print(err)
	}
}

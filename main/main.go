package main

import (
	"fmt"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/client"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/client/channel"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/client/ledger"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/client/msp"
	mspclient "github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/client/msp"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/client/resmgmt"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/errors/retry"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/logging"
	pmsp "github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/providers/msp"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/core/config"
	lcpackager "github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/fab/ccpackager/lifecycle"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/fabsdk"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"log"
	"os"
)

var (
	cc            = ""
	user          = ""
	secret        = ""
	channelName   = ""
	lvl           = logging.DEBUG
	chaincodrPath = "github.com/VoneChain-CS/fabric-gm/scripts/fabric-samples/chaincode/abstore/go"
)

const (
	channelID      = "byfn-sys-channel"
	newChannelID   = "byfn-sys-channel"
	orgName        = "org1"
	orgName2       = "org2"
	orgAdmin       = "Admin"
	ordererOrgName = "OrdererOrg"
	peer1          = "peer0.org1.example.com"
	peer2          = "peer0.org2.example.com"
	configPath     = "/mnt/nfs/vbaas/fabric/networks/gm10/test2.yaml"
)

func queryInstalledCC(sdk *fabsdk.FabricSDK) {
	userContext := sdk.Context(fabsdk.WithUser(user))

	resClient, err := resmgmt.New(userContext)
	if err != nil {
		fmt.Println("Failed to create resmgmt: ", err)
	}

	resp2, err := resClient.QueryInstalledChaincodes()
	if err != nil {
		fmt.Println("Failed to query installed cc: ", err)
	}
	fmt.Println("Installed cc: ", resp2.GetChaincodes())
}

func queryCC(client *channel.Client, name []byte) string {
	var queryArgs = [][]byte{name}
	response, err := client.Query(channel.Request{
		ChaincodeID: cc,
		Fcn:         "query",
		Args:        queryArgs,
	})

	if err != nil {
		fmt.Println("Failed to query: ", err)
	}

	ret := string(response.Payload)
	fmt.Println("Chaincode status: ", response.ChaincodeStatus)
	fmt.Println("Payload: ", ret)
	return ret
}

func invokeCC(client *channel.Client) {
	invokeArgs := [][]byte{[]byte("a"), []byte("b"), []byte("10")}

	_, err := client.Execute(channel.Request{
		ChaincodeID: cc,
		Fcn:         "invoke",
		Args:        invokeArgs,
	})

	if err != nil {
		fmt.Printf("Failed to invoke: %+v\n", err)
	}
}
func initCC(client *channel.Client) {
	invokeArgs := [][]byte{[]byte("a"), []byte("100"), []byte("b"), []byte("100")}

	_, err := client.Execute(channel.Request{
		ChaincodeID: cc,
		Fcn:         "Init",
		Args:        invokeArgs,
		IsInit:      true,
	})

	if err != nil {
		fmt.Printf("Failed to invoke: %+v\n", err)
	}
}

func enrollUser(sdk *fabsdk.FabricSDK) {
	ctx := sdk.Context()
	mspClient, err := msp.New(ctx)
	if err != nil {
		fmt.Printf("Failed to create msp client: %s\n", err)
	}

	_, err = mspClient.GetSigningIdentity(user)
	if err == msp.ErrUserNotFound {
		fmt.Println("Going to enroll user")
		err = mspClient.Enroll(user, msp.WithSecret(secret))

		if err != nil {
			fmt.Printf("Failed to enroll user: %s\n", err)
		} else {
			fmt.Printf("Success enroll user: %s\n", user)
		}

	} else if err != nil {
		fmt.Printf("Failed to get user: %s\n", err)
	} else {
		fmt.Printf("User %s already enrolled, skip enrollment.\n", user)
	}
}

func registerUser(user string, secret string, sdk *fabsdk.FabricSDK) {

	ctxProvider := sdk.Context()

	// Get the Client.
	// Without WithOrg option, it uses default client organization.
	msp1, err := msp.New(ctxProvider)
	if err != nil {
		fmt.Printf("failed to create CA client: %s", err)
	}

	request := &msp.RegistrationRequest{Name: user, Secret: secret, Type: "client", Affiliation: "org1.department1"}
	_, err = msp1.Register(request)
	if err != nil {
		fmt.Printf("Register return error %s", err)
	}

}

func queryChannelConfig(ledgerClient *ledger.Client) {
	resp1, err := ledgerClient.QueryConfig()
	if err != nil {
		fmt.Printf("Failed to queryConfig: %s", err)
	}
	fmt.Println("ChannelID: ", resp1.ID())
	fmt.Println("Channel Orderers: ", resp1.Orderers())
	fmt.Println("Channel Versions: ", resp1.Versions())
}

func queryChannelInfo(ledgerClient *ledger.Client) {
	resp, err := ledgerClient.QueryInfo()
	if err != nil {
		fmt.Printf("Failed to queryInfo: %s", err)
	}
	fmt.Println("BlockChainInfo:", resp.BCI)
	fmt.Println("Endorser:", resp.Endorser)
	fmt.Println("Status:", resp.Status)
}

func setupLogLevel() {
	logging.SetLevel("fabsdk", lvl)
	logging.SetLevel("fabsdk/common", lvl)
	logging.SetLevel("fabsdk/fab", lvl)
	logging.SetLevel("fabsdk/client", lvl)
}

func readInput() {
	if len(os.Args) != 5 {
		fmt.Printf("Usage: main.go <user-name> <user-secret> <channel> <chaincode-name>\n")
		os.Exit(1)
	}
	user = os.Args[1]
	secret = os.Args[2]
	channelName = os.Args[3]
	cc = os.Args[4]
}

func main2() {
	//readInput()
	user = "admin"
	secret = "adminpw"
	channelName = "mychannel"
	cc = "mycc_3"
	fmt.Println("Reading connection profile..")
	c := config.FromFile("/opt/goworkspace/src/github.com/VoneChain-CS/fabric-sdk-go-gm/main/config_test.yaml")
	sdk, err := fabsdk.New(c)
	if err != nil {
		fmt.Printf("Failed to create new SDK: %s\n", err)
		os.Exit(1)
	}
	defer sdk.Close()

	setupLogLevel()
	//registerUser(user,secret,sdk)
	//enrollUser(sdk)

	//prepare context
	adminContext := sdk.Context(fabsdk.WithUser(orgAdmin), fabsdk.WithOrg(ordererOrgName))

	// Org resource management client
	orgResMgmt, err := resmgmt.New(adminContext)

	if err != nil {

	}
	/*label ,ccPkg :=packageCC("/opt/goworkspace/src/github.com/VoneChain-CS/fabric-gm/scripts/fabric-samples/chaincode/abstore/go")
	installCC(label,ccPkg,orgResMgmt)
	packageID := lcpackager.ComputePackageID(label, ccPkg)
	approveCC(cc,packageID,orgResMgmt)*/
	//approveCC(cc,"mycc14:40d82c2d3d346c5d39110fb19b8ba574da67efcbf5751608e89f2b6c46217531,",orgResMgmt)

	//commitCC(orgResMgmt)
	//queryInstalled(cc,"mycc_1:e3f65f810b94ef30acada1caaf823e2e919d97df56672493e65d8fa0fcad4d6c",orgResMgmt)
	//queryApprovedCC(orgResMgmt)

	joinChannel(orgResMgmt)
	//QueryConfigFromOrderer(channelID,orgResMgmt)
	//QueryInstantiatedChaincodes(channelID,orgResMgmt)
	//checkCCCommitReadiness("mycc_1:e3f65f810b94ef30acada1caaf823e2e919d97df56672493e65d8fa0fcad4d6c",orgResMgmt)
	//queryCommittedCC(orgResMgmt)
	//clientContext allows creation of transactions using the supplied identity as the credential.
	/*clientContext := sdk.Context(fabsdk.WithUser(orgAdmin), fabsdk.WithOrg(ordererOrgName))

		// Resource management client is responsible for managing channels (create/update channel)
		// Supply user that has privileges to create channel (in this case orderer admin)
		resMgmtClient, err := resmgmt.New(clientContext)
		if err != nil {
			log.Print(err)
		}

	createChannel(sdk,resMgmtClient)*/

	/*clientChannelContext := sdk.ChannelContext(channelName, fabsdk.WithUser(user))
	ledgerClient, err := ledger.New(clientChannelContext)
	if err != nil {
		fmt.Printf("Failed to create channel [%s] client: %#v", channelName, err)
		os.Exit(1)
	}

	fmt.Printf("\n===== Channel: %s ===== \n", channelName)
	queryChannelInfo(ledgerClient)
	queryChannelConfig(ledgerClient)*/

	/*	fmt.Println("\n====== Chaincode =========")

		client, err := channel.New(clientChannelContext)
		if err != nil {
			fmt.Printf("Failed to create channel [%s]:", channelName, err)
		}

		initCC(client)
		old := queryCC(client, []byte("a"))

		fmt.Println(old)*/
	fmt.Println("Done.")
}

func main10() {
	client.Update(configPath, "gm10Org2", "orderer.gm10.vbaas.com", "/mnt/nfs/vbaas/fabric/networks/gm10/channel-artifacts/", "successchannel")
}
func main() {
	setupLogLevel()
	client.Create("Admin", "successchannel", "/mnt/nfs/vbaas/fabric/networks/gm10/channel-artifacts/gm10Org2MSPanchors.tx", configPath, "orderer.gm10.vbaas.com", "ordererorg")
}
func main6() {
	client.E2eModifyChannel("successchannel", configPath, "gm10Org2", "orderer.gm10.vbaas.com", "/mnt/nfs/vbaas/fabric/networks/gm10/channel-artifacts/modified_config2.json")
}

func main3() {
	client.JoinChannel("successchannel", "orderer.gm10.vbaas.com", "gm10Org2", configPath)
}
func main1() {
	client.IsJoin("successchannel", configPath, "gm10Org2", "peer0.gm10Org2.gm10.vbaas.com")
}
func packageCC(path string) (string, []byte) {
	desc := &lcpackager.Descriptor{
		Path:  path,
		Type:  pb.ChaincodeSpec_GOLANG,
		Label: cc,
	}
	ccPkg, err := lcpackager.NewCCPackage(desc)
	if err != nil {
		fmt.Print(err)
	}
	return desc.Label, ccPkg
}

func installCC(label string, ccPkg []byte, orgResMgmt *resmgmt.Client) {
	installCCReq := resmgmt.LifecycleInstallCCRequest{
		Label:   label,
		Package: ccPkg,
	}

	resp, err := orgResMgmt.LifecycleInstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		fmt.Print(err)
	}
	fmt.Print(resp)
}

func approveCC(ccID string, packageID string, orgResMgmt *resmgmt.Client) {
	//ccPolicy, _ := policydsl.FromString("OR('Org1MSP.member')")
	approveCCReq := resmgmt.LifecycleApproveCCRequest{

		Name:      ccID,
		Version:   "1",
		PackageID: packageID,
		Sequence:  1,
		/*		EndorsementPlugin: "escc",
				ValidationPlugin:  "vscc",
				SignaturePolicy:   ccPolicy,*/
		InitRequired: true,
	}

	txnID, err := orgResMgmt.LifecycleApproveCC(channelID, approveCCReq, resmgmt.WithTargetEndpoints(peer1), resmgmt.WithOrdererEndpoint("orderer.example.com"), resmgmt.WithRetry(retry.DefaultResMgmtOpts))

	if err != nil {
		fmt.Print(err)
	}

	fmt.Print(txnID)
}

func queryInstalled(label string, packageID string, orgResMgmt *resmgmt.Client) {
	resp, err := orgResMgmt.LifecycleQueryInstalledCC(resmgmt.WithTargetEndpoints(peer1), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		fmt.Print(err)
	}
	fmt.Print(resp)
}

func queryApprovedCC(orgResMgmt *resmgmt.Client) {
	queryApprovedCCReq := resmgmt.LifecycleQueryApprovedCCRequest{
		Name:     cc,
		Sequence: 1,
	}
	resp, err := orgResMgmt.LifecycleQueryApprovedCC(channelID, queryApprovedCCReq, resmgmt.WithTargetEndpoints(peer1), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		fmt.Print(err)
	}
	fmt.Print(resp)
}

func getInstalledCCPackage(packageID string, orgResMgmt *resmgmt.Client) []byte {
	resp, err := orgResMgmt.LifecycleGetInstalledCCPackage(packageID, resmgmt.WithTargetEndpoints(peer1), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		fmt.Print(err)
		return nil
	}
	return resp
}

func commitCC(orgResMgmt *resmgmt.Client) {
	//ccPolicy, _ := policydsl.FromString("OR('Org2MSP.member')")
	req := resmgmt.LifecycleCommitCCRequest{
		Name:     cc,
		Version:  "1",
		Sequence: 1,
		/*		EndorsementPlugin: "escc",
				ValidationPlugin:  "vscc",
				SignaturePolicy:   ccPolicy,*/
		InitRequired: true,
	}
	txnID, err := orgResMgmt.LifecycleCommitCC(channelID, req, resmgmt.WithRetry(retry.DefaultResMgmtOpts),
		resmgmt.WithTargetEndpoints(peer1),
		resmgmt.WithOrdererEndpoint("orderer.example.com"))
	if err != nil {
		log.Print(err)
	}
	log.Print(txnID)
}

func createChannel(sdk *fabsdk.FabricSDK, resMgmtClient *resmgmt.Client) {
	mspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg(orgName))
	if err != nil {
		log.Print(err)
	}
	adminIdentity, err := mspClient.GetSigningIdentity(orgAdmin)
	if err != nil {
		log.Print(err)
	}

	req := resmgmt.SaveChannelRequest{ChannelID: "mychannel",
		ChannelConfigPath: "/opt/goworkspace/src/github.com/VoneChain-CS/fabric-gm/scripts/fabric-samples/first-network/channel-artifacts/" + "channel.tx",
		//ChannelConfig: channelConfig,
		SigningIdentities: []pmsp.SigningIdentity{adminIdentity}}
	txID, _ := resMgmtClient.SaveChannel(req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	log.Print(txID)

}

func joinChannel(orgResMgmt *resmgmt.Client) {

	// Org peers join channel
	if err := orgResMgmt.JoinChannel(newChannelID, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com")); err != nil {
		fmt.Print(err)
	}
}

func checkCCCommitReadiness(packageID string, orgResMgmt *resmgmt.Client) {
	//ccPolicy := policydsl.SignedByAnyMember([]string{"Org1MSP"})
	req := resmgmt.LifecycleCheckCCCommitReadinessRequest{
		Name:      cc,
		Version:   "0",
		PackageID: packageID,
		/*EndorsementPlugin: "escc",
		ValidationPlugin:  "vscc",
		SignaturePolicy:   ccPolicy,*/
		Sequence:     1,
		InitRequired: true,
	}
	resp, err := orgResMgmt.LifecycleCheckCCCommitReadiness(channelID, req, resmgmt.WithTargetEndpoints(peer1), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		fmt.Print(err)
	}
	fmt.Print(resp)
}

func QueryConfigFromOrderer(channelID string, orgResMgmt *resmgmt.Client) {
	resp, err := orgResMgmt.QueryConfigFromOrderer(channelID, resmgmt.WithOrdererEndpoint("orderer.example.com"))
	if err != nil {
		fmt.Print(err)
	} else {
		anchorPeers := resp.AnchorPeers()
		for _, v := range anchorPeers {
			log.Printf("getAnchorPeerUrls, Org: %s, Host: %s, Port: %s", v.Org, v.Host, v.Port)
		}
		fmt.Print("----------------")
		fmt.Print(resp)
	}
}
func QueryInstantiatedChaincodes(channelID string, orgResMgmt *resmgmt.Client) {
	resp, err := orgResMgmt.QueryInstantiatedChaincodes(channelID, resmgmt.WithTargetEndpoints(peer1))
	if err != nil {
		fmt.Print(err)
	} else {
		fmt.Print(resp)
	}
}

func queryCommittedCC(orgResMgmt *resmgmt.Client) {
	req := resmgmt.LifecycleQueryCommittedCCRequest{
		Name: cc,
	}
	resp, err := orgResMgmt.LifecycleQueryCommittedCC(channelID, req, resmgmt.WithTargetEndpoints(peer1), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		fmt.Print(err)
	}
	fmt.Print(resp)
}

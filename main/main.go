package main

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/internal/github.com/hyperledger/fabric/common/policydsl"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/client/channel"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/client/ledger"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/client/msp"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/client/resmgmt"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/errors/retry"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/logging"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/core/config"
	lcpackager "github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/fab/ccpackager/lifecycle"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/fabsdk"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/tjfoc/gmsm/sm2"
	"io/ioutil"
	"log"
	"os"
)

var (
	cc            = ""
	user          = ""
	secret        = ""
	channelName   = ""
	lvl           = logging.INFO
	chaincodrPath = "github.com/VoneChain-CS/fabric-gm/scripts/fabric-samples/chaincode/abstore/go"
)

const (
	channelID      = "mychannel"
	orgName        = "Org1"
	orgName2       = "Org2"
	orgAdmin       = "Admin"
	ordererOrgName = "OrdererOrg"
	peer1          = "peer0.org1.example.com"
	peer2          = "peer0.org2.example.com"
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

func registerUser(sdk *fabsdk.FabricSDK) {

	ctxProvider := sdk.Context()

	// Get the Client.
	// Without WithOrg option, it uses default client organization.
	msp1, err := msp.New(ctxProvider)
	if err != nil {
		fmt.Printf("failed to create CA client: %s", err)
	}

	request := &msp.RegistrationRequest{Name: "testuser", Secret: "testuserpw", Type: "client", Affiliation: "org1.department1"}
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

func main1() {
	cert, err := sm2.ReadCertificateFromPem("/opt/goworkspace/src/github.com/VoneChain-CS/fabric-sdk-go-gm/main/cert.pem")
	cert1, err := sm2.ReadCertificateFromPem("/opt/goworkspace/src/github.com/VoneChain-CS/fabric-sdk-go-gm/main/cert1.pem")
	log.Printf("cert---,%v", cert)
	bytes, _ := ioutil.ReadFile("/opt/goworkspace/src/github.com/VoneChain-CS/fabric-sdk-go-gm/main/cert.pem")
	log.Printf("cert---,%v", bytes)
	if err != nil {
		fmt.Printf("failed to read cert file")
	}
	err = cert1.CheckSignature(cert.SignatureAlgorithm, cert.RawTBSCertificate, cert.Signature)
	pub := cert.PublicKey
	puk := pub.(*ecdsa.PublicKey)
	fmt.Printf("puk.X", puk.X)
	fmt.Printf("puk.Y", puk.Y)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("CheckSignature ok\n")
	}
}

func main() {
	//readInput()
	user = "admin"
	secret = "adminpw"
	channelName = "mychannel"
	cc = "mycc2"
	fmt.Println("Reading connection profile..")
	c := config.FromFile("/opt/goworkspace/src/github.com/VoneChain-CS/fabric-sdk-go-gm/main/config_test.yaml")
	sdk, err := fabsdk.New(c)
	if err != nil {
		fmt.Printf("Failed to create new SDK: %s\n", err)
		os.Exit(1)
	}
	defer sdk.Close()

	setupLogLevel()
	//registerUser(sdk)
	enrollUser(sdk)

	//prepare context
	adminContext := sdk.Context(fabsdk.WithUser(user), fabsdk.WithOrg(orgName))

	// Org resource management client
	orgResMgmt, err := resmgmt.New(adminContext)

	if err != nil {

	}
	/*	label ,ccPkg :=packageCC("/opt/goworkspace/src/github.com/VoneChain-CS/fabric-gm/scripts/fabric-samples/chaincode/abstore/go")
	    installCC(label,ccPkg,orgResMgmt)
		packageID := lcpackager.ComputePackageID(label, ccPkg)
		approveCC(cc,packageID,orgResMgmt)*/
	//approveCC(cc,"mycc14:40d82c2d3d346c5d39110fb19b8ba574da67efcbf5751608e89f2b6c46217531,",orgResMgmt)

	commitCC(orgResMgmt)
	//queryInstalled(cc,"mycc13:85304300b7945ef1516aa2196c3c9bca25c712a2266a99ce47b6ae44cf159e6a",orgResMgmt)
	//queryApprovedCC(orgResMgmt)
	/*clientChannelContext := sdk.ChannelContext(channelName, fabsdk.WithUser(user))
	ledgerClient, err := ledger.New(clientChannelContext)
	if err != nil {
		fmt.Printf("Failed to create channel [%s] client: %#v", channelName, err)
		os.Exit(1)
	}

	fmt.Printf("\n===== Channel: %s ===== \n", channelName)
	queryChannelInfo(ledgerClient)
	queryChannelConfig(ledgerClient)

	fmt.Println("\n====== Chaincode =========")

	client, err := channel.New(clientChannelContext)
	if err != nil {
		fmt.Printf("Failed to create channel [%s]:", channelName, err)
	}

	invokeCC(client)
	old := queryCC(client, []byte("a"))

	fmt.Println(old)*/
	fmt.Println("Done.")
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
	ccPolicy, _ := policydsl.FromString("OR('Org1MSP.member','Org2MSP.member')")
	approveCCReq := resmgmt.LifecycleApproveCCRequest{

		Name:              ccID,
		Version:           "1",
		PackageID:         packageID,
		Sequence:          1,
		EndorsementPlugin: "escc",
		ValidationPlugin:  "vscc",
		SignaturePolicy:   ccPolicy,
		InitRequired:      true,
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
	ccPolicy, _ := policydsl.FromString("OR('Org1MSP.member','Org2MSP.member')")
	req := resmgmt.LifecycleCommitCCRequest{
		Name:              cc,
		Version:           "1",
		Sequence:          1,
		EndorsementPlugin: "escc",
		ValidationPlugin:  "vscc",
		SignaturePolicy:   ccPolicy,
		InitRequired:      true,
	}
	txnID, err := orgResMgmt.LifecycleCommitCC(channelID, req, resmgmt.WithRetry(retry.DefaultResMgmtOpts),
		resmgmt.WithTargetEndpoints(peer1, peer2),
		resmgmt.WithOrdererEndpoint("orderer.example.com"))
	if err != nil {
		log.Print(err)
	}
	log.Print(txnID)
}

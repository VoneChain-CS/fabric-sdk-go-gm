package main

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/tjfoc/gmsm/sm2"
	"io/ioutil"
	"log"
	"os"
)

var (
	cc          = ""
	user        = ""
	secret      = ""
	channelName = ""
	lvl         = logging.INFO
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
	cert, err := sm2.ReadCertificateFromPem("/opt/goworkspace/src/github.com/hyperledger/fabric-sdk-go/main/cert.pem")
	cert1, err := sm2.ReadCertificateFromPem("/opt/goworkspace/src/github.com/hyperledger/fabric-sdk-go/main/cert1.pem")
	log.Printf("cert---,%v", cert)
	bytes, _ := ioutil.ReadFile("/opt/goworkspace/src/github.com/hyperledger/fabric-sdk-go/main/cert.pem")
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
	cc = "mycc"
	fmt.Println("Reading connection profile..")
	c := config.FromFile("/opt/goworkspace/src/github.com/hyperledger/fabric-sdk-go/main/config_test.yaml")
	sdk, err := fabsdk.New(c)
	if err != nil {
		fmt.Printf("Failed to create new SDK: %s\n", err)
		os.Exit(1)
	}
	defer sdk.Close()

	setupLogLevel()
	enrollUser(sdk)

	clientChannelContext := sdk.ChannelContext(channelName, fabsdk.WithUser(user))
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

	fmt.Println(old)
	fmt.Println("Done.")
}

package main

import (
	"fmt"
	"log"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/lilingj/lilingj-app/service"
)

func main() {

	service.InstallExternalCCToMychannelForTowOrgs()

	//ccPkg, err := service.PackageCC("chaincode/marriage", "golang", "marriage_1")
	//ioutil.WriteFile("abc.tar.gz", ccPkg, 0644)
	//service.InstallExternalCCToMychannelForTowOrgs()
	// service.InstallCCToMychannelForTowOrgs()

	panic("===================================================")

	configPath := "config/sdk-config-org1.yaml"

	sdk, err := fabsdk.New(config.FromFile(configPath))
	if err != nil {
		log.Fatal("sdk 初始化失败：", err)
	}
	defer sdk.Close()

	channelClient, err := channel.New(sdk.ChannelContext("mychannel", fabsdk.WithUser("User1"), fabsdk.WithOrg("org2")))
	if err != nil {
		log.Fatal("channelClient 初始化失败：", err)
	}

	rcp := sdk.Context(fabsdk.WithUser("Admin"), fabsdk.WithOrg("org1"))
	rc, err := resmgmt.New(rcp)
	if err != nil {
		log.Fatal("rc 初始化失败: ", err)
	}

	mspClient, err := msp.New(sdk.Context(), msp.WithCAInstance("ca.org3.example.com"))
	if err != nil {
		log.Fatal("mspClient 初始化失败：", err)
	}

	fuckyou(channelClient, rc, mspClient)

}

func fuckyou(...interface{}) {
}

func byte2string1(bs [][]byte) []string {
	ret := []string{}
	for _, b := range bs {
		ret = append(ret, string(b))
	}
	return ret
}

func exeuteCCTest() {
	configPath := "config/sdk-config-org1.yaml"

	sdk, err := fabsdk.New(config.FromFile(configPath))
	if err != nil {
		log.Fatal("sdk 初始化失败：", err)
	}
	defer sdk.Close()

	channelClient, err := channel.New(sdk.ChannelContext("mychannel", fabsdk.WithUser("User1"), fabsdk.WithOrg("org2")))
	if err != nil {
		log.Fatal("channelClient 初始化失败：", err)
	}

	resp, err := channelClient.Execute(channel.Request{ChaincodeID: "mycc", Fcn: "invoke", Args: [][]byte{[]byte("a"), []byte("b"), []byte("10")}},
		channel.WithRetry(retry.DefaultChannelOpts),
		channel.WithTargetEndpoints("peer1.org1.example.com", "peer0.org2.example.com"),
	)
	if err != nil {
		log.Fatal("链码执行失败", err)
	}

	for i, res := range resp.Responses {

		fmt.Println("--------------------------------第" + fmt.Sprint(i) + "--------------------------------------")
		fmt.Println("Endoreser            = " + res.Endorser)
		fmt.Println("Status               = " + fmt.Sprint(res.Status))
		fmt.Println("ChaincodeStatus      = " + fmt.Sprint(res.ChaincodeStatus))
		fmt.Println("Endorsement.Endorser = ", string(res.Endorsement.Endorser))
		fmt.Println("Endorsement.Signature= ", string(res.Endorsement.Signature))
		fmt.Println("--------------------------------第" + fmt.Sprint(i) + "--------------------------------------")
	}
}

func getACertTest() {
	configPath := "config/org3WithCATest.yaml"
	username := "lilingj"
	password := "lilingjpw"
	caID := "ca.org3.example.com"

	sdk, err := fabsdk.New(config.FromFile(configPath))
	if err != nil {
		log.Fatal("sdk 初始化失败：", err)
	}

	mspClient, err := msp.New(sdk.Context(), msp.WithCAInstance(caID))
	if err != nil {
		log.Fatal("mspClient 初始化失败：", err)
	}

	mspClient.Enroll(username, msp.WithSecret(password))
	resp, err := mspClient.GetSigningIdentity(username)
	if err != nil {
		log.Fatal("获取身份失败：", err)
	}

	priv, err := resp.PrivateKey().Bytes()
	publ := resp.PublicVersion().EnrollmentCertificate()
	log.Println("私钥：" + string(priv))
	log.Println("证书：" + string(publ))
}

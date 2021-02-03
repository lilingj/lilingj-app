package service

import (
	"fmt"
	"log"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"

	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
)

/*
安装外部链码
*/
func InstallExternalCCToMychannelForTowOrgs() {
	sdkPath := "config/sdk-config-org1.yaml"
	ccLabel := "ex_marriage_1"
	ccAddress := "marriage.example.com:9999"
	// ccPolicyStr := "AND('Org2.admin')"
	// ccPolicyStr := "OR('Org1.member', 'Org2.member')"
	ccPolicyStr := "OR('Org1MSP.member')"
	peer0org1 := "peer0.org1.example.com"
	peer0org2 := "peer0.org2.example.com"
	ordererURL := "orderer.example.com"
	channelID := "mychannel"
	ccID := "marriage"
	version := "1"

	// 1.获取与 fabric 交互的相关资源： sdk、org1rc、org2rc、
	sdk, err := fabsdk.New(config.FromFile(sdkPath))
	if err != nil {
		log.Panicf("sdk 创建错误: %s", err)
	}
	defer sdk.Close()

	org1rc, err := resmgmt.New(sdk.Context(fabsdk.WithUser("Admin"), fabsdk.WithOrg("Org1")))
	if err != nil {
		log.Panicf("org1rc 创建错误: %s", err)
	}
	org2rc, err := resmgmt.New(sdk.Context(fabsdk.WithUser("Admin"), fabsdk.WithOrg("Org2")))
	if err != nil {
		log.Panicf("org2rc 创建错误: %s", err)
	}

	// 2.打包外部链码
	ccPkg, err := PackageExternalCC(ccLabel, ccAddress)
	if err != nil {
		log.Fatal("PackageExternalCC 错误：", err)
	}
	log.Println("链码打包成功")

	// 3.在两个组织所有 peer 节点上安装链码
	err = InstallCC(org1rc, ccLabel, ccPkg)
	if err != nil {
		log.Fatal("Org1 install CC 错误：", err)
	}
	log.Println("org1 所属 peer 节点全部安装成功")
	err = InstallCC(org2rc, ccLabel, ccPkg)
	if err != nil {
		log.Fatal("Org2 install CC 错误：", err)
	}
	log.Println("org2 所属 peer 节点全部安装成功")

	// 4.获取已经安装的包ID
	packageID := GetPackageID(ccLabel, ccPkg)

	// 5.生成策略实例
	ccPolicy, err := GeneratePolicyFromString(ccPolicyStr)
	if err != nil {
		log.Fatal("策略生成错误！")
	}
	log.Println("策略生成成功")

	// 6.两个组织分别批准
	err = ApproveCC(org1rc, ccPolicy, ordererURL, channelID, ccID, version, packageID, []string{peer0org1}, 1, true)
	if err != nil {
		log.Fatal("Org1 approveCC 错误：", err)
	}
	log.Println("Org1 批准成功")
	err = ApproveCC(org2rc, ccPolicy, ordererURL, channelID, ccID, version, packageID, []string{peer0org2}, 1, true)
	if err != nil {
		log.Fatal("Org2 approveCC 错误：", err)
	}
	log.Println("Org2 批准成功")

	// 7.随便找个人提交到通道
	err = CommitCC(org1rc, ccPolicy, ordererURL, channelID, ccID, version, packageID, 1, true)
	if err != nil {
		log.Fatal("提交错误：", err)
	}
	log.Println("链码已经提交到通道")

	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	fmt.Println("!即将进入链码调用测试环节，请手动启动链码                                                                          !")
	fmt.Println("!外部链码的地址应该是      :", ccAddress)
	fmt.Println("!外部链码的packageID应该是:", packageID)
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	for true {
		ans := ""
		fmt.Println("请问外部链码是否已经能正常访问?[y/n]:")
		fmt.Scanf("%s", &ans)
		if ans == "y" {
			break
		}
	}

	// 8.获得一个与通道交互的资源客户端 cc
	ccp := sdk.ChannelContext(channelID, fabsdk.WithUser("Admin"), fabsdk.WithOrg("Org2"))
	cc, err := channel.New(ccp)
	if err != nil {
		log.Fatal("cc 创建错误：", err)
	}

	// 9.批准时指定了链码要初始化才能用，此处初始化
	resp, err := InitCC(cc, ccID, []string{"InitLedger"})
	if err != nil {
		log.Fatal("初始化链码错误: ", err)
	}
	log.Println("链码已经完成初始化")

	// 看看哪些人在为initCC交易背书
	for i, res := range resp.Responses {
		fmt.Println("--------------------------------第" + fmt.Sprint(i) + "--------------------------------------")
		fmt.Println("Endoreser            = " + res.Endorser)
		fmt.Println("Status               = " + fmt.Sprint(res.Status))
		fmt.Println("ChaincodeStatus      = " + fmt.Sprint(res.ChaincodeStatus))
		fmt.Println("--------------------------------第" + fmt.Sprint(i) + "--------------------------------------")
	}

	// 10.查询链码
	ans, err := QueryCC(cc, ccID, []string{"QueryAllLovers"})
	if err != nil {
		log.Fatal("查询失败！", err)
	}
	log.Println("链码查询结果：" + string(ans))

}

/*
在两个组织节点上安装链码
*/

func InstallCCToMychannelForTowOrgs() {
	sdkPath := "config/sdk-config-org1.yaml"
	ccPath := "chaincode/marriage"
	ccType := "golang"
	ccLabel := "marriage_1"
	// ccPolicyStr := "AND('Org2.admin')"
	// ccPolicyStr := "OR('Org1.member', 'Org2.member')"
	ccPolicyStr := "OR('Org1MSP.member')"
	peer0org1 := "peer0.org1.example.com"
	peer0org2 := "peer0.org2.example.com"
	ordererURL := "orderer.example.com"
	channelID := "mychannel"
	ccID := "marriage"
	version := "1"

	// 1.获取与 fabric 交互的相关资源： sdk、org1rc、org2rc、
	sdk, err := fabsdk.New(config.FromFile(sdkPath))
	if err != nil {
		log.Panicf("sdk 创建错误: %s", err)
	}
	defer sdk.Close()

	org1rc, err := resmgmt.New(sdk.Context(fabsdk.WithUser("Admin"), fabsdk.WithOrg("Org1")))
	if err != nil {
		log.Panicf("org1rc 创建错误: %s", err)
	}
	org2rc, err := resmgmt.New(sdk.Context(fabsdk.WithUser("Admin"), fabsdk.WithOrg("Org2")))
	if err != nil {
		log.Panicf("org2rc 创建错误: %s", err)
	}

	// 2.打包链码
	ccPkg, err := PackageCC(ccPath, ccType, ccLabel)
	if err != nil {
		log.Fatal("PackageCC 错误：", err)
	}
	log.Println("链码打包成功")

	// 3.在两个组织所有 peer 节点上安装链码
	err = InstallCC(org1rc, ccLabel, ccPkg)
	if err != nil {
		log.Fatal("Org1 install CC 错误：", err)
	}
	log.Println("org1 所属 peer 节点全部安装成功")
	err = InstallCC(org2rc, ccLabel, ccPkg)
	if err != nil {
		log.Fatal("Org2 install CC 错误：", err)
	}
	log.Println("org2 所属 peer 节点全部安装成功")

	// 4.获取已经安装的包ID
	packageID := GetPackageID(ccLabel, ccPkg)

	// 5.生成策略实例
	ccPolicy, err := GeneratePolicyFromString(ccPolicyStr)
	if err != nil {
		log.Fatal("策略生成错误！")
	}
	log.Println("策略生成成功")

	// 6.两个组织分别批准
	err = ApproveCC(org1rc, ccPolicy, ordererURL, channelID, ccID, version, packageID, []string{peer0org1}, 1, true)
	if err != nil {
		log.Fatal("Org1 approveCC 错误：", err)
	}
	log.Println("Org1 批准成功")
	err = ApproveCC(org2rc, ccPolicy, ordererURL, channelID, ccID, version, packageID, []string{peer0org2}, 1, true)
	if err != nil {
		log.Fatal("Org2 approveCC 错误：", err)
	}
	log.Println("Org2 批准成功")

	// 7.随便找个人提交到通道
	err = CommitCC(org1rc, ccPolicy, ordererURL, channelID, ccID, version, packageID, 1, true)
	if err != nil {
		log.Fatal("提交错误：", err)
	}
	log.Println("链码已经提交到通道")

	// 8.获得一个与通道交互的资源客户端 cc
	ccp := sdk.ChannelContext(channelID, fabsdk.WithUser("Admin"), fabsdk.WithOrg("Org2"))
	cc, err := channel.New(ccp)
	if err != nil {
		log.Fatal("cc 创建错误：", err)
	}

	// 9.批准时指定了链码要初始化才能用，此处初始化
	resp, err := InitCC(cc, ccID, []string{"InitLedger"})
	if err != nil {
		log.Fatal("初始化链码错误: ", err)
	}
	log.Println("链码已经完成初始化")

	// 看看哪些人在为initCC交易背书
	for i, res := range resp.Responses {
		fmt.Println("--------------------------------第" + fmt.Sprint(i) + "--------------------------------------")
		fmt.Println("Endoreser            = " + res.Endorser)
		fmt.Println("Status               = " + fmt.Sprint(res.Status))
		fmt.Println("ChaincodeStatus      = " + fmt.Sprint(res.ChaincodeStatus))
		fmt.Println("--------------------------------第" + fmt.Sprint(i) + "--------------------------------------")
	}

	// 10.查询链码
	ans, err := QueryCC(cc, ccID, []string{"QueryAllLovers"})
	if err != nil {
		log.Fatal("查询失败！", err)
	}
	log.Println("链码查询结果：" + string(ans))

}

func CreateAndJoinChannel() {
	sdk, err := fabsdk.New(config.FromFile("config/sdk-config-org1.yaml"))
	if err != nil {
		log.Panicf("sdk 创建错误: %s", err)
	}
	defer sdk.Close()

	mspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg("Org1"))
	if err != nil {
		log.Panicf("msp 创建错误: %s", err)
	}

	rcp := sdk.Context(fabsdk.WithUser("Admin"), fabsdk.WithOrg("OrdererOrg"))
	rc_orderer, err := resmgmt.New(rcp)
	if err != nil {
		log.Panicf("rc_orderer 创建错误: %s", err)
	}

	err = CreateChannel(*mspClient, rc_orderer, "orderer.example.com", "Admin", "lljchannel", "channel-artifacts/lljchannel.tx")
	if err != nil {
		log.Panicf("通道创建错误: %s", err)
	}

	rcp = sdk.Context(fabsdk.WithUser("Admin"), fabsdk.WithOrg("Org1"))
	rc_org1, err := resmgmt.New(rcp)
	if err != nil {
		log.Panicf("rc_org1 创建错误: %s", err)
	}

	err = JoinChannel(rc_org1, "lljchannel", "orderer.example.com")
	if err != nil {
		log.Panicf("Org1's peer 加入错误： %s", err)
	}
}

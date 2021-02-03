package service

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"log"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	lcpackager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/lifecycle"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"

	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/policydsl"

	cb "github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

// GetPP return abspath(".")
func GetPP() string {
	return metadata.GetProjectPath()
}

func GetPackageID(label string, ccPkg []byte) string {
	return lcpackager.ComputePackageID(label, ccPkg)
}

/*
func createCCLifecycle(orgResMgmt *resmgmt.Client, sdk *fabsdk.FabricSDK) {
	// Package cc
	label, ccPkg := PackageCC()
	packageID := lcpackager.ComputePackageID(label, ccPkg)

	// Install cc
	installCC(t, label, ccPkg, orgResMgmt)

	// Get installed cc package
	getInstalledCCPackage(t, packageID, ccPkg, orgResMgmt)

	// Query installed cc
	queryInstalled(t, label, packageID, orgResMgmt)

	// Approve cc
	approveCC(t, packageID, orgResMgmt)

	// Query approve cc
	queryApprovedCC(t, orgResMgmt)

	// Check commit readiness
	checkCCCommitReadiness(t, orgResMgmt)

	// Commit cc
	commitCC(t, orgResMgmt)

	// Query committed cc
	queryCommittedCC(t, orgResMgmt)

	// Init cc
	initCC(t, sdk)

}
*/

// PackageCC return label, ccPkg, err
func PackageCC(ccPath, ccType, ccLabel string) (ccPkg []byte, err error) {

	_ccType := pb.ChaincodeSpec_UNDEFINED
	switch ccType {
	case "go", "Go", "golang", "Golang":
		_ccType = pb.ChaincodeSpec_GOLANG
	case "node", "Node", "node.js", "Node.js":
		_ccType = pb.ChaincodeSpec_NODE
	case "java", "Java":
		_ccType = pb.ChaincodeSpec_JAVA
	default:
		return nil, errors.New("不支持该语言链码: " + ccType)
	}

	desc := &lcpackager.Descriptor{
		Path:  ccPath,
		Type:  _ccType,
		Label: ccLabel,
	}
	ccPkg, err = lcpackager.NewCCPackage(desc)
	return ccPkg, err
}

/*
PackageExternalCC：
生成connection.json，把该connection.json打包成code.tar.gz
生成metadata.json文件，把metadata.json和code.tar.gz打包成文件，return 这个文件
默认不使用tls与链码通信
*/
func PackageExternalCC(label, address string) (ccPkg []byte, err error) {
	payload1 := bytes.NewBuffer(nil)
	gw1 := gzip.NewWriter(payload1)
	tw1 := tar.NewWriter(gw1)
	content := []byte(`{
		"address": "` + address + `",
		"dial_timeout": "10s",
		"tls_required": false,
		"client_auth_required": false,
		"client_key": "-----BEGIN EC PRIVATE KEY----- ... -----END EC PRIVATE KEY-----",
		"client_cert": "-----BEGIN CERTIFICATE----- ... -----END CERTIFICATE-----",
		"root_cert": "-----BEGIN CERTIFICATE---- ... -----END CERTIFICATE-----"
	}`)

	err = writePackage(tw1, "connection.json", content)
	if err != nil {
		return nil, errors.WithMessage(err, "connection.json生成错误")
	}

	err = tw1.Close()
	if err == nil {
		err = gw1.Close()
	}
	if err != nil {
		return nil, errors.WithMessage(err, "connection.json生成错误")
	}

	content = []byte(`{"path":"","type":"external","label":"` + label + `"}`)
	payload2 := bytes.NewBuffer(nil)
	gw2 := gzip.NewWriter(payload2)
	tw2 := tar.NewWriter(gw2)

	writePackage(tw2, "code.tar.gz", payload1.Bytes())
	writePackage(tw2, "metadata.json", content)

	err = tw2.Close()
	if err == nil {
		err = gw2.Close()
	}
	if err != nil {
		return nil, errors.WithMessage(err, "包文件生成错误")
	}

	return payload2.Bytes(), nil
}
func writePackage(tw *tar.Writer, name string, payload []byte) error {
	err := tw.WriteHeader(
		&tar.Header{
			Name: name,
			Size: int64(len(payload)),
			Mode: 0100644,
		},
	)
	if err != nil {
		return err
	}

	_, err = tw.Write(payload)
	return err
}

/*
* 生成 rc 时 指定的组织的peer节点会都安装上
* 我关了关于代理的什么设置，直接安装的时候，连服务器连不上，手动 go mod vendor 连接的是我设置的代理服务器，可以安装
 */
func InstallCC(orgResMgmt *resmgmt.Client, label string, ccPkg []byte) error {
	installCCReq := resmgmt.LifecycleInstallCCRequest{
		Label:   label,
		Package: ccPkg,
	}

	// packageID := lcpackager.ComputePackageID(installCCReq.Label, installCCReq.Package)

	_, err := orgResMgmt.LifecycleInstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))

	return err
	//require.Equal(packageID, resp[0].PackageID)
}

// 查询某peer是否安装某链码
func QueryInstalled(orgResMgmt *resmgmt.Client, label string, packageID string, peerUrl string) (bool, error) {
	resps, err := orgResMgmt.LifecycleQueryInstalledCC(resmgmt.WithTargetEndpoints(peerUrl), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		return false, errors.WithMessage(err, "链码安装查询失败！")
	}
	//log.Printf("\nlabel: %s\npackageID:%s\n", resp[0].Label, resp[0].PackageID)
	for _, resp := range resps {
		//fmt.Println(resp.Label + " " + resp.PackageID)
		if label == resp.Label && packageID == resp.PackageID {
			return true, nil
		}
	}
	return false, nil
	// require.Equal(t, packageID, resp[0].PackageID)
	// require.Equal(t, label, resp[0].Label)
}

func GeneratePolicyFromString(ccPolicyStr string) (*cb.SignaturePolicyEnvelope, error) {
	ccPolicy, err := policydsl.FromString(ccPolicyStr)
	if err != nil {
		return nil, errors.WithMessage(err, "生成策略错误！")
	}
	return ccPolicy, nil
}

/*
peer lifecycle chaincode checkcommitreadiness --channelID mychannel --name marriage --version 1 --sequence 1 --init-required --signature-policy "OR('Org1MSP.member')"
批准链码时指定的这些信息共同标识了链码，查询的时候少了是查不出来的，比如批准的时候InitRequired = true
查询的时候没有--init-required，查不出来；批准的时候指定了策略，查询的时候没有指定策略或者指定的和开始
不一样，都查不出来
*/
func ApproveCC(orgResMgmt *resmgmt.Client, ccPolicy *cb.SignaturePolicyEnvelope, ordererURL, channelID, ccID, version, packageID string, peerURLs []string, sequence int64, initRequired bool) error {
	approveCCReq := resmgmt.LifecycleApproveCCRequest{
		Name:              ccID,
		Version:           version,
		PackageID:         packageID,
		Sequence:          sequence,
		EndorsementPlugin: "escc",
		ValidationPlugin:  "vscc",
		SignaturePolicy:   ccPolicy,     // !!
		InitRequired:      initRequired, // !!
	}

	txnID, err := orgResMgmt.LifecycleApproveCC(channelID, approveCCReq, resmgmt.WithTargetEndpoints(peerURLs...), resmgmt.WithOrdererEndpoint(ordererURL), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil || txnID == "" {
		return errors.WithMessage(err, "批准链码失败")
	}
	//log.Println(txnID)
	return nil
	//require.NotEmpty(t, txnID)
}

func CheckCCCommitReadiness(orgResMgmt *resmgmt.Client, ccPolicy *cb.SignaturePolicyEnvelope, channelID, ccID, version, packageID string, sequence int64, initRequired bool) error {
	req := resmgmt.LifecycleCheckCCCommitReadinessRequest{
		Name:              ccID,
		Version:           version,
		EndorsementPlugin: "escc",
		ValidationPlugin:  "vscc",
		SignaturePolicy:   ccPolicy,
		Sequence:          sequence,
		InitRequired:      initRequired,
	}
	resp, err := orgResMgmt.LifecycleCheckCCCommitReadiness(channelID, req, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		return err
	}
	log.Println(resp.Approvals)
	return nil
}

func CommitCC(orgResMgmt *resmgmt.Client, ccPolicy *cb.SignaturePolicyEnvelope, ordererUrl, channelID, ccID, version, packageID string, sequence int64, initRequired bool) error {
	//ccPolicy := policydsl.SignedByAnyMember([]string{"Org1MSP"})
	req := resmgmt.LifecycleCommitCCRequest{
		Name:              ccID,
		Version:           version,
		Sequence:          sequence,
		EndorsementPlugin: "escc",
		ValidationPlugin:  "vscc",
		SignaturePolicy:   ccPolicy,
		InitRequired:      initRequired,
	}
	txID, err := orgResMgmt.LifecycleCommitCC(channelID, req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint(ordererUrl))
	log.Println(txID)
	return err
}

func QueryCommittedCC(orgResMgmt *resmgmt.Client, ccID, channelID string) error {
	req := resmgmt.LifecycleQueryCommittedCCRequest{
		Name: ccID,
	}
	resps, err := orgResMgmt.LifecycleQueryCommittedCC(channelID, req, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		return err
	}
	log.Println(resps)
	for _, resp := range resps {
		if resp.Name == ccID {
			return nil
		}
	}
	return errors.New("未查询到已提交的链码：" + ccID)
}

// shell批准时指定--init-required，或者sdk批准时指定 InitRequired = true，运行链码时都需要先初始化链码，用--isInit或者IsInit: true
func InitCC(channelClient *channel.Client, ccID string, args []string, targetEndpoints ...string) (channel.Response, error) {
	response, err := channelClient.Execute(channel.Request{ChaincodeID: ccID, Fcn: args[0], Args: packArgs(args[1:]), IsInit: true},
		channel.WithRetry(retry.DefaultChannelOpts),
		channel.WithTargetEndpoints(targetEndpoints...),
	)
	if err != nil {
		return response, errors.WithMessage(err, "链码初始化失败！")
	}
	return response, err
}

func ExecuteCC(channelClient *channel.Client, ccID string, args []string, targetEndpoints ...string) ([]byte, error) {
	response, err := channelClient.Execute(channel.Request{ChaincodeID: ccID, Fcn: args[0], Args: packArgs(args[1:])},
		channel.WithRetry(retry.DefaultChannelOpts),
		channel.WithTargetEndpoints(targetEndpoints...),
	)
	if err != nil {
		return response.Payload, errors.WithMessage(err, "链码执行失败！")
	}
	return response.Payload, err
}

// eg: QueryCC(cc, "mycc", []string{"Query", "a"}, "peer0.org1.example.com")
func QueryCC(channelClient *channel.Client, ccID string, args []string, targetEndpoints ...string) ([]byte, error) {
	response, err := channelClient.Query(channel.Request{ChaincodeID: ccID, Fcn: args[0], Args: packArgs(args[1:])},
		channel.WithRetry(retry.DefaultChannelOpts),
		channel.WithTargetEndpoints(targetEndpoints...),
	)
	if err != nil {
		return response.Payload, errors.WithMessage(err, "链码查询失败！")
	}
	return response.Payload, nil
}

/*
* []string to [][]byte
 */
func packArgs(paras []string) [][]byte {
	var args [][]byte
	for _, k := range paras {
		args = append(args, []byte(k))
	}
	return args
}

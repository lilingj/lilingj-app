package service

import (
	"log"
	"os/exec"
	"path/filepath"

	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/pkg/errors"
)

// HasPeerJoinedChannel checks whether the peer has already joined the channel.
// It returns true if it has, false otherwise, or an error
func HasPeerJoinedChannel(client *resmgmt.Client, target string, channel string) (bool, error) {
	foundChannel := false
	response, err := client.QueryChannels(resmgmt.WithTargetEndpoints(target), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		return false, errors.WithMessage(err, "failed to query channel for peer")
	}
	for _, responseChannel := range response.Channels {
		if responseChannel.ChannelId == channel {
			foundChannel = true
			break
		}
	}

	return foundChannel, nil
}

// FilterTargetsJoinedChannel filters targets to those that have joined the named channel.
// targets must belong to rc
func FilterTargetsJoinedChannel(rc *resmgmt.Client, channelID string, targets []string) ([]string, error) {
	var joinedTargets []string
	for _, target := range targets {
		// Check if primary peer has joined channel
		alreadyJoined, err := HasPeerJoinedChannel(rc, target, channelID)
		if err != nil {
			return nil, errors.WithMessage(err, "failed while checking if primary peer has already joined channel")
		}
		if alreadyJoined {
			joinedTargets = append(joinedTargets, target)
		}
	}
	return joinedTargets, nil
}

// !!!!!!!!!!!!!未完成!!!!!!!!!!!!!!!!!
// 没找到生成channel.tx文件的api
// configtxgen -profile TwoOrgsChannel -configPath $PWD -outputCreateChannelTx lljchannel.tx -channelID lljchannel
func CreateChannelTx(profile, configPath, outputCreateChannelTxPath, channelID string) error {
	command := exec.Command("configtxgen", "-profile", profile, "-configPath", configPath, "-outputCreateChannelTx", outputCreateChannelTxPath, "-channelID", channelID)
	return command.Run()
}

func CreateChannel(mspClient mspclient.Client, resMgmtClient *resmgmt.Client, ordererURL, orgAdmin, channelID, channelConfigTxPath string) error {
	adminIdentity, err := mspClient.GetSigningIdentity(orgAdmin)
	if err != nil {
		return errors.WithMessage(err, "签名错误！")
	}
	req := resmgmt.SaveChannelRequest{ChannelID: channelID,
		ChannelConfigPath: filepath.Join(GetPP(), channelConfigTxPath),
		SigningIdentities: []msp.SigningIdentity{adminIdentity},
	}
	txID, err := resMgmtClient.SaveChannel(req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint(ordererURL))
	log.Println(txID)
	return err
}

func JoinChannel(resMgmtClient *resmgmt.Client, channelID, ordererURL string) error {
	err := resMgmtClient.JoinChannel(channelID, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint(ordererURL))
	return err
}

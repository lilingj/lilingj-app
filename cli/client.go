package cli

import (
	"log"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

type Client struct {
	// Fabric network information
	ConfigPath string
	OrgName    string
	OrgAdmin   string
	OrgUser    string

	// sdk clients
	SDK           *fabsdk.FabricSDK
	ResmgmtClient *resmgmt.Client
	ChannelClient *channel.Client

	// Same for each peer
	ChannelID string
}

func NewClient(configPath, orgName, orgAdmin, orgUser, channelID string) *Client {
	sdk, err := fabsdk.New(config.FromFile(configPath))
	if err != nil {
		log.Panicf("failed to create fabric sdk: %s", err)
	}

	rcp := sdk.Context(fabsdk.WithUser(orgAdmin), fabsdk.WithOrg(orgName))
	rc, err := resmgmt.New(rcp)
	if err != nil {
		log.Panicf("failed getting admin user session for org: %s", err)
	}

	ccp := sdk.ChannelContext(channelID, fabsdk.WithUser(orgAdmin), fabsdk.WithOrg(orgName))
	cc, err := channel.New(ccp)
	if err != nil {
		log.Panicf("failed to create channel client: %s", err)
	}

	return &Client{
		ConfigPath: configPath,
		OrgName:    orgName,
		OrgAdmin:   orgAdmin,
		OrgUser:    orgUser,
		ChannelID:  channelID,

		SDK:           sdk,
		ResmgmtClient: rc,
		ChannelClient: cc,
	}

}

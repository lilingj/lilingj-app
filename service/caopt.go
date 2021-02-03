package service

import (
	"fmt"
	"log"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

//EnrollUser enroll a user have registerd
func EnrollUser(sdk *fabsdk.FabricSDK, username string, password string) (bool, error) {
	ctx := sdk.Context()
	mspClient, err := msp.New(ctx)
	if err != nil {
		log.Printf("Failed to create msp client: %s\n", err)
		return true, err
	}

	resp, err := mspClient.GetSigningIdentity(username)
	log.Println(resp)
	if err == msp.ErrUserNotFound {
		log.Println("Going to enroll user")
		err = mspClient.Enroll(username, msp.WithSecret(password))
		if err != nil {
			log.Printf("Failed to enroll user: %s\n", err)
			return true, err
		}
		log.Printf("Success enroll user: %s\n", username)
	} else if err != nil {
		log.Printf("Failed to get user: %s\n", err)
		return false, err
	}
	log.Printf("User %s already enrolled, skip enrollment.\n", username)
	return true, nil
}

//Register a new user with username and password.
func RegisterUser(sdk *fabsdk.FabricSDK, username, password string) error {
	ctx := sdk.Context()
	mspClient, err := msp.New(ctx)

	if err != nil {
		fmt.Printf("Failed to create msp client: %s\n", err)
	}
	request := &msp.RegistrationRequest{
		Name:   username,
		Type:   "user",
		Secret: password,
	}

	secret, err := mspClient.Register(request)
	if err != nil {
		fmt.Printf("register %s [%s]\n", username, err)
		return err
	}
	fmt.Printf("register %s successfully,with password %s\n", username, secret)
	return nil
}

package utils

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

// Basically a package which holds a bunch of context around the AWS environment
// we are launched within
//
var accountID string
var region = "eu-west-1"

// Populate must be called at some stage to populate all the environment configs
//
func Populate() error {

	sess, _ := session.NewSession(&aws.Config{
		// Region: aws.String("eu-west-1")},
	})

	stsClient := sts.New(sess)

	identity, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return err
	}

	accountID = *identity.Account

	log.Info("Populated values", "AccountID", accountID, "Region", region)

	return nil
}

// AccountID returns the AWS account number
//
func AccountID() string {
	return accountID
}

// Region returns the AWS region
//
func Region() string {
	return region
}

package utils

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
)

// Basically a package which holds a bunch of context around the AWS environment
// we are launched within
//

var accountID string
var region string

// Populate must be called at some stage to populate all the environment configs
//
func Populate() error {

	return nil

	sess, _ := session.NewSession(&aws.Config{
		// Region: aws.String("eu-west-1")},
	})

	ec2metadataClient := ec2metadata.New(sess)

	identity, err := ec2metadataClient.GetInstanceIdentityDocument()
	if err != nil {
		return err
	}

	accountID = identity.AccountID
	region = identity.Region

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

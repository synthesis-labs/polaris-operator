package utils

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

// Basically a package which holds a bunch of context around the AWS environment
// we are launched within
//
var accountID string
var availabilityZone = "eu-west-1a"
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
	ec2metadataClient := ec2metadata.New(sess)

	instanceIdentityDocument, err := ec2metadataClient.GetInstanceIdentityDocument()
	if err != nil {
		log.Error(err, "Unable to get ec2metadata info (possibly running outside of AWS?). Using default values.")
		return nil
	}

	availabilityZone = instanceIdentityDocument.AvailabilityZone

	// Drop the last character of the AZ for the region
	//
	region = availabilityZone[0:len(availabilityZone)]
	log.Info("Populated values", "AccountID", accountID, "AvailabilityZone", availabilityZone, "Region", region)

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

// AvailabilityZone returns the AWS availability zone
//
func AvailabilityZone() string {
	return availabilityZone
}

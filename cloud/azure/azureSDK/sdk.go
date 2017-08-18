package azureSDK

import (
	"os"
	"github.com/aws/aws-sdk-go/service/ec2"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
)

type Sdk struct {

}

func NewSdk() (*Sdk, error) {
	return &Sdk{}, nil
}

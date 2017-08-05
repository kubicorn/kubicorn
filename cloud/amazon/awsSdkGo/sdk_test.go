package awsSdkGo

import (
	"os"
	"testing"
)

var (
	AwsAccessKey       = os.Getenv("AWS_ACCESS_KEY_ID")
	AwsSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
)

func TestMain(m *testing.M) {
	m.Run()
	os.Setenv("AWS_ACCESS_KEY_ID", AwsAccessKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", AwsSecretAccessKey)
}

func TestSdkHappy(t *testing.T) {
	os.Setenv("AWS_ACCESS_KEY_ID", "123")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "123")
	_, err := NewSdk("us-west-2")
	if err != nil {
		t.Fatalf("Unable to get Amazon SDK: %v", err)
	}
}

func TestSdkSad(t *testing.T) {
	os.Setenv("AWS_ACCESS_KEY_ID", "")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "")
	_, err := NewSdk("us-west-2")
	if err == nil {
		t.Fatalf("Able to get Amazon SDK with empty variables")
	}
}

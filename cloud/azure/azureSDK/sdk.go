package azureSDK

import (
	"github.com/Azure/azure-sdk-for-go/arm/examples/helpers"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"os"
	"fmt"
)

type Sdk struct {
	ServicePrincipal *ServicePrincipal
}

type ServicePrincipal struct {
	ClientID           string
	ClientSecret       string
	SubscriptionID     string
	TenantId           string
	HashMap            map[string]string
	AuthenticatedToken *adal.ServicePrincipalToken
}

func NewSdk() (*Sdk, error) {
	clientID := os.Getenv("AZURE_CLIENT_ID")
	if clientID == "" {
		return nil, fmt.Errorf("Empty $AZURE_CLIENT_ID")
	}
	clientSecret := os.Getenv("AZURE_CLIENT_SECRET")
	if clientSecret == "" {
		return nil, fmt.Errorf("Empty $AZURE_CLIENT_SECRET")
	}
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	if subscriptionID == "" {
		return nil, fmt.Errorf("Empty $AZURE_SUBSCRIPTION_ID")
	}
	tenantID := os.Getenv("AZURE_TENANT_ID")
	if tenantID == "" {
		return nil, fmt.Errorf("Empty $AZURE_TENANT_ID")
	}

	sdk := &Sdk{
		ServicePrincipal: &ServicePrincipal{
			ClientID:       clientID,
			ClientSecret:   clientSecret,
			SubscriptionID: subscriptionID,
			TenantId:       tenantID,
			HashMap: map[string]string{
				"AZURE_CLIENT_ID":       clientID,
				"AZURE_CLIENT_SECRET":   clientSecret,
				"AZURE_SUBSCRIPTION_ID": subscriptionID,
				"AZURE_TENANT_ID":       tenantID,
			},
		},
	}

	authenticatedToken, err := helpers.NewServicePrincipalTokenFromCredentials(sdk.ServicePrincipal.HashMap, azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return nil, err
	}
	sdk.ServicePrincipal.AuthenticatedToken = authenticatedToken
	return sdk, nil
}

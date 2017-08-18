package azureSDK

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/arm/examples/helpers"
	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"os"
)

type Sdk struct {
	ServicePrincipal *ServicePrincipal
	Network          *network.ManagementClient
	Vnet             *network.VirtualNetworksClient
	ResourceGroup    *resources.GroupsClient
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

	//-------------------------
	// Azure Client Resources
	//-------------------------

	//-------------------------
	// Resource Group
	resourceGroup := resources.NewGroupsClient(sdk.ServicePrincipal.SubscriptionID)
	resourceGroup.Authorizer = autorest.NewBearerAuthorizer(sdk.ServicePrincipal.AuthenticatedToken)
	sdk.ResourceGroup = &resourceGroup

	//------------------------
	// Network
	networkClient := network.New(sdk.ServicePrincipal.SubscriptionID)
	networkClient.Authorizer = autorest.NewBearerAuthorizer(sdk.ServicePrincipal.AuthenticatedToken)
	sdk.Network = &networkClient

	//------------------------
	// Vnet
	vnetClient := network.NewVirtualNetworksClient(sdk.ServicePrincipal.SubscriptionID)
	vnetClient.Authorizer = autorest.NewBearerAuthorizer(sdk.ServicePrincipal.AuthenticatedToken)
	sdk.Vnet = &vnetClient

	return sdk, nil
}

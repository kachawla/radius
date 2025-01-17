/*
Copyright 2023 The Radius Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package credentials

import (
	"context"
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"

	"github.com/radius-project/radius/pkg/components/secret"
	"github.com/radius-project/radius/pkg/components/secret/secretprovider"
	"github.com/radius-project/radius/pkg/sdk"
	"github.com/radius-project/radius/pkg/to"
	ucpapi "github.com/radius-project/radius/pkg/ucp/api/v20231001preview"
)

var _ CredentialProvider[AzureCredential] = (*AzureCredentialProvider)(nil)

// AzureCredentialProvider is UCP credential provider for Azure.
type AzureCredentialProvider struct {
	secretProvider *secretprovider.SecretProvider
	client         *ucpapi.AzureCredentialsClient
}

// NewAzureCredentialProvider creates a new AzureCredentialProvider by creating a new AzureCredentialClient with the given
// credential and connection, and returns an error if one occurs.
func NewAzureCredentialProvider(provider *secretprovider.SecretProvider, ucpConn sdk.Connection, credential azcore.TokenCredential) (*AzureCredentialProvider, error) {
	cli, err := ucpapi.NewAzureCredentialsClient(credential, sdk.NewClientOptions(ucpConn))
	if err != nil {
		return nil, err
	}

	return &AzureCredentialProvider{
		secretProvider: provider,
		client:         cli,
	}, nil
}

// Fetch fetches the Azure credentials from UCP and the internal storage (e.g. Kubernetes secret store)
// and returns an AzureCredential struct. If an error occurs, an error is returned.
func (p *AzureCredentialProvider) Fetch(ctx context.Context, planeName, name string) (*AzureCredential, error) {
	// 1. Fetch the secret name of Azure credentials from UCP.
	cred, err := p.client.Get(ctx, planeName, name, &ucpapi.AzureCredentialsClientGetOptions{})
	if err != nil {
		return nil, err
	}

	// We support only kubernetes secret, but we may support multiple secret stores.
	var storage *ucpapi.InternalCredentialStorageProperties

	switch p := cred.Properties.(type) {
	case *ucpapi.AzureServicePrincipalProperties:
		storage, err = getStorageProperties(p.Storage)
	case *ucpapi.AzureWorkloadIdentityProperties:
		storage, err = getStorageProperties(p.Storage)
	default:
		return nil, errors.New("Azure Credential is invalid - field 'properties' is not AzureServicePrincipalProperties or AzureWorkloadIdentityProperties")
	}

	if err != nil {
		return nil, err
	}

	secretName := to.String(storage.SecretName)
	if secretName == "" {
		return nil, errors.New("unspecified SecretName for internal storage")
	}

	// 2. Fetch the credential from internal storage (e.g. Kubernetes secret store)
	secretClient, err := p.secretProvider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	s, err := secret.GetSecret[AzureCredential](ctx, secretClient, secretName)
	if err != nil {
		return nil, errors.New("failed to get credential info: " + err.Error())
	}

	return &s, nil
}

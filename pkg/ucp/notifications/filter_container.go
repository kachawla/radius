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

package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	v1 "github.com/radius-project/radius/pkg/armrpc/api/v1"
	"github.com/radius-project/radius/pkg/armrpc/asyncoperation/statusmanager"
	aztoken "github.com/radius-project/radius/pkg/azure/tokencredentials"
	"github.com/radius-project/radius/pkg/cli/clients_new/generated"
	"github.com/radius-project/radius/pkg/sdk"
	"github.com/radius-project/radius/pkg/ucp/dataprovider"
	queue "github.com/radius-project/radius/pkg/ucp/queue/client"
	queueprovider "github.com/radius-project/radius/pkg/ucp/queue/provider"
	"github.com/radius-project/radius/pkg/ucp/resources"
	"github.com/radius-project/radius/pkg/ucp/store"
)

type ContainerFilter struct {
	UCP   sdk.Connection
	Data  dataprovider.StorageProviderOptions
	Queue queueprovider.QueueProviderOptions
}

func (f *ContainerFilter) Send(ctx context.Context, notification Notification) error {
	impacted, err := f.impactedResources(ctx, "Applications.Core/containers", notification)
	if err != nil {
		return err
	}

	for _, resource := range impacted {
		err := f.notify(ctx, resource)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *ContainerFilter) impactedResources(ctx context.Context, resourceType string, notification Notification) ([]resources.ID, error) {
	client, err := generated.NewGenericResourcesClient("/planes/radius/local", resourceType, &aztoken.AnonymousCredential{}, sdk.NewClientOptions(f.UCP))
	if err != nil {
		return nil, err
	}

	results := []resources.ID{}

	pager := client.NewListByRootScopePager(nil)
	for pager.More() {
		response, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, resource := range response.Value {
			connections := f.connections(resource)

			matched := false
			for _, id := range connections {
				if strings.EqualFold(id.String(), notification.ID.String()) {
					results = append(results, resources.MustParse(*resource.ID))
					matched = true
				}

				// Avoid duplicates if a resource mentioned an dependency resource twice.
				if matched {
					break
				}
			}
		}
	}

	return results, nil
}

func (f *ContainerFilter) connections(resource *generated.GenericResource) []resources.ID {
	// extract $properties.connnections.*.source
	obj, ok := resource.Properties["connections"]
	if !ok {
		return nil
	}

	connections, ok := obj.(map[string]any)
	if !ok {
		return nil
	}

	results := []resources.ID{}
	for _, obj := range connections {
		connection, ok := obj.(map[string]any)
		if !ok {
			continue
		}

		obj, ok = connection["source"]
		if !ok {
			continue
		}

		source, ok := obj.(string)
		if !ok {
			continue
		}

		results = append(results, resources.MustParse(source))
	}

	return results
}

func (f *ContainerFilter) notify(ctx context.Context, id resources.ID) error {
	storage, err := f.storageClient(ctx, id.Type())
	if err != nil {
		return err
	}

	obj, err := storage.Get(ctx, id.String(), nil)
	if err != nil {
		return err
	}

	ps := f.provisioningState(obj)
	if !ps.IsTerminal() {
		return fmt.Errorf("resource %s is not in a terminal state: %s", id, ps)
	}

	f.setProvisioningState(obj, v1.ProvisioningStateUpdating)

	err = storage.Save(ctx, obj, store.WithETag(obj.ETag))
	if err != nil {
		return err
	}

	options := statusmanager.QueueOperationOptions{
		OperationTimeout: time.Minute * 60, // TODO
		RetryAfter:       v1.DefaultRetryAfterDuration,
	}

	sCtx := v1.ARMRequestContext{
		ResourceID:    id,
		OperationID:   uuid.New(),
		OperationType: v1.OperationType{Type: strings.ToUpper(id.Type()), Method: "PUT"},
	}

	ctx = v1.WithARMRequestContext(ctx, &sCtx)
	err = f.statusManager(ctx).QueueAsyncOperation(ctx, &sCtx, options)

	if err != nil {
		f.setProvisioningState(obj, v1.ProvisioningStateFailed)
		saveErr := storage.Save(ctx, obj, store.WithETag(obj.ETag))
		if saveErr != nil {
			return saveErr
		}

		return err
	}

	return nil
}

func (f *ContainerFilter) provisioningState(resource any) v1.ProvisioningState {
	b, err := json.Marshal(resource)
	if err != nil {
		panic(err)
	}

	data := map[string]any{}
	err = json.Unmarshal(b, &data)
	if err != nil {
		panic(err)
	}

	obj, ok := data["properties"]
	if !ok {
		obj = map[string]any{}
		data["properties"] = obj
	}

	properties, ok := obj.(map[string]any)
	if !ok {
		return v1.ProvisioningStateSucceeded
	}

	obj, ok = properties["provisioningState"]
	if !ok {
		return v1.ProvisioningStateSucceeded
	}

	ps, ok := obj.(string)
	if !ok {
		return v1.ProvisioningStateSucceeded
	}

	return v1.ProvisioningState(ps)
}

func (f *ContainerFilter) setProvisioningState(resource any, ps v1.ProvisioningState) {
	b, err := json.Marshal(resource)
	if err != nil {
		panic(err)
	}

	data := map[string]any{}
	err = json.Unmarshal(b, &data)
	if err != nil {
		panic(err)
	}

	obj, ok := data["properties"]
	if !ok {
		obj = map[string]any{}
		data["properties"] = obj
	}

	properties, ok := obj.(map[string]any)
	if !ok {
		return
	}

	properties["provisioningState"] = string(ps)
}

func (f *ContainerFilter) storageClient(ctx context.Context, resourceType string) (store.StorageClient, error) {
	return dataprovider.NewStorageProvider(f.Data).GetStorageClient(ctx, resourceType)
}

func (f *ContainerFilter) queueClient(ctx context.Context) (queue.Client, error) {
	return queueprovider.New(f.Queue).GetClient(ctx)
}

func (f *ContainerFilter) statusManager(ctx context.Context) statusmanager.StatusManager {
	queueClient, err := f.queueClient(ctx)
	if err != nil {
		return nil
	}

	return statusmanager.New(dataprovider.NewStorageProvider(f.Data), queueClient, "global")
}

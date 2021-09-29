// +build go1.13

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.
// Code generated by Microsoft (R) AutoRest Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

package radclientv3

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/armcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RabbitmqComMessageQueueClient contains the methods for the RabbitmqComMessageQueue group.
// Don't use this type directly, use NewRabbitmqComMessageQueueClient() instead.
type RabbitmqComMessageQueueClient struct {
	con *armcore.Connection
	subscriptionID string
}

// NewRabbitmqComMessageQueueClient creates a new instance of RabbitmqComMessageQueueClient with the specified values.
func NewRabbitmqComMessageQueueClient(con *armcore.Connection, subscriptionID string) *RabbitmqComMessageQueueClient {
	return &RabbitmqComMessageQueueClient{con: con, subscriptionID: subscriptionID}
}

// BeginCreateOrUpdate - Creates or updates a rabbitmq.com.MessageQueue resource.
// If the operation fails it returns the *ErrorResponse error type.
func (client *RabbitmqComMessageQueueClient) BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, applicationName string, rabbitMQComponentName string, parameters RabbitMQComponentResource, options *RabbitmqComMessageQueueBeginCreateOrUpdateOptions) (RabbitMQComponentResourcePollerResponse, error) {
	resp, err := client.createOrUpdate(ctx, resourceGroupName, applicationName, rabbitMQComponentName, parameters, options)
	if err != nil {
		return RabbitMQComponentResourcePollerResponse{}, err
	}
	result := RabbitMQComponentResourcePollerResponse{
		RawResponse: resp.Response,
	}
	pt, err := armcore.NewLROPoller("RabbitmqComMessageQueueClient.CreateOrUpdate", "location", resp, client.con.Pipeline(), client.createOrUpdateHandleError)
	if err != nil {
		return RabbitMQComponentResourcePollerResponse{}, err
	}
	poller := &rabbitMQComponentResourcePoller{
		pt: pt,
	}
	result.Poller = poller
	result.PollUntilDone = func(ctx context.Context, frequency time.Duration) (RabbitMQComponentResourceResponse, error) {
		return poller.pollUntilDone(ctx, frequency)
	}
	return result, nil
}

// ResumeCreateOrUpdate creates a new RabbitMQComponentResourcePoller from the specified resume token.
// token - The value must come from a previous call to RabbitMQComponentResourcePoller.ResumeToken().
func (client *RabbitmqComMessageQueueClient) ResumeCreateOrUpdate(ctx context.Context, token string) (RabbitMQComponentResourcePollerResponse, error) {
	pt, err := armcore.NewLROPollerFromResumeToken("RabbitmqComMessageQueueClient.CreateOrUpdate", token, client.con.Pipeline(), client.createOrUpdateHandleError)
	if err != nil {
		return RabbitMQComponentResourcePollerResponse{}, err
	}
	poller := &rabbitMQComponentResourcePoller{
		pt: pt,
	}
	resp, err := poller.Poll(ctx)
	if err != nil {
		return RabbitMQComponentResourcePollerResponse{}, err
	}
	result := RabbitMQComponentResourcePollerResponse{
		RawResponse: resp,
	}
	result.Poller = poller
	result.PollUntilDone = func(ctx context.Context, frequency time.Duration) (RabbitMQComponentResourceResponse, error) {
		return poller.pollUntilDone(ctx, frequency)
	}
	return result, nil
}

// CreateOrUpdate - Creates or updates a rabbitmq.com.MessageQueue resource.
// If the operation fails it returns the *ErrorResponse error type.
func (client *RabbitmqComMessageQueueClient) createOrUpdate(ctx context.Context, resourceGroupName string, applicationName string, rabbitMQComponentName string, parameters RabbitMQComponentResource, options *RabbitmqComMessageQueueBeginCreateOrUpdateOptions) (*azcore.Response, error) {
	req, err := client.createOrUpdateCreateRequest(ctx, resourceGroupName, applicationName, rabbitMQComponentName, parameters, options)
	if err != nil {
		return nil, err
	}
	resp, err := client.con.Pipeline().Do(req)
	if err != nil {
		return nil, err
	}
	if !resp.HasStatusCode(http.StatusOK, http.StatusCreated, http.StatusAccepted) {
		return nil, client.createOrUpdateHandleError(resp)
	}
	 return resp, nil
}

// createOrUpdateCreateRequest creates the CreateOrUpdate request.
func (client *RabbitmqComMessageQueueClient) createOrUpdateCreateRequest(ctx context.Context, resourceGroupName string, applicationName string, rabbitMQComponentName string, parameters RabbitMQComponentResource, options *RabbitmqComMessageQueueBeginCreateOrUpdateOptions) (*azcore.Request, error) {
	urlPath := "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.CustomProviders/resourceProviders/radiusv3/Application/{applicationName}/rabbitmq.com.MessageQueue/{rabbitMQComponentName}"
	if client.subscriptionID == "" {
		return nil, errors.New("parameter client.subscriptionID cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionId}", url.PathEscape(client.subscriptionID))
	if resourceGroupName == "" {
		return nil, errors.New("parameter resourceGroupName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{resourceGroupName}", url.PathEscape(resourceGroupName))
	if applicationName == "" {
		return nil, errors.New("parameter applicationName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{applicationName}", url.PathEscape(applicationName))
	if rabbitMQComponentName == "" {
		return nil, errors.New("parameter rabbitMQComponentName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{rabbitMQComponentName}", url.PathEscape(rabbitMQComponentName))
	req, err := azcore.NewRequest(ctx, http.MethodPut, azcore.JoinPaths(client.con.Endpoint(), urlPath))
	if err != nil {
		return nil, err
	}
	req.Telemetry(telemetryInfo)
	reqQP := req.URL.Query()
	reqQP.Set("api-version", "2018-09-01-preview")
	req.URL.RawQuery = reqQP.Encode()
	req.Header.Set("Accept", "application/json")
	return req, req.MarshalAsJSON(parameters)
}

// createOrUpdateHandleError handles the CreateOrUpdate error response.
func (client *RabbitmqComMessageQueueClient) createOrUpdateHandleError(resp *azcore.Response) error {
	body, err := resp.Payload()
	if err != nil {
		return azcore.NewResponseError(err, resp.Response)
	}
		errType := ErrorResponse{raw: string(body)}
	if err := resp.UnmarshalAsJSON(&errType); err != nil {
		return azcore.NewResponseError(fmt.Errorf("%s\n%s", string(body), err), resp.Response)
	}
	return azcore.NewResponseError(&errType, resp.Response)
}

// BeginDelete - Deletes a rabbitmq.com.MessageQueue resource.
// If the operation fails it returns the *ErrorResponse error type.
func (client *RabbitmqComMessageQueueClient) BeginDelete(ctx context.Context, resourceGroupName string, applicationName string, rabbitMQComponentName string, options *RabbitmqComMessageQueueBeginDeleteOptions) (HTTPPollerResponse, error) {
	resp, err := client.deleteOperation(ctx, resourceGroupName, applicationName, rabbitMQComponentName, options)
	if err != nil {
		return HTTPPollerResponse{}, err
	}
	result := HTTPPollerResponse{
		RawResponse: resp.Response,
	}
	pt, err := armcore.NewLROPoller("RabbitmqComMessageQueueClient.Delete", "location", resp, client.con.Pipeline(), client.deleteHandleError)
	if err != nil {
		return HTTPPollerResponse{}, err
	}
	poller := &httpPoller{
		pt: pt,
	}
	result.Poller = poller
	result.PollUntilDone = func(ctx context.Context, frequency time.Duration) (*http.Response, error) {
		return poller.pollUntilDone(ctx, frequency)
	}
	return result, nil
}

// ResumeDelete creates a new HTTPPoller from the specified resume token.
// token - The value must come from a previous call to HTTPPoller.ResumeToken().
func (client *RabbitmqComMessageQueueClient) ResumeDelete(ctx context.Context, token string) (HTTPPollerResponse, error) {
	pt, err := armcore.NewLROPollerFromResumeToken("RabbitmqComMessageQueueClient.Delete", token, client.con.Pipeline(), client.deleteHandleError)
	if err != nil {
		return HTTPPollerResponse{}, err
	}
	poller := &httpPoller{
		pt: pt,
	}
	resp, err := poller.Poll(ctx)
	if err != nil {
		return HTTPPollerResponse{}, err
	}
	result := HTTPPollerResponse{
		RawResponse: resp,
	}
	result.Poller = poller
	result.PollUntilDone = func(ctx context.Context, frequency time.Duration) (*http.Response, error) {
		return poller.pollUntilDone(ctx, frequency)
	}
	return result, nil
}

// Delete - Deletes a rabbitmq.com.MessageQueue resource.
// If the operation fails it returns the *ErrorResponse error type.
func (client *RabbitmqComMessageQueueClient) deleteOperation(ctx context.Context, resourceGroupName string, applicationName string, rabbitMQComponentName string, options *RabbitmqComMessageQueueBeginDeleteOptions) (*azcore.Response, error) {
	req, err := client.deleteCreateRequest(ctx, resourceGroupName, applicationName, rabbitMQComponentName, options)
	if err != nil {
		return nil, err
	}
	resp, err := client.con.Pipeline().Do(req)
	if err != nil {
		return nil, err
	}
	if !resp.HasStatusCode(http.StatusAccepted, http.StatusNoContent) {
		return nil, client.deleteHandleError(resp)
	}
	 return resp, nil
}

// deleteCreateRequest creates the Delete request.
func (client *RabbitmqComMessageQueueClient) deleteCreateRequest(ctx context.Context, resourceGroupName string, applicationName string, rabbitMQComponentName string, options *RabbitmqComMessageQueueBeginDeleteOptions) (*azcore.Request, error) {
	urlPath := "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.CustomProviders/resourceProviders/radiusv3/Application/{applicationName}/rabbitmq.com.MessageQueue/{rabbitMQComponentName}"
	if client.subscriptionID == "" {
		return nil, errors.New("parameter client.subscriptionID cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionId}", url.PathEscape(client.subscriptionID))
	if resourceGroupName == "" {
		return nil, errors.New("parameter resourceGroupName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{resourceGroupName}", url.PathEscape(resourceGroupName))
	if applicationName == "" {
		return nil, errors.New("parameter applicationName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{applicationName}", url.PathEscape(applicationName))
	if rabbitMQComponentName == "" {
		return nil, errors.New("parameter rabbitMQComponentName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{rabbitMQComponentName}", url.PathEscape(rabbitMQComponentName))
	req, err := azcore.NewRequest(ctx, http.MethodDelete, azcore.JoinPaths(client.con.Endpoint(), urlPath))
	if err != nil {
		return nil, err
	}
	req.Telemetry(telemetryInfo)
	reqQP := req.URL.Query()
	reqQP.Set("api-version", "2018-09-01-preview")
	req.URL.RawQuery = reqQP.Encode()
	req.Header.Set("Accept", "application/json")
	return req, nil
}

// deleteHandleError handles the Delete error response.
func (client *RabbitmqComMessageQueueClient) deleteHandleError(resp *azcore.Response) error {
	body, err := resp.Payload()
	if err != nil {
		return azcore.NewResponseError(err, resp.Response)
	}
		errType := ErrorResponse{raw: string(body)}
	if err := resp.UnmarshalAsJSON(&errType); err != nil {
		return azcore.NewResponseError(fmt.Errorf("%s\n%s", string(body), err), resp.Response)
	}
	return azcore.NewResponseError(&errType, resp.Response)
}

// Get - Gets a rabbitmq.com.MessageQueue resource by name.
// If the operation fails it returns the *ErrorResponse error type.
func (client *RabbitmqComMessageQueueClient) Get(ctx context.Context, resourceGroupName string, applicationName string, rabbitMQComponentName string, options *RabbitmqComMessageQueueGetOptions) (RabbitMQComponentResourceResponse, error) {
	req, err := client.getCreateRequest(ctx, resourceGroupName, applicationName, rabbitMQComponentName, options)
	if err != nil {
		return RabbitMQComponentResourceResponse{}, err
	}
	resp, err := client.con.Pipeline().Do(req)
	if err != nil {
		return RabbitMQComponentResourceResponse{}, err
	}
	if !resp.HasStatusCode(http.StatusOK) {
		return RabbitMQComponentResourceResponse{}, client.getHandleError(resp)
	}
	return client.getHandleResponse(resp)
}

// getCreateRequest creates the Get request.
func (client *RabbitmqComMessageQueueClient) getCreateRequest(ctx context.Context, resourceGroupName string, applicationName string, rabbitMQComponentName string, options *RabbitmqComMessageQueueGetOptions) (*azcore.Request, error) {
	urlPath := "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.CustomProviders/resourceProviders/radiusv3/Application/{applicationName}/rabbitmq.com.MessageQueue/{rabbitMQComponentName}"
	if client.subscriptionID == "" {
		return nil, errors.New("parameter client.subscriptionID cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionId}", url.PathEscape(client.subscriptionID))
	if resourceGroupName == "" {
		return nil, errors.New("parameter resourceGroupName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{resourceGroupName}", url.PathEscape(resourceGroupName))
	if applicationName == "" {
		return nil, errors.New("parameter applicationName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{applicationName}", url.PathEscape(applicationName))
	if rabbitMQComponentName == "" {
		return nil, errors.New("parameter rabbitMQComponentName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{rabbitMQComponentName}", url.PathEscape(rabbitMQComponentName))
	req, err := azcore.NewRequest(ctx, http.MethodGet, azcore.JoinPaths(client.con.Endpoint(), urlPath))
	if err != nil {
		return nil, err
	}
	req.Telemetry(telemetryInfo)
	reqQP := req.URL.Query()
	reqQP.Set("api-version", "2018-09-01-preview")
	req.URL.RawQuery = reqQP.Encode()
	req.Header.Set("Accept", "application/json")
	return req, nil
}

// getHandleResponse handles the Get response.
func (client *RabbitmqComMessageQueueClient) getHandleResponse(resp *azcore.Response) (RabbitMQComponentResourceResponse, error) {
	var val *RabbitMQComponentResource
	if err := resp.UnmarshalAsJSON(&val); err != nil {
		return RabbitMQComponentResourceResponse{}, err
	}
return RabbitMQComponentResourceResponse{RawResponse: resp.Response, RabbitMQComponentResource: val}, nil
}

// getHandleError handles the Get error response.
func (client *RabbitmqComMessageQueueClient) getHandleError(resp *azcore.Response) error {
	body, err := resp.Payload()
	if err != nil {
		return azcore.NewResponseError(err, resp.Response)
	}
		errType := ErrorResponse{raw: string(body)}
	if err := resp.UnmarshalAsJSON(&errType); err != nil {
		return azcore.NewResponseError(fmt.Errorf("%s\n%s", string(body), err), resp.Response)
	}
	return azcore.NewResponseError(&errType, resp.Response)
}

// List - List the rabbitmq.com.MessageQueue resources deployed in the application.
// If the operation fails it returns the *ErrorResponse error type.
func (client *RabbitmqComMessageQueueClient) List(ctx context.Context, resourceGroupName string, applicationName string, options *RabbitmqComMessageQueueListOptions) (RabbitMQComponentListResponse, error) {
	req, err := client.listCreateRequest(ctx, resourceGroupName, applicationName, options)
	if err != nil {
		return RabbitMQComponentListResponse{}, err
	}
	resp, err := client.con.Pipeline().Do(req)
	if err != nil {
		return RabbitMQComponentListResponse{}, err
	}
	if !resp.HasStatusCode(http.StatusOK) {
		return RabbitMQComponentListResponse{}, client.listHandleError(resp)
	}
	return client.listHandleResponse(resp)
}

// listCreateRequest creates the List request.
func (client *RabbitmqComMessageQueueClient) listCreateRequest(ctx context.Context, resourceGroupName string, applicationName string, options *RabbitmqComMessageQueueListOptions) (*azcore.Request, error) {
	urlPath := "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.CustomProviders/resourceProviders/radiusv3/Application/{applicationName}/rabbitmq.com.MessageQueue"
	if client.subscriptionID == "" {
		return nil, errors.New("parameter client.subscriptionID cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionId}", url.PathEscape(client.subscriptionID))
	if resourceGroupName == "" {
		return nil, errors.New("parameter resourceGroupName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{resourceGroupName}", url.PathEscape(resourceGroupName))
	if applicationName == "" {
		return nil, errors.New("parameter applicationName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{applicationName}", url.PathEscape(applicationName))
	req, err := azcore.NewRequest(ctx, http.MethodGet, azcore.JoinPaths(client.con.Endpoint(), urlPath))
	if err != nil {
		return nil, err
	}
	req.Telemetry(telemetryInfo)
	reqQP := req.URL.Query()
	reqQP.Set("api-version", "2018-09-01-preview")
	req.URL.RawQuery = reqQP.Encode()
	req.Header.Set("Accept", "application/json")
	return req, nil
}

// listHandleResponse handles the List response.
func (client *RabbitmqComMessageQueueClient) listHandleResponse(resp *azcore.Response) (RabbitMQComponentListResponse, error) {
	var val *RabbitMQComponentList
	if err := resp.UnmarshalAsJSON(&val); err != nil {
		return RabbitMQComponentListResponse{}, err
	}
return RabbitMQComponentListResponse{RawResponse: resp.Response, RabbitMQComponentList: val}, nil
}

// listHandleError handles the List error response.
func (client *RabbitmqComMessageQueueClient) listHandleError(resp *azcore.Response) error {
	body, err := resp.Payload()
	if err != nil {
		return azcore.NewResponseError(err, resp.Response)
	}
		errType := ErrorResponse{raw: string(body)}
	if err := resp.UnmarshalAsJSON(&errType); err != nil {
		return azcore.NewResponseError(fmt.Errorf("%s\n%s", string(body), err), resp.Response)
	}
	return azcore.NewResponseError(&errType, resp.Response)
}


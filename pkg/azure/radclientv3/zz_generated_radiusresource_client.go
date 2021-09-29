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

// RadiusResourceClient contains the methods for the RadiusResource group.
// Don't use this type directly, use NewRadiusResourceClient() instead.
type RadiusResourceClient struct {
	con *armcore.Connection
	subscriptionID string
}

// NewRadiusResourceClient creates a new instance of RadiusResourceClient with the specified values.
func NewRadiusResourceClient(con *armcore.Connection, subscriptionID string) *RadiusResourceClient {
	return &RadiusResourceClient{con: con, subscriptionID: subscriptionID}
}

// BeginDelete - Deletes a RadiusResource resource.
// If the operation fails it returns the *ErrorResponse error type.
func (client *RadiusResourceClient) BeginDelete(ctx context.Context, resourceGroupName string, applicationName string, radiusResourceType string, radiusResourceName string, options *RadiusResourceBeginDeleteOptions) (HTTPPollerResponse, error) {
	resp, err := client.deleteOperation(ctx, resourceGroupName, applicationName, radiusResourceType, radiusResourceName, options)
	if err != nil {
		return HTTPPollerResponse{}, err
	}
	result := HTTPPollerResponse{
		RawResponse: resp.Response,
	}
	pt, err := armcore.NewLROPoller("RadiusResourceClient.Delete", "location", resp, client.con.Pipeline(), client.deleteHandleError)
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
func (client *RadiusResourceClient) ResumeDelete(ctx context.Context, token string) (HTTPPollerResponse, error) {
	pt, err := armcore.NewLROPollerFromResumeToken("RadiusResourceClient.Delete", token, client.con.Pipeline(), client.deleteHandleError)
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

// Delete - Deletes a RadiusResource resource.
// If the operation fails it returns the *ErrorResponse error type.
func (client *RadiusResourceClient) deleteOperation(ctx context.Context, resourceGroupName string, applicationName string, radiusResourceType string, radiusResourceName string, options *RadiusResourceBeginDeleteOptions) (*azcore.Response, error) {
	req, err := client.deleteCreateRequest(ctx, resourceGroupName, applicationName, radiusResourceType, radiusResourceName, options)
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
func (client *RadiusResourceClient) deleteCreateRequest(ctx context.Context, resourceGroupName string, applicationName string, radiusResourceType string, radiusResourceName string, options *RadiusResourceBeginDeleteOptions) (*azcore.Request, error) {
	urlPath := "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.CustomProviders/resourceProviders/radiusv3/Application/{applicationName}/{radiusResourceType}/{radiusResourceName}"
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
	if radiusResourceType == "" {
		return nil, errors.New("parameter radiusResourceType cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{radiusResourceType}", url.PathEscape(radiusResourceType))
	if radiusResourceName == "" {
		return nil, errors.New("parameter radiusResourceName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{radiusResourceName}", url.PathEscape(radiusResourceName))
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
func (client *RadiusResourceClient) deleteHandleError(resp *azcore.Response) error {
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

// Get - Gets a RadiusResource resource by name.
// If the operation fails it returns the *ErrorResponse error type.
func (client *RadiusResourceClient) Get(ctx context.Context, resourceGroupName string, applicationName string, radiusResourceType string, radiusResourceName string, options *RadiusResourceGetOptions) (RadiusResourceResponse, error) {
	req, err := client.getCreateRequest(ctx, resourceGroupName, applicationName, radiusResourceType, radiusResourceName, options)
	if err != nil {
		return RadiusResourceResponse{}, err
	}
	resp, err := client.con.Pipeline().Do(req)
	if err != nil {
		return RadiusResourceResponse{}, err
	}
	if !resp.HasStatusCode(http.StatusOK) {
		return RadiusResourceResponse{}, client.getHandleError(resp)
	}
	return client.getHandleResponse(resp)
}

// getCreateRequest creates the Get request.
func (client *RadiusResourceClient) getCreateRequest(ctx context.Context, resourceGroupName string, applicationName string, radiusResourceType string, radiusResourceName string, options *RadiusResourceGetOptions) (*azcore.Request, error) {
	urlPath := "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.CustomProviders/resourceProviders/radiusv3/Application/{applicationName}/{radiusResourceType}/{radiusResourceName}"
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
	if radiusResourceType == "" {
		return nil, errors.New("parameter radiusResourceType cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{radiusResourceType}", url.PathEscape(radiusResourceType))
	if radiusResourceName == "" {
		return nil, errors.New("parameter radiusResourceName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{radiusResourceName}", url.PathEscape(radiusResourceName))
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
func (client *RadiusResourceClient) getHandleResponse(resp *azcore.Response) (RadiusResourceResponse, error) {
	var val *RadiusResource
	if err := resp.UnmarshalAsJSON(&val); err != nil {
		return RadiusResourceResponse{}, err
	}
return RadiusResourceResponse{RawResponse: resp.Response, RadiusResource: val}, nil
}

// getHandleError handles the Get error response.
func (client *RadiusResourceClient) getHandleError(resp *azcore.Response) error {
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

// List - List the RadiusResource resources deployed in the application.
// If the operation fails it returns the *ErrorResponse error type.
func (client *RadiusResourceClient) List(ctx context.Context, resourceGroupName string, applicationName string, options *RadiusResourceListOptions) (RadiusResourceListResponse, error) {
	req, err := client.listCreateRequest(ctx, resourceGroupName, applicationName, options)
	if err != nil {
		return RadiusResourceListResponse{}, err
	}
	resp, err := client.con.Pipeline().Do(req)
	if err != nil {
		return RadiusResourceListResponse{}, err
	}
	if !resp.HasStatusCode(http.StatusOK) {
		return RadiusResourceListResponse{}, client.listHandleError(resp)
	}
	return client.listHandleResponse(resp)
}

// listCreateRequest creates the List request.
func (client *RadiusResourceClient) listCreateRequest(ctx context.Context, resourceGroupName string, applicationName string, options *RadiusResourceListOptions) (*azcore.Request, error) {
	urlPath := "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.CustomProviders/resourceProviders/radiusv3/Application/{applicationName}/RadiusResource"
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
func (client *RadiusResourceClient) listHandleResponse(resp *azcore.Response) (RadiusResourceListResponse, error) {
	var val *RadiusResourceList
	if err := resp.UnmarshalAsJSON(&val); err != nil {
		return RadiusResourceListResponse{}, err
	}
return RadiusResourceListResponse{RawResponse: resp.Response, RadiusResourceList: val}, nil
}

// listHandleError handles the List error response.
func (client *RadiusResourceClient) listHandleError(resp *azcore.Response) error {
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


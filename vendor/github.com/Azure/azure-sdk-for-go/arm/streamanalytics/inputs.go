package streamanalytics

// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Code generated by Microsoft (R) AutoRest Code Generator 1.1.0.0
// Changes may cause incorrect behavior and will be lost if the code is
// regenerated.

import (
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"net/http"
)

// InputsClient is the composite Swagger for Stream Analytics Client
type InputsClient struct {
	ManagementClient
}

// NewInputsClient creates an instance of the InputsClient client.
func NewInputsClient(subscriptionID string) InputsClient {
	return NewInputsClientWithBaseURI(DefaultBaseURI, subscriptionID)
}

// NewInputsClientWithBaseURI creates an instance of the InputsClient client.
func NewInputsClientWithBaseURI(baseURI string, subscriptionID string) InputsClient {
	return InputsClient{NewWithBaseURI(baseURI, subscriptionID)}
}

// CreateOrReplace creates an input or replaces an already existing input under
// an existing streaming job.
//
// input is the definition of the input that will be used to create a new input
// or replace the existing one under the streaming job. resourceGroupName is
// the name of the resource group that contains the resource. You can obtain
// this value from the Azure Resource Manager API or the portal. jobName is the
// name of the streaming job. inputName is the name of the input. ifMatch is
// the ETag of the input. Omit this value to always overwrite the current
// input. Specify the last-seen ETag value to prevent accidentally overwritting
// concurrent changes. ifNoneMatch is set to '*' to allow a new input to be
// created, but to prevent updating an existing input. Other values will result
// in a 412 Pre-condition Failed response.
func (client InputsClient) CreateOrReplace(input Input, resourceGroupName string, jobName string, inputName string, ifMatch string, ifNoneMatch string) (result Input, err error) {
	req, err := client.CreateOrReplacePreparer(input, resourceGroupName, jobName, inputName, ifMatch, ifNoneMatch)
	if err != nil {
		err = autorest.NewErrorWithError(err, "streamanalytics.InputsClient", "CreateOrReplace", nil, "Failure preparing request")
		return
	}

	resp, err := client.CreateOrReplaceSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		err = autorest.NewErrorWithError(err, "streamanalytics.InputsClient", "CreateOrReplace", resp, "Failure sending request")
		return
	}

	result, err = client.CreateOrReplaceResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "streamanalytics.InputsClient", "CreateOrReplace", resp, "Failure responding to request")
	}

	return
}

// CreateOrReplacePreparer prepares the CreateOrReplace request.
func (client InputsClient) CreateOrReplacePreparer(input Input, resourceGroupName string, jobName string, inputName string, ifMatch string, ifNoneMatch string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"inputName":         autorest.Encode("path", inputName),
		"jobName":           autorest.Encode("path", jobName),
		"resourceGroupName": autorest.Encode("path", resourceGroupName),
		"subscriptionId":    autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2016-03-01"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsJSON(),
		autorest.AsPut(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}/providers/Microsoft.StreamAnalytics/streamingjobs/{jobName}/inputs/{inputName}", pathParameters),
		autorest.WithJSON(input),
		autorest.WithQueryParameters(queryParameters))
	if len(ifMatch) > 0 {
		preparer = autorest.DecoratePreparer(preparer,
			autorest.WithHeader("If-Match", autorest.String(ifMatch)))
	}
	if len(ifNoneMatch) > 0 {
		preparer = autorest.DecoratePreparer(preparer,
			autorest.WithHeader("If-None-Match", autorest.String(ifNoneMatch)))
	}
	return preparer.Prepare(&http.Request{})
}

// CreateOrReplaceSender sends the CreateOrReplace request. The method will close the
// http.Response Body if it receives an error.
func (client InputsClient) CreateOrReplaceSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client, req)
}

// CreateOrReplaceResponder handles the response to the CreateOrReplace request. The method always
// closes the http.Response Body.
func (client InputsClient) CreateOrReplaceResponder(resp *http.Response) (result Input, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// Delete deletes an input from the streaming job.
//
// resourceGroupName is the name of the resource group that contains the
// resource. You can obtain this value from the Azure Resource Manager API or
// the portal. jobName is the name of the streaming job. inputName is the name
// of the input.
func (client InputsClient) Delete(resourceGroupName string, jobName string, inputName string) (result autorest.Response, err error) {
	req, err := client.DeletePreparer(resourceGroupName, jobName, inputName)
	if err != nil {
		err = autorest.NewErrorWithError(err, "streamanalytics.InputsClient", "Delete", nil, "Failure preparing request")
		return
	}

	resp, err := client.DeleteSender(req)
	if err != nil {
		result.Response = resp
		err = autorest.NewErrorWithError(err, "streamanalytics.InputsClient", "Delete", resp, "Failure sending request")
		return
	}

	result, err = client.DeleteResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "streamanalytics.InputsClient", "Delete", resp, "Failure responding to request")
	}

	return
}

// DeletePreparer prepares the Delete request.
func (client InputsClient) DeletePreparer(resourceGroupName string, jobName string, inputName string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"inputName":         autorest.Encode("path", inputName),
		"jobName":           autorest.Encode("path", jobName),
		"resourceGroupName": autorest.Encode("path", resourceGroupName),
		"subscriptionId":    autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2016-03-01"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsDelete(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}/providers/Microsoft.StreamAnalytics/streamingjobs/{jobName}/inputs/{inputName}", pathParameters),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare(&http.Request{})
}

// DeleteSender sends the Delete request. The method will close the
// http.Response Body if it receives an error.
func (client InputsClient) DeleteSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client, req)
}

// DeleteResponder handles the response to the Delete request. The method always
// closes the http.Response Body.
func (client InputsClient) DeleteResponder(resp *http.Response) (result autorest.Response, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusNoContent),
		autorest.ByClosing())
	result.Response = resp
	return
}

// Get gets details about the specified input.
//
// resourceGroupName is the name of the resource group that contains the
// resource. You can obtain this value from the Azure Resource Manager API or
// the portal. jobName is the name of the streaming job. inputName is the name
// of the input.
func (client InputsClient) Get(resourceGroupName string, jobName string, inputName string) (result Input, err error) {
	req, err := client.GetPreparer(resourceGroupName, jobName, inputName)
	if err != nil {
		err = autorest.NewErrorWithError(err, "streamanalytics.InputsClient", "Get", nil, "Failure preparing request")
		return
	}

	resp, err := client.GetSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		err = autorest.NewErrorWithError(err, "streamanalytics.InputsClient", "Get", resp, "Failure sending request")
		return
	}

	result, err = client.GetResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "streamanalytics.InputsClient", "Get", resp, "Failure responding to request")
	}

	return
}

// GetPreparer prepares the Get request.
func (client InputsClient) GetPreparer(resourceGroupName string, jobName string, inputName string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"inputName":         autorest.Encode("path", inputName),
		"jobName":           autorest.Encode("path", jobName),
		"resourceGroupName": autorest.Encode("path", resourceGroupName),
		"subscriptionId":    autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2016-03-01"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsGet(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}/providers/Microsoft.StreamAnalytics/streamingjobs/{jobName}/inputs/{inputName}", pathParameters),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare(&http.Request{})
}

// GetSender sends the Get request. The method will close the
// http.Response Body if it receives an error.
func (client InputsClient) GetSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client, req)
}

// GetResponder handles the response to the Get request. The method always
// closes the http.Response Body.
func (client InputsClient) GetResponder(resp *http.Response) (result Input, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// ListByStreamingJob lists all of the inputs under the specified streaming
// job.
//
// resourceGroupName is the name of the resource group that contains the
// resource. You can obtain this value from the Azure Resource Manager API or
// the portal. jobName is the name of the streaming job. selectParameter is the
// $select OData query parameter. This is a comma-separated list of structural
// properties to include in the response, or “*” to include all properties. By
// default, all properties are returned except diagnostics. Currently only
// accepts '*' as a valid value.
func (client InputsClient) ListByStreamingJob(resourceGroupName string, jobName string, selectParameter string) (result InputListResult, err error) {
	req, err := client.ListByStreamingJobPreparer(resourceGroupName, jobName, selectParameter)
	if err != nil {
		err = autorest.NewErrorWithError(err, "streamanalytics.InputsClient", "ListByStreamingJob", nil, "Failure preparing request")
		return
	}

	resp, err := client.ListByStreamingJobSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		err = autorest.NewErrorWithError(err, "streamanalytics.InputsClient", "ListByStreamingJob", resp, "Failure sending request")
		return
	}

	result, err = client.ListByStreamingJobResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "streamanalytics.InputsClient", "ListByStreamingJob", resp, "Failure responding to request")
	}

	return
}

// ListByStreamingJobPreparer prepares the ListByStreamingJob request.
func (client InputsClient) ListByStreamingJobPreparer(resourceGroupName string, jobName string, selectParameter string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"jobName":           autorest.Encode("path", jobName),
		"resourceGroupName": autorest.Encode("path", resourceGroupName),
		"subscriptionId":    autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2016-03-01"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}
	if len(selectParameter) > 0 {
		queryParameters["$select"] = autorest.Encode("query", selectParameter)
	}

	preparer := autorest.CreatePreparer(
		autorest.AsGet(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}/providers/Microsoft.StreamAnalytics/streamingjobs/{jobName}/inputs", pathParameters),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare(&http.Request{})
}

// ListByStreamingJobSender sends the ListByStreamingJob request. The method will close the
// http.Response Body if it receives an error.
func (client InputsClient) ListByStreamingJobSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client, req)
}

// ListByStreamingJobResponder handles the response to the ListByStreamingJob request. The method always
// closes the http.Response Body.
func (client InputsClient) ListByStreamingJobResponder(resp *http.Response) (result InputListResult, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// ListByStreamingJobNextResults retrieves the next set of results, if any.
func (client InputsClient) ListByStreamingJobNextResults(lastResults InputListResult) (result InputListResult, err error) {
	req, err := lastResults.InputListResultPreparer()
	if err != nil {
		return result, autorest.NewErrorWithError(err, "streamanalytics.InputsClient", "ListByStreamingJob", nil, "Failure preparing next results request")
	}
	if req == nil {
		return
	}

	resp, err := client.ListByStreamingJobSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "streamanalytics.InputsClient", "ListByStreamingJob", resp, "Failure sending next results request")
	}

	result, err = client.ListByStreamingJobResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "streamanalytics.InputsClient", "ListByStreamingJob", resp, "Failure responding to next results request")
	}

	return
}

// ListByStreamingJobComplete gets all elements from the list without paging.
func (client InputsClient) ListByStreamingJobComplete(resourceGroupName string, jobName string, selectParameter string, cancel <-chan struct{}) (<-chan Input, <-chan error) {
	resultChan := make(chan Input)
	errChan := make(chan error, 1)
	go func() {
		defer func() {
			close(resultChan)
			close(errChan)
		}()
		list, err := client.ListByStreamingJob(resourceGroupName, jobName, selectParameter)
		if err != nil {
			errChan <- err
			return
		}
		if list.Value != nil {
			for _, item := range *list.Value {
				select {
				case <-cancel:
					return
				case resultChan <- item:
					// Intentionally left blank
				}
			}
		}
		for list.NextLink != nil {
			list, err = client.ListByStreamingJobNextResults(list)
			if err != nil {
				errChan <- err
				return
			}
			if list.Value != nil {
				for _, item := range *list.Value {
					select {
					case <-cancel:
						return
					case resultChan <- item:
						// Intentionally left blank
					}
				}
			}
		}
	}()
	return resultChan, errChan
}

// Test tests whether an input’s datasource is reachable and usable by the
// Azure Stream Analytics service. This method may poll for completion. Polling
// can be canceled by passing the cancel channel argument. The channel will be
// used to cancel polling and any outstanding HTTP requests.
//
// resourceGroupName is the name of the resource group that contains the
// resource. You can obtain this value from the Azure Resource Manager API or
// the portal. jobName is the name of the streaming job. inputName is the name
// of the input. input is if the input specified does not already exist, this
// parameter must contain the full input definition intended to be tested. If
// the input specified already exists, this parameter can be left null to test
// the existing input as is or if specified, the properties specified will
// overwrite the corresponding properties in the existing input (exactly like a
// PATCH operation) and the resulting input will be tested.
func (client InputsClient) Test(resourceGroupName string, jobName string, inputName string, input *Input, cancel <-chan struct{}) (<-chan ResourceTestStatus, <-chan error) {
	resultChan := make(chan ResourceTestStatus, 1)
	errChan := make(chan error, 1)
	go func() {
		var err error
		var result ResourceTestStatus
		defer func() {
			if err != nil {
				errChan <- err
			}
			resultChan <- result
			close(resultChan)
			close(errChan)
		}()
		req, err := client.TestPreparer(resourceGroupName, jobName, inputName, input, cancel)
		if err != nil {
			err = autorest.NewErrorWithError(err, "streamanalytics.InputsClient", "Test", nil, "Failure preparing request")
			return
		}

		resp, err := client.TestSender(req)
		if err != nil {
			result.Response = autorest.Response{Response: resp}
			err = autorest.NewErrorWithError(err, "streamanalytics.InputsClient", "Test", resp, "Failure sending request")
			return
		}

		result, err = client.TestResponder(resp)
		if err != nil {
			err = autorest.NewErrorWithError(err, "streamanalytics.InputsClient", "Test", resp, "Failure responding to request")
		}
	}()
	return resultChan, errChan
}

// TestPreparer prepares the Test request.
func (client InputsClient) TestPreparer(resourceGroupName string, jobName string, inputName string, input *Input, cancel <-chan struct{}) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"inputName":         autorest.Encode("path", inputName),
		"jobName":           autorest.Encode("path", jobName),
		"resourceGroupName": autorest.Encode("path", resourceGroupName),
		"subscriptionId":    autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2016-03-01"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsJSON(),
		autorest.AsPost(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}/providers/Microsoft.StreamAnalytics/streamingjobs/{jobName}/inputs/{inputName}/test", pathParameters),
		autorest.WithQueryParameters(queryParameters))
	if input != nil {
		preparer = autorest.DecoratePreparer(preparer,
			autorest.WithJSON(input))
	}
	return preparer.Prepare(&http.Request{Cancel: cancel})
}

// TestSender sends the Test request. The method will close the
// http.Response Body if it receives an error.
func (client InputsClient) TestSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client,
		req,
		azure.DoPollForAsynchronous(client.PollingDelay))
}

// TestResponder handles the response to the Test request. The method always
// closes the http.Response Body.
func (client InputsClient) TestResponder(resp *http.Response) (result ResourceTestStatus, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusAccepted),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// Update updates an existing input under an existing streaming job. This can
// be used to partially update (ie. update one or two properties) an input
// without affecting the rest the job or input definition.
//
// input is an Input object. The properties specified here will overwrite the
// corresponding properties in the existing input (ie. Those properties will be
// updated). Any properties that are set to null here will mean that the
// corresponding property in the existing input will remain the same and not
// change as a result of this PATCH operation. resourceGroupName is the name of
// the resource group that contains the resource. You can obtain this value
// from the Azure Resource Manager API or the portal. jobName is the name of
// the streaming job. inputName is the name of the input. ifMatch is the ETag
// of the input. Omit this value to always overwrite the current input. Specify
// the last-seen ETag value to prevent accidentally overwritting concurrent
// changes.
func (client InputsClient) Update(input Input, resourceGroupName string, jobName string, inputName string, ifMatch string) (result Input, err error) {
	req, err := client.UpdatePreparer(input, resourceGroupName, jobName, inputName, ifMatch)
	if err != nil {
		err = autorest.NewErrorWithError(err, "streamanalytics.InputsClient", "Update", nil, "Failure preparing request")
		return
	}

	resp, err := client.UpdateSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		err = autorest.NewErrorWithError(err, "streamanalytics.InputsClient", "Update", resp, "Failure sending request")
		return
	}

	result, err = client.UpdateResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "streamanalytics.InputsClient", "Update", resp, "Failure responding to request")
	}

	return
}

// UpdatePreparer prepares the Update request.
func (client InputsClient) UpdatePreparer(input Input, resourceGroupName string, jobName string, inputName string, ifMatch string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"inputName":         autorest.Encode("path", inputName),
		"jobName":           autorest.Encode("path", jobName),
		"resourceGroupName": autorest.Encode("path", resourceGroupName),
		"subscriptionId":    autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2016-03-01"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsJSON(),
		autorest.AsPatch(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}/providers/Microsoft.StreamAnalytics/streamingjobs/{jobName}/inputs/{inputName}", pathParameters),
		autorest.WithJSON(input),
		autorest.WithQueryParameters(queryParameters))
	if len(ifMatch) > 0 {
		preparer = autorest.DecoratePreparer(preparer,
			autorest.WithHeader("If-Match", autorest.String(ifMatch)))
	}
	return preparer.Prepare(&http.Request{})
}

// UpdateSender sends the Update request. The method will close the
// http.Response Body if it receives an error.
func (client InputsClient) UpdateSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client, req)
}

// UpdateResponder handles the response to the Update request. The method always
// closes the http.Response Body.
func (client InputsClient) UpdateResponder(resp *http.Response) (result Input, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

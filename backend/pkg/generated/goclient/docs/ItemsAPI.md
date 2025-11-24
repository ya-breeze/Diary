# \ItemsAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetItems**](ItemsAPI.md#GetItems) | **Get** /v1/items | get diary items
[**PutItems**](ItemsAPI.md#PutItems) | **Put** /v1/items | upsert diary item



## GetItems

> ItemsListResponse GetItems(ctx).Date(date).Search(search).Tags(tags).Execute()

get diary items

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	date := time.Now() // string | filter items by date (optional) (optional)
	search := "vacation" // string | search text to filter items by title and body content (optional)
	tags := "personal,work" // string | comma-separated list of tags to filter items (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ItemsAPI.GetItems(context.Background()).Date(date).Search(search).Tags(tags).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ItemsAPI.GetItems``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetItems`: ItemsListResponse
	fmt.Fprintf(os.Stdout, "Response from `ItemsAPI.GetItems`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetItemsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **date** | **string** | filter items by date (optional) | 
 **search** | **string** | search text to filter items by title and body content | 
 **tags** | **string** | comma-separated list of tags to filter items | 

### Return type

[**ItemsListResponse**](ItemsListResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## PutItems

> ItemsResponse PutItems(ctx).ItemsRequest(itemsRequest).Execute()

upsert diary item

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	itemsRequest := *openapiclient.NewItemsRequest(time.Now(), "My diary entry", "Today was a great day...") // ItemsRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ItemsAPI.PutItems(context.Background()).ItemsRequest(itemsRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ItemsAPI.PutItems``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `PutItems`: ItemsResponse
	fmt.Fprintf(os.Stdout, "Response from `ItemsAPI.PutItems`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiPutItemsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **itemsRequest** | [**ItemsRequest**](ItemsRequest.md) |  | 

### Return type

[**ItemsResponse**](ItemsResponse.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


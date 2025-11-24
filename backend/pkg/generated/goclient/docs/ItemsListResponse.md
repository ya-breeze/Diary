# ItemsListResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Items** | [**[]ItemsResponse**](ItemsResponse.md) | List of diary items matching the search criteria | 
**TotalCount** | **int32** | Total number of items found | 

## Methods

### NewItemsListResponse

`func NewItemsListResponse(items []ItemsResponse, totalCount int32, ) *ItemsListResponse`

NewItemsListResponse instantiates a new ItemsListResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewItemsListResponseWithDefaults

`func NewItemsListResponseWithDefaults() *ItemsListResponse`

NewItemsListResponseWithDefaults instantiates a new ItemsListResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetItems

`func (o *ItemsListResponse) GetItems() []ItemsResponse`

GetItems returns the Items field if non-nil, zero value otherwise.

### GetItemsOk

`func (o *ItemsListResponse) GetItemsOk() (*[]ItemsResponse, bool)`

GetItemsOk returns a tuple with the Items field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetItems

`func (o *ItemsListResponse) SetItems(v []ItemsResponse)`

SetItems sets Items field to given value.


### GetTotalCount

`func (o *ItemsListResponse) GetTotalCount() int32`

GetTotalCount returns the TotalCount field if non-nil, zero value otherwise.

### GetTotalCountOk

`func (o *ItemsListResponse) GetTotalCountOk() (*int32, bool)`

GetTotalCountOk returns a tuple with the TotalCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalCount

`func (o *ItemsListResponse) SetTotalCount(v int32)`

SetTotalCount sets TotalCount field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



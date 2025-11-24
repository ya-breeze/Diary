# ItemsRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Date** | **string** |  | 
**Title** | **string** |  | 
**Tags** | Pointer to **[]string** |  | [optional] 
**Body** | **string** |  | 

## Methods

### NewItemsRequest

`func NewItemsRequest(date string, title string, body string, ) *ItemsRequest`

NewItemsRequest instantiates a new ItemsRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewItemsRequestWithDefaults

`func NewItemsRequestWithDefaults() *ItemsRequest`

NewItemsRequestWithDefaults instantiates a new ItemsRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDate

`func (o *ItemsRequest) GetDate() string`

GetDate returns the Date field if non-nil, zero value otherwise.

### GetDateOk

`func (o *ItemsRequest) GetDateOk() (*string, bool)`

GetDateOk returns a tuple with the Date field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDate

`func (o *ItemsRequest) SetDate(v string)`

SetDate sets Date field to given value.


### GetTitle

`func (o *ItemsRequest) GetTitle() string`

GetTitle returns the Title field if non-nil, zero value otherwise.

### GetTitleOk

`func (o *ItemsRequest) GetTitleOk() (*string, bool)`

GetTitleOk returns a tuple with the Title field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTitle

`func (o *ItemsRequest) SetTitle(v string)`

SetTitle sets Title field to given value.


### GetTags

`func (o *ItemsRequest) GetTags() []string`

GetTags returns the Tags field if non-nil, zero value otherwise.

### GetTagsOk

`func (o *ItemsRequest) GetTagsOk() (*[]string, bool)`

GetTagsOk returns a tuple with the Tags field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTags

`func (o *ItemsRequest) SetTags(v []string)`

SetTags sets Tags field to given value.

### HasTags

`func (o *ItemsRequest) HasTags() bool`

HasTags returns a boolean if a field has been set.

### GetBody

`func (o *ItemsRequest) GetBody() string`

GetBody returns the Body field if non-nil, zero value otherwise.

### GetBodyOk

`func (o *ItemsRequest) GetBodyOk() (*string, bool)`

GetBodyOk returns a tuple with the Body field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBody

`func (o *ItemsRequest) SetBody(v string)`

SetBody sets Body field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



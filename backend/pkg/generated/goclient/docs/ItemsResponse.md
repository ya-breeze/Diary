# ItemsResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Date** | **string** |  | 
**Title** | **string** |  | 
**Tags** | Pointer to **[]string** |  | [optional] 
**Body** | **string** |  | 
**PreviousDate** | Pointer to **NullableString** |  | [optional] 
**NextDate** | Pointer to **NullableString** |  | [optional] 

## Methods

### NewItemsResponse

`func NewItemsResponse(date string, title string, body string, ) *ItemsResponse`

NewItemsResponse instantiates a new ItemsResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewItemsResponseWithDefaults

`func NewItemsResponseWithDefaults() *ItemsResponse`

NewItemsResponseWithDefaults instantiates a new ItemsResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDate

`func (o *ItemsResponse) GetDate() string`

GetDate returns the Date field if non-nil, zero value otherwise.

### GetDateOk

`func (o *ItemsResponse) GetDateOk() (*string, bool)`

GetDateOk returns a tuple with the Date field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDate

`func (o *ItemsResponse) SetDate(v string)`

SetDate sets Date field to given value.


### GetTitle

`func (o *ItemsResponse) GetTitle() string`

GetTitle returns the Title field if non-nil, zero value otherwise.

### GetTitleOk

`func (o *ItemsResponse) GetTitleOk() (*string, bool)`

GetTitleOk returns a tuple with the Title field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTitle

`func (o *ItemsResponse) SetTitle(v string)`

SetTitle sets Title field to given value.


### GetTags

`func (o *ItemsResponse) GetTags() []string`

GetTags returns the Tags field if non-nil, zero value otherwise.

### GetTagsOk

`func (o *ItemsResponse) GetTagsOk() (*[]string, bool)`

GetTagsOk returns a tuple with the Tags field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTags

`func (o *ItemsResponse) SetTags(v []string)`

SetTags sets Tags field to given value.

### HasTags

`func (o *ItemsResponse) HasTags() bool`

HasTags returns a boolean if a field has been set.

### GetBody

`func (o *ItemsResponse) GetBody() string`

GetBody returns the Body field if non-nil, zero value otherwise.

### GetBodyOk

`func (o *ItemsResponse) GetBodyOk() (*string, bool)`

GetBodyOk returns a tuple with the Body field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBody

`func (o *ItemsResponse) SetBody(v string)`

SetBody sets Body field to given value.


### GetPreviousDate

`func (o *ItemsResponse) GetPreviousDate() string`

GetPreviousDate returns the PreviousDate field if non-nil, zero value otherwise.

### GetPreviousDateOk

`func (o *ItemsResponse) GetPreviousDateOk() (*string, bool)`

GetPreviousDateOk returns a tuple with the PreviousDate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreviousDate

`func (o *ItemsResponse) SetPreviousDate(v string)`

SetPreviousDate sets PreviousDate field to given value.

### HasPreviousDate

`func (o *ItemsResponse) HasPreviousDate() bool`

HasPreviousDate returns a boolean if a field has been set.

### SetPreviousDateNil

`func (o *ItemsResponse) SetPreviousDateNil(b bool)`

 SetPreviousDateNil sets the value for PreviousDate to be an explicit nil

### UnsetPreviousDate
`func (o *ItemsResponse) UnsetPreviousDate()`

UnsetPreviousDate ensures that no value is present for PreviousDate, not even an explicit nil
### GetNextDate

`func (o *ItemsResponse) GetNextDate() string`

GetNextDate returns the NextDate field if non-nil, zero value otherwise.

### GetNextDateOk

`func (o *ItemsResponse) GetNextDateOk() (*string, bool)`

GetNextDateOk returns a tuple with the NextDate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNextDate

`func (o *ItemsResponse) SetNextDate(v string)`

SetNextDate sets NextDate field to given value.

### HasNextDate

`func (o *ItemsResponse) HasNextDate() bool`

HasNextDate returns a boolean if a field has been set.

### SetNextDateNil

`func (o *ItemsResponse) SetNextDateNil(b bool)`

 SetNextDateNil sets the value for NextDate to be an explicit nil

### UnsetNextDate
`func (o *ItemsResponse) UnsetNextDate()`

UnsetNextDate ensures that no value is present for NextDate, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



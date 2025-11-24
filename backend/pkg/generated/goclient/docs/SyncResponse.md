# SyncResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Changes** | [**[]SyncChangeResponse**](SyncChangeResponse.md) | List of changes since the requested ID | 
**HasMore** | **bool** | Whether there are more changes available | 
**NextId** | Pointer to **int32** | ID to use for the next sync request (if hasMore is true) | [optional] 

## Methods

### NewSyncResponse

`func NewSyncResponse(changes []SyncChangeResponse, hasMore bool, ) *SyncResponse`

NewSyncResponse instantiates a new SyncResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSyncResponseWithDefaults

`func NewSyncResponseWithDefaults() *SyncResponse`

NewSyncResponseWithDefaults instantiates a new SyncResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetChanges

`func (o *SyncResponse) GetChanges() []SyncChangeResponse`

GetChanges returns the Changes field if non-nil, zero value otherwise.

### GetChangesOk

`func (o *SyncResponse) GetChangesOk() (*[]SyncChangeResponse, bool)`

GetChangesOk returns a tuple with the Changes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChanges

`func (o *SyncResponse) SetChanges(v []SyncChangeResponse)`

SetChanges sets Changes field to given value.


### GetHasMore

`func (o *SyncResponse) GetHasMore() bool`

GetHasMore returns the HasMore field if non-nil, zero value otherwise.

### GetHasMoreOk

`func (o *SyncResponse) GetHasMoreOk() (*bool, bool)`

GetHasMoreOk returns a tuple with the HasMore field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHasMore

`func (o *SyncResponse) SetHasMore(v bool)`

SetHasMore sets HasMore field to given value.


### GetNextId

`func (o *SyncResponse) GetNextId() int32`

GetNextId returns the NextId field if non-nil, zero value otherwise.

### GetNextIdOk

`func (o *SyncResponse) GetNextIdOk() (*int32, bool)`

GetNextIdOk returns a tuple with the NextId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNextId

`func (o *SyncResponse) SetNextId(v int32)`

SetNextId sets NextId field to given value.

### HasNextId

`func (o *SyncResponse) HasNextId() bool`

HasNextId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



# SyncChangeResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **int32** | Unique change ID | 
**UserId** | **string** | User ID who made the change | 
**Date** | **string** | Date of the diary entry that was changed | 
**OperationType** | **string** | Type of operation performed | 
**Timestamp** | **time.Time** | When the change occurred | 
**ItemSnapshot** | Pointer to [**NullableItemsResponse**](ItemsResponse.md) | Current state of the item (null for deleted items) | [optional] 
**Metadata** | Pointer to **[]string** | Additional metadata about the change | [optional] 

## Methods

### NewSyncChangeResponse

`func NewSyncChangeResponse(id int32, userId string, date string, operationType string, timestamp time.Time, ) *SyncChangeResponse`

NewSyncChangeResponse instantiates a new SyncChangeResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSyncChangeResponseWithDefaults

`func NewSyncChangeResponseWithDefaults() *SyncChangeResponse`

NewSyncChangeResponseWithDefaults instantiates a new SyncChangeResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *SyncChangeResponse) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *SyncChangeResponse) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *SyncChangeResponse) SetId(v int32)`

SetId sets Id field to given value.


### GetUserId

`func (o *SyncChangeResponse) GetUserId() string`

GetUserId returns the UserId field if non-nil, zero value otherwise.

### GetUserIdOk

`func (o *SyncChangeResponse) GetUserIdOk() (*string, bool)`

GetUserIdOk returns a tuple with the UserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserId

`func (o *SyncChangeResponse) SetUserId(v string)`

SetUserId sets UserId field to given value.


### GetDate

`func (o *SyncChangeResponse) GetDate() string`

GetDate returns the Date field if non-nil, zero value otherwise.

### GetDateOk

`func (o *SyncChangeResponse) GetDateOk() (*string, bool)`

GetDateOk returns a tuple with the Date field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDate

`func (o *SyncChangeResponse) SetDate(v string)`

SetDate sets Date field to given value.


### GetOperationType

`func (o *SyncChangeResponse) GetOperationType() string`

GetOperationType returns the OperationType field if non-nil, zero value otherwise.

### GetOperationTypeOk

`func (o *SyncChangeResponse) GetOperationTypeOk() (*string, bool)`

GetOperationTypeOk returns a tuple with the OperationType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOperationType

`func (o *SyncChangeResponse) SetOperationType(v string)`

SetOperationType sets OperationType field to given value.


### GetTimestamp

`func (o *SyncChangeResponse) GetTimestamp() time.Time`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *SyncChangeResponse) GetTimestampOk() (*time.Time, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *SyncChangeResponse) SetTimestamp(v time.Time)`

SetTimestamp sets Timestamp field to given value.


### GetItemSnapshot

`func (o *SyncChangeResponse) GetItemSnapshot() ItemsResponse`

GetItemSnapshot returns the ItemSnapshot field if non-nil, zero value otherwise.

### GetItemSnapshotOk

`func (o *SyncChangeResponse) GetItemSnapshotOk() (*ItemsResponse, bool)`

GetItemSnapshotOk returns a tuple with the ItemSnapshot field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetItemSnapshot

`func (o *SyncChangeResponse) SetItemSnapshot(v ItemsResponse)`

SetItemSnapshot sets ItemSnapshot field to given value.

### HasItemSnapshot

`func (o *SyncChangeResponse) HasItemSnapshot() bool`

HasItemSnapshot returns a boolean if a field has been set.

### SetItemSnapshotNil

`func (o *SyncChangeResponse) SetItemSnapshotNil(b bool)`

 SetItemSnapshotNil sets the value for ItemSnapshot to be an explicit nil

### UnsetItemSnapshot
`func (o *SyncChangeResponse) UnsetItemSnapshot()`

UnsetItemSnapshot ensures that no value is present for ItemSnapshot, not even an explicit nil
### GetMetadata

`func (o *SyncChangeResponse) GetMetadata() []string`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *SyncChangeResponse) GetMetadataOk() (*[]string, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *SyncChangeResponse) SetMetadata(v []string)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *SyncChangeResponse) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



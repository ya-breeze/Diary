# AssetsBatchResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Files** | [**[]AssetsBatchFile**](AssetsBatchFile.md) |  | 
**Count** | **int32** | Number of files successfully uploaded | 

## Methods

### NewAssetsBatchResponse

`func NewAssetsBatchResponse(files []AssetsBatchFile, count int32, ) *AssetsBatchResponse`

NewAssetsBatchResponse instantiates a new AssetsBatchResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetsBatchResponseWithDefaults

`func NewAssetsBatchResponseWithDefaults() *AssetsBatchResponse`

NewAssetsBatchResponseWithDefaults instantiates a new AssetsBatchResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFiles

`func (o *AssetsBatchResponse) GetFiles() []AssetsBatchFile`

GetFiles returns the Files field if non-nil, zero value otherwise.

### GetFilesOk

`func (o *AssetsBatchResponse) GetFilesOk() (*[]AssetsBatchFile, bool)`

GetFilesOk returns a tuple with the Files field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFiles

`func (o *AssetsBatchResponse) SetFiles(v []AssetsBatchFile)`

SetFiles sets Files field to given value.


### GetCount

`func (o *AssetsBatchResponse) GetCount() int32`

GetCount returns the Count field if non-nil, zero value otherwise.

### GetCountOk

`func (o *AssetsBatchResponse) GetCountOk() (*int32, bool)`

GetCountOk returns a tuple with the Count field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCount

`func (o *AssetsBatchResponse) SetCount(v int32)`

SetCount sets Count field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



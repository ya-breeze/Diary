# AssetsBatchFile

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**OriginalName** | **string** | Original filename provided by the client | 
**SavedName** | **string** | Server-side stored filename | 
**Size** | **int64** | File size in bytes | 
**ContentType** | Pointer to **string** | MIME type detected for the file | [optional] 

## Methods

### NewAssetsBatchFile

`func NewAssetsBatchFile(originalName string, savedName string, size int64, ) *AssetsBatchFile`

NewAssetsBatchFile instantiates a new AssetsBatchFile object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetsBatchFileWithDefaults

`func NewAssetsBatchFileWithDefaults() *AssetsBatchFile`

NewAssetsBatchFileWithDefaults instantiates a new AssetsBatchFile object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetOriginalName

`func (o *AssetsBatchFile) GetOriginalName() string`

GetOriginalName returns the OriginalName field if non-nil, zero value otherwise.

### GetOriginalNameOk

`func (o *AssetsBatchFile) GetOriginalNameOk() (*string, bool)`

GetOriginalNameOk returns a tuple with the OriginalName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOriginalName

`func (o *AssetsBatchFile) SetOriginalName(v string)`

SetOriginalName sets OriginalName field to given value.


### GetSavedName

`func (o *AssetsBatchFile) GetSavedName() string`

GetSavedName returns the SavedName field if non-nil, zero value otherwise.

### GetSavedNameOk

`func (o *AssetsBatchFile) GetSavedNameOk() (*string, bool)`

GetSavedNameOk returns a tuple with the SavedName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSavedName

`func (o *AssetsBatchFile) SetSavedName(v string)`

SetSavedName sets SavedName field to given value.


### GetSize

`func (o *AssetsBatchFile) GetSize() int64`

GetSize returns the Size field if non-nil, zero value otherwise.

### GetSizeOk

`func (o *AssetsBatchFile) GetSizeOk() (*int64, bool)`

GetSizeOk returns a tuple with the Size field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSize

`func (o *AssetsBatchFile) SetSize(v int64)`

SetSize sets Size field to given value.


### GetContentType

`func (o *AssetsBatchFile) GetContentType() string`

GetContentType returns the ContentType field if non-nil, zero value otherwise.

### GetContentTypeOk

`func (o *AssetsBatchFile) GetContentTypeOk() (*string, bool)`

GetContentTypeOk returns a tuple with the ContentType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContentType

`func (o *AssetsBatchFile) SetContentType(v string)`

SetContentType sets ContentType field to given value.

### HasContentType

`func (o *AssetsBatchFile) HasContentType() bool`

HasContentType returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)



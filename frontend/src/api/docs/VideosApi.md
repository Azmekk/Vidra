# VideosApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**createVideo**](#createvideo) | **POST** /api/videos | Create a new video download task|
|[**deleteVideo**](#deletevideo) | **DELETE** /api/videos/{id} | Delete a video|
|[**getMetadata**](#getmetadata) | **POST** /api/videos/metadata | Get video metadata and options|
|[**getProgress**](#getprogress) | **GET** /api/videos/{id}/progress | Get download progress|
|[**getVideo**](#getvideo) | **GET** /api/videos/{id} | Get a video by ID|
|[**listAllProgress**](#listallprogress) | **GET** /api/videos/progress | List all video download progress|
|[**listVideos**](#listvideos) | **GET** /api/videos | List all videos|

# **createVideo**
> HandlersVideoResponse createVideo(video)

Create a new video record and start background download

### Example

```typescript
import {
    VideosApi,
    Configuration,
    HandlersCreateVideoRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new VideosApi(configuration);

let video: HandlersCreateVideoRequest; //Video details

const { status, data } = await apiInstance.createVideo(
    video
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **video** | **HandlersCreateVideoRequest**| Video details | |


### Return type

**HandlersVideoResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**201** | Created |  -  |
|**400** | Bad Request |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteVideo**
> deleteVideo()

Delete a video record by ID

### Example

```typescript
import {
    VideosApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new VideosApi(configuration);

let id: string; //Video ID (default to undefined)

const { status, data } = await apiInstance.deleteVideo(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] | Video ID | defaults to undefined|


### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**204** | No Content |  -  |
|**400** | Bad Request |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getMetadata**
> ServicesVideoMetadata getMetadata(request)

Fetch available formats and metadata for a given URL using yt-dlp

### Example

```typescript
import {
    VideosApi,
    Configuration,
    HandlersMetadataRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new VideosApi(configuration);

let request: HandlersMetadataRequest; //Video URL

const { status, data } = await apiInstance.getMetadata(
    request
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **request** | **HandlersMetadataRequest**| Video URL | |


### Return type

**ServicesVideoMetadata**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getProgress**
> ServicesDownloadProgressDTO getProgress()

Get the current download progress of a video by ID

### Example

```typescript
import {
    VideosApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new VideosApi(configuration);

let id: string; //Video ID (default to undefined)

const { status, data } = await apiInstance.getProgress(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] | Video ID | defaults to undefined|


### Return type

**ServicesDownloadProgressDTO**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**404** | Not Found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getVideo**
> HandlersVideoResponse getVideo()

Get details of a specific video

### Example

```typescript
import {
    VideosApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new VideosApi(configuration);

let id: string; //Video ID (default to undefined)

const { status, data } = await apiInstance.getVideo(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] | Video ID | defaults to undefined|


### Return type

**HandlersVideoResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listAllProgress**
> { [key: string]: ServicesDownloadProgressDTO; } listAllProgress()

Get the current download progress for all active video downloads

### Example

```typescript
import {
    VideosApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new VideosApi(configuration);

const { status, data } = await apiInstance.listAllProgress();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**{ [key: string]: ServicesDownloadProgressDTO; }**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listVideos**
> Array<HandlersVideoResponse> listVideos()

Get a list of all videos

### Example

```typescript
import {
    VideosApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new VideosApi(configuration);

const { status, data } = await apiInstance.listVideos();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**Array<HandlersVideoResponse>**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


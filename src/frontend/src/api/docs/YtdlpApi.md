# YtdlpApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**updateYtdlp**](#updateytdlp) | **POST** /api/yt-dlp/update | Update yt-dlp|

# **updateYtdlp**
> { [key: string]: string; } updateYtdlp()

Execute yt-dlp -U to update the binary

### Example

```typescript
import {
    YtdlpApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new YtdlpApi(configuration);

const { status, data } = await apiInstance.updateYtdlp();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**{ [key: string]: string; }**

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


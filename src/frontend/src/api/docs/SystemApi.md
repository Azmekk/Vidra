# SystemApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**getSystemInfo**](#getsysteminfo) | **GET** /api/system/info | Get system information|

# **getSystemInfo**
> HandlersSystemInfoResponse getSystemInfo()

Get server status and downloads directory size

### Example

```typescript
import {
    SystemApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new SystemApi(configuration);

const { status, data } = await apiInstance.getSystemInfo();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**HandlersSystemInfoResponse**

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


# ErrorsApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**listRecentErrors**](#listrecenterrors) | **GET** /api/errors | List recent errors|

# **listRecentErrors**
> Array<HandlersErrorResponse> listRecentErrors()

Get a list of the most recent system errors

### Example

```typescript
import {
    ErrorsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new ErrorsApi(configuration);

let limit: number; //Limit number of results (optional) (default to 10)

const { status, data } = await apiInstance.listRecentErrors(
    limit
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **limit** | [**number**] | Limit number of results | (optional) defaults to 10|


### Return type

**Array<HandlersErrorResponse>**

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


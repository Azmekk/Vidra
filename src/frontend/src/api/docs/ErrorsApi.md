# ErrorsApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**listRecentErrors**](#listrecenterrors) | **GET** /api/errors | List recent errors|

# **listRecentErrors**
> HandlersPaginatedErrorResponse listRecentErrors()

Get a paginated list of the most recent system errors with optional searching

### Example

```typescript
import {
    ErrorsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new ErrorsApi(configuration);

let search: string; //Search by error message or command (optional) (default to undefined)
let page: number; //Page number (default: 1) (optional) (default to undefined)
let limit: number; //Number of items per page (default: 10) (optional) (default to undefined)

const { status, data } = await apiInstance.listRecentErrors(
    search,
    page,
    limit
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **search** | [**string**] | Search by error message or command | (optional) defaults to undefined|
| **page** | [**number**] | Page number (default: 1) | (optional) defaults to undefined|
| **limit** | [**number**] | Number of items per page (default: 10) | (optional) defaults to undefined|


### Return type

**HandlersPaginatedErrorResponse**

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


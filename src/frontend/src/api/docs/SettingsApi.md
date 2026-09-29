# SettingsApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**getSettings**](#getsettings) | **GET** /api/settings | Get application settings|
|[**updateSettings**](#updatesettings) | **PUT** /api/settings | Update application settings|

# **getSettings**
> HandlersSettingsResponse getSettings()

Get current application settings

### Example

```typescript
import {
    SettingsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new SettingsApi(configuration);

const { status, data } = await apiInstance.getSettings();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**HandlersSettingsResponse**

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

# **updateSettings**
> HandlersSettingsResponse updateSettings(settings)

Update application settings

### Example

```typescript
import {
    SettingsApi,
    Configuration,
    HandlersUpdateSettingsRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new SettingsApi(configuration);

let settings: HandlersUpdateSettingsRequest; //Settings to update

const { status, data } = await apiInstance.updateSettings(
    settings
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **settings** | **HandlersUpdateSettingsRequest**| Settings to update | |


### Return type

**HandlersSettingsResponse**

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


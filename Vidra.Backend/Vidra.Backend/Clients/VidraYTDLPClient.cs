using System.Net.WebSockets;
using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;
using Microsoft.AspNetCore.WebUtilities;
using Vidra.Backend.Clients.Types.VidraYTDLPClientTypes;
using Vidra.Backend.Controllers.Models.YtDlpControllerModels;

namespace Vidra.Backend.Clients;

public class ThumbnailResponse
{
    public string ThumbnailUrl { get; set; } = "";
}

public class SizeResponse
{
    public int? FilesizeBytes { get; set; }
}

public class FormatsResponse
{
    public List<Format> Formats { get; set; } = [];
}

public class DownloadIdResponse
{
    public string DownloadId { get; set; } = "";
}

public class VidraYtdlpClient(HttpClient client)
{
    static readonly JsonSerializerOptions jsonOptions = new JsonSerializerOptions
    {
        PropertyNameCaseInsensitive = true,
        PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower,
        DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingNull
    };
    
    public async Task<VideoInfo?> GetMetadataAsync(string url)
    {
        var requestBody = new { url };
        var response = await client.PostAsJsonAsync("/metadata", requestBody);
        return await response.Content.ReadFromJsonAsync<VideoInfo>(jsonOptions);
    }

    public async Task<ThumbnailResponse?> GetThumbnailAsync(string url)
    {
        var requestBody = new { url };
        var response = await client.PostAsJsonAsync("/thumbnail", requestBody);
        return await response.Content.ReadFromJsonAsync<ThumbnailResponse>(jsonOptions);
    }

    public async Task<SizeResponse?> GetSizeAsync(string url)
    {
        var requestBody = new { url };
        var response = await client.PostAsJsonAsync("/size", requestBody);
        return await response.Content.ReadFromJsonAsync<SizeResponse>(jsonOptions);
    }

    public async Task<DownloadIdResponse?> DownloadVideoAsync(DownloadVideoRequestBody request)
    {
        var requestBody = new
        {
            url = request.Url,
            video_format_id = request.VideoFormatId,
            audio_format_id = request.AudioFormatId,
            include_thumbnail = true,
        };

        var response = await client.PostAsJsonAsync("/download", requestBody);
        return await response.Content.ReadFromJsonAsync<DownloadIdResponse>(jsonOptions);
    }

    public async Task<DownloadStatus?> GetDownloadStatusAsync(string downloadId)
    {
        var query = new Dictionary<string, string?>
        {
            { "download_id", Uri.EscapeDataString(downloadId) }
        };

        var uri = QueryHelpers.AddQueryString("/download_status", query);
        var response = await client.GetFromJsonAsync<DownloadStatus>(uri, jsonOptions);
        
        return response;
    }

    public async Task<FormatsResponse?> GetFormatsAsync(string url)
    {
        var requestBody = new { url };
        var response = await client.PostAsJsonAsync("/formats", requestBody);
        return await response.Content.ReadFromJsonAsync<FormatsResponse>(jsonOptions);
    }

    public async Task<CombinedVideoInfo?> GetCombinedVideoInfoAsync(string url)
    {
        var requestBody = new { url };
        var response = await client.PostAsJsonAsync("/combined_video_info", requestBody);
        return await response.Content.ReadFromJsonAsync<CombinedVideoInfo>(jsonOptions);
    }

}
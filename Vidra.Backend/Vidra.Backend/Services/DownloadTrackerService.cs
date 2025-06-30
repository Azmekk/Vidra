using System.Collections.Concurrent;
using System.Net.WebSockets;
using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace Vidra.Backend.Services;

public class DownloadStatus
{
    public string Status { get; set; }
    public float? Percent { get; set; }
    public string? Filename { get; set; }
    public int? DownloadedBytes { get; set; }
    public int? TotalBytes { get; set; }
    public string? Error { get; set; }
}

public delegate Task DownloadStatusChangedHandler(ConcurrentDictionary<string, DownloadStatus> downloadStatuses);

public static class DownloadTrackerContainer
{
    public static DownloadStatusChangedHandler? OnDownloadStatusChanged = null;
    public static ConcurrentDictionary<string, DownloadStatus> DownloadStatuses { get; } = new();
}
public class DownloadTrackerService(IConfiguration configuration, ILogger<DownloadTrackerService> logger) : BackgroundService
{
    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        PropertyNameCaseInsensitive = true,
        PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower,
        DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingNull
    };
    
    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        var wsStringUri = configuration.GetValue<string>("DOWNLOAD_STATUS_WS_URL");

        if (string.IsNullOrEmpty(wsStringUri))
        {
            throw new InvalidOperationException("DOWNLOAD_STATUS_WS_URL is not configured.");
        }

        var uri = new Uri(wsStringUri);
        using var webSocket = new ClientWebSocket();

        await webSocket.ConnectAsync(uri, stoppingToken);

        var buffer = new byte[1024 * 1024]; // 10 MB buffer
        while (webSocket.State == WebSocketState.Open && !stoppingToken.IsCancellationRequested)
        {
            var result = await webSocket.ReceiveAsync(new ArraySegment<byte>(buffer), stoppingToken);
            var message = Encoding.UTF8.GetString(buffer, 0, result.Count);

            if (message.Equals("{}", StringComparison.InvariantCultureIgnoreCase))
            {
                // Empty message, continue to next iteration
                continue;
            }
            
            try
            {
                var statuses = System.Text.Json.JsonSerializer.Deserialize<Dictionary<string, DownloadStatus>>(message, JsonOptions);
                if (statuses != null)
                {
                    foreach (var kvp in statuses)
                    {
                        DownloadTrackerContainer.DownloadStatuses.AddOrUpdate(kvp.Key, kvp.Value, (key, oldValue) => kvp.Value);
                    }
                    
                    DownloadTrackerContainer.OnDownloadStatusChanged?.Invoke(DownloadTrackerContainer.DownloadStatuses);
                }
            }
            catch (Exception ex)
            {
                logger.LogError(ex, "Failed to deserialize download status message: {Message}", message);
            }
        }
    }
}
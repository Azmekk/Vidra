using System.Collections.Concurrent;
using Microsoft.AspNetCore.SignalR;
using Vidra.Backend.Services;

namespace Vidra.Backend.SignalRHubs;

public class VideoProgressHub : Hub
{
    public VideoProgressHub()
    {
        if (DownloadTrackerContainer.OnDownloadStatusChanged is null)
        {
            DownloadTrackerContainer.OnDownloadStatusChanged += OnDownloadStatusChanged;
        }
    }
    
    public async Task OnDownloadStatusChanged(ConcurrentDictionary<string, DownloadStatus> downloadStatuses)
    {
        await Clients.All.SendAsync("DownloadStatusChanged", downloadStatuses);
    }
}
namespace Vidra.Backend.Clients.Types.VidraYTDLPClientTypes;

public class DownloadStatus
{
    public string Status { get; set; }
    public float? Percent { get; set; }
    public string Filename { get; set; }
    public int? DownloadedBytes { get; set; }
    public int? TotalBytes { get; set; }
    public string Error { get; set; }
    public string Thumbnail { get; set; }
    public string ThumbnailError { get; set; }
}
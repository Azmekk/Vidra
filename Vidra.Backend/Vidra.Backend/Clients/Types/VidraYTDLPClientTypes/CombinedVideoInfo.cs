namespace Vidra.Backend.Clients.Types.VidraYTDLPClientTypes;

public class CombinedVideoInfo
{
    public string Id { get; set; }
    public string Title { get; set; }
    public float Duration { get; set; }
    public string Uploader { get; set; }
    public int? ViewCount { get; set; }
    public int? LikeCount { get; set; }
    public string Thumbnail { get; set; }
    public int? Filesize { get; set; }
    public string Url { get; set; }
    public List<Format> Formats { get; set; }
}
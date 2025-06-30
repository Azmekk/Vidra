using System.ComponentModel.DataAnnotations;
using System.Text.Json.Serialization;

namespace Vidra.Backend.Database.Entities;

public class Video : BaseEntity
{
    public string ExternalId { get; set; } = "";
    public string Guid { get; set; } = "";
    
    //Metadata
    public string Title { get; set; } = "";
    public double Duration { get; set; } = 0.0;
    public string Uploader { get; set; } = "";
    public int? ViewCount { get; set; } = 0;
    public int? LikeCount { get; set; } = null;
    public string ThumbnailUrl { get; set; } = "";
    public string ThumbnailPath { get; set; } = "";
    public int? Filesize { get; set; } = null;
    public string Url { get; set; } = "";

    public List<VideoVariant> VideoVariants { get; set; } = [];
}
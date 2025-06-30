namespace Vidra.Backend.Database.Entities;

public class VideoVariant : BaseEntity
{
    public string? ExternalFormatId { get; set; } = "";
    public string IsCustom { get; set; } = "";
    public string FilePath { get; set; } = "";
    public string VideoCodec { get; set; } = "";
    public int VideoBitrate { get; set; } = 0;
    public string AudioCodec { get; set; } = "";
    public int AudioBitrate { get; set; } = 0;
    public long Size { get; set; } = 0;
    public string Extension { get; set; } = "";
    public string Note { get; set; } = "";
    public double? Fps { get; set; } = null;
    public bool IsDefault { get; set; } = false;
    
    public int VideoId { get; set; }
    public Video? Video { get; set; } = null!;
}
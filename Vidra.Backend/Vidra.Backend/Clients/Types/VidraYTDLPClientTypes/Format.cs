namespace Vidra.Backend.Clients.Types.VidraYTDLPClientTypes;

public class Format
{
    public string FormatId { get; set; }
    public string Ext { get; set; }
    public string FormatNote { get; set; }
    public int? Filesize { get; set; }
    public string Resolution { get; set; }
    public float? Fps { get; set; }
    public string Vcodec { get; set; }
    public string Acodec { get; set; }
}
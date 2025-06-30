using System.ComponentModel.DataAnnotations;

namespace Vidra.Backend.Controllers.Models.YtDlpControllerModels;

public class DownloadVideoRequestBody
{
    [Required]
    public string Url { get; set; } = "";

    public string? Name { get; set; } = "";
    public string? VideoFormatId { get; set; } = null;
    public string? AudioFormatId { get; set; } = null;
}
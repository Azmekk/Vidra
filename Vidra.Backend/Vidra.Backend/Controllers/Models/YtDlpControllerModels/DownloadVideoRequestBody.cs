using System.ComponentModel.DataAnnotations;

namespace Vidra.Backend.Controllers.Models.YtDlpControllerModels;

public class DownloadVideoRequestBody
{
    [Required]
    public string Url { get; set; } = "";

    public string? Name { get; set; } = "";
    public string? FormatId { get; set; } = null;
}
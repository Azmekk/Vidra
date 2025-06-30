using System.ComponentModel.DataAnnotations;

namespace Vidra.Backend.Controllers.Models.YtDlpControllerModels;

public class GetCombinedVideoInformationRequestBody
{
    [Required] public string Url { get; set; } = "";
}
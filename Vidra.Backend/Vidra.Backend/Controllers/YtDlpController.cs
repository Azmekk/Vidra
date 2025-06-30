using Microsoft.AspNetCore.Mvc;
using Vidra.Backend.Clients;
using Vidra.Backend.Controllers.Models.YtDlpControllerModels;

namespace Vidra.Backend.Controllers;

[Route("api/[controller]/[action]")]
public class YtDlpController(VidraYtdlpClient vidraYtdlpClient) : Controller
{
    [HttpPost]
    public async Task<IActionResult> DownloadVideo([FromBody] DownloadVideoRequestBody requestBody)
    {
        if (!ModelState.IsValid)
        {
            return BadRequest(ModelState);
        }

        try
        {
            var video = await vidraYtdlpClient.DownloadVideoAsync(requestBody);
            return Ok(video);
        }
        catch (Exception ex)
        {
            return StatusCode(500, $"Internal server error: {ex.Message}");
        }
    }

    [HttpPost]
    public async Task<IActionResult> GetCombinedVideoInformation([FromBody] GetCombinedVideoInformationRequestBody requestBody)
    {
        if (!ModelState.IsValid)
        {
            return BadRequest(ModelState);
        }

        try
        {
            var formats = await vidraYtdlpClient.GetCombinedVideoInfoAsync(requestBody.Url);
            return Ok(formats);
        }
        catch (Exception ex)
        {
            return StatusCode(500, $"Internal server error: {ex.Message}");
        }
    }
}
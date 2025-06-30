using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using Vidra.Backend.Clients;
using Vidra.Backend.Database;
using Vidra.Backend.Database.Entities;

namespace Vidra.Backend.Controllers;

[Route("api/[controller]/[action]")]
public class VideoController(VidraDbContext vidraDbContext) : Controller
{
    [HttpGet]
    public async Task<ActionResult<List<Video>>> GetVideosInfo([FromQuery] int take, [FromQuery] int page,
        [FromQuery] string? search = null)
    {
        var query = vidraDbContext.Videos.Include(x => x.VideoVariants).AsQueryable();

        if (!string.IsNullOrEmpty(search))
        {
            query = query.Where(v => v.Title.ToLower().Contains(search.ToLower()));
        }

        var result =  await query
            .OrderByDescending(v => v.CreatedAt)
            .Skip(take * (page - 1))
            .Take(take)
            .ToListAsync();
        
        return Ok(result);
    }
    
    [HttpGet]
    public async Task<ActionResult<Video>> GetVideoInfo([FromQuery] int videoId)
    {
        var video = await vidraDbContext.Videos
            .Include(x => x.VideoVariants)
            .FirstOrDefaultAsync(x => x.Id == videoId && x.DeletedAt == null);

        if (video == null)
        {
            return NotFound("Video not found.");
        }

        return video;
    }

    public async Task<IActionResult> DeleteVideo([FromQuery] int videoId)
    {
        using var transaction = await vidraDbContext.Database.BeginTransactionAsync();

        try
        {
            var video = await vidraDbContext.Videos
                .Include(x => x.VideoVariants)
                .FirstOrDefaultAsync(x => x.Id == videoId && x.DeletedAt == null);

            if (video == null)
            {
                return NotFound("Video not found.");
            }

            if (System.IO.File.Exists(video.ThumbnailPath))
            {
                System.IO.File.Delete(video.ThumbnailPath);
            }

            foreach (var variant in video.VideoVariants.Where(variant => System.IO.File.Exists(variant.FilePath)))
            {
                System.IO.File.Delete(variant.FilePath);
            }

            vidraDbContext.Videos.Remove(video);
            await vidraDbContext.SaveChangesAsync();

            await transaction.CommitAsync();

            return Ok("Video deleted successfully.");
        }
        catch (Exception ex)
        {
            await transaction.RollbackAsync();
            return StatusCode(500, $"Error during deletion: {ex.Message}");
        }
    }
}
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.FileProviders;
using Vidra.Backend.Database;

namespace Vidra.Backend.Extensions;

public static class WebApplicationExtensions
{
    public static async Task<WebApplication> EnsureDatabasesCreatedAsync(this WebApplication app)
    {
        using var scope = app.Services.CreateScope();
        var dbContext = scope.ServiceProvider.GetRequiredService<VidraDbContext>();
        
        await dbContext.Database.EnsureCreatedAsync();
        
        return app;
    }

    public static async Task<WebApplication> EnsureDatabasesMigratedAsync(this WebApplication app)
    {
        using var scope = app.Services.CreateScope();
        var dbContext = scope.ServiceProvider.GetRequiredService<VidraDbContext>();
        
        await dbContext.Database.MigrateAsync();
        
        return app;
    }
    
    public static WebApplication ConfigureFileServer(this WebApplication app)
    {
        string? downloadPath = app.Configuration.GetValue<string>("VIDEO_DOWNLOAD_PATH");
        
        if (string.IsNullOrEmpty(downloadPath))
        {
            throw new InvalidOperationException("VIDEO_DOWNLOAD_PATH is not configured.");
        }

        var fileProvider = new PhysicalFileProvider(downloadPath);
        var videoDownloadOptions = new FileServerOptions()
        {
            RequestPath = "/Downloads",
            FileProvider = fileProvider,
            EnableDirectoryBrowsing = true,
            StaticFileOptions = 
            {
                ServeUnknownFileTypes = true,
                DefaultContentType = "application/octet-stream"
            },
            
        };
        
        app.UseFileServer(videoDownloadOptions);
        
        return app;
    }
}
using Microsoft.EntityFrameworkCore;
using Vidra.Backend.Clients;
using Vidra.Backend.Database;

namespace Vidra.Backend.Extensions;

public static class WebApplicationBuilderExtensions
{
    public static WebApplicationBuilder RegisterDatabases(this WebApplicationBuilder builder)
    {
        var connectionString = builder.Configuration.GetValue<string>("VIDRA_DB");
        
        builder.Services.AddDbContext<VidraDbContext>(options =>
        {
            options.UseNpgsql(connectionString);
        });
        return builder;
    }

    public static WebApplicationBuilder RegisterClients(this WebApplicationBuilder builder)
    {
        var baseYtDlpApiUrl = builder.Configuration.GetValue<string>("YTDLP_API_BASE_URL");
        
        if (string.IsNullOrEmpty(baseYtDlpApiUrl))
        {
            throw new InvalidOperationException("YTDLP_API_BASE_URL is not configured.");
        }
        
        builder.Services.AddHttpClient<VidraYtdlpClient>(client =>
        {
            client.BaseAddress = new Uri(baseYtDlpApiUrl);
        });
        
        return builder;
    }
}
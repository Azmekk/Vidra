using Vidra.Backend.Controllers;
using Vidra.Backend.Extensions;
using Vidra.Backend.SignalRHubs;

namespace Vidra.Backend;

public class Program
{
    public static async Task Main(string[] args)
    {
        var builder = WebApplication.CreateBuilder(args);

        // Add services to the container.

        builder.Services.AddControllers();
        // Learn more about configuring OpenAPI at https://aka.ms/aspnet/openapi
        builder.Services.AddOpenApi();
        
        builder.RegisterDatabases();
        builder.RegisterClients();
        builder.RegisterBackgroundServices();

        builder.Services.AddSignalR();

        var app = builder.Build();

        await app.EnsureDatabasesCreatedAsync();
        await app.EnsureDatabasesMigratedAsync();
        
        // Configure the HTTP request pipeline.
        if (app.Environment.IsDevelopment())
        {
            app.MapOpenApi();
        }
        
        app.UseAuthorization();

        
        app.MapControllers();
        app.ConfigureFileServer();
        
        app.MapHub<VideoProgressHub>("/hubs/video-progress");

        await app.RunAsync();
    }
    
    
}
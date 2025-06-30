using Vidra.Backend.Extensions;

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

        await app.RunAsync();
    }
    
    
}
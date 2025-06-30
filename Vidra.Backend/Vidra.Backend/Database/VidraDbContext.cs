using Microsoft.EntityFrameworkCore;
using Vidra.Backend.Database.Entities;

namespace Vidra.Backend.Database;

public class VidraDbContext(DbContextOptions<VidraDbContext> options) : DbContext(options)
{
    public DbSet<Video> Videos { get; set; }
    public DbSet<VideoVariant> VideoVariants { get; set; }
}
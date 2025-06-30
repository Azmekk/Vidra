using Microsoft.EntityFrameworkCore;
using Vidra.Backend.Database.Entities;

namespace Vidra.Backend.Database;

public class VidraDbContext(DbContextOptions<VidraDbContext> options) : DbContext(options)
{
    public DbSet<Video> Videos { get; set; }
    public DbSet<VideoVariant> VideoVariants { get; set; }

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        base.OnModelCreating(modelBuilder);

        modelBuilder.Entity<Video>().HasQueryFilter(v => v.DeletedAt != null);
        modelBuilder.Entity<VideoVariant>().HasQueryFilter(vv => vv.DeletedAt != null);
    }
    
    
    public override int SaveChanges()
    {
        AddTimestamps();
        return base.SaveChanges();
    }
    
    public override async Task<int> SaveChangesAsync(CancellationToken cancellationToken = default)
    {
        AddTimestamps();
        return await base.SaveChangesAsync(cancellationToken);
    }

    private void AddTimestamps()
    {
        var entities = ChangeTracker.Entries()
            .Where(x => x.Entity is BaseEntity && (x.State == EntityState.Added || x.State == EntityState.Modified));

        foreach (var entity in entities)
        {
            var now = DateTime.UtcNow;

            if (entity.State == EntityState.Added)
            {
                ((BaseEntity)entity.Entity).CreatedAt = now;
            }
            ((BaseEntity)entity.Entity).UpdatedAt = now;
        }
    }
}
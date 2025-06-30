using System.ComponentModel.DataAnnotations;

namespace Vidra.Backend.Database.Entities;

public class BaseEntity
{
    [Key]
    public int Id { get; set; }
    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
    public DateTime? DeletedAt { get; set; } = null;
}
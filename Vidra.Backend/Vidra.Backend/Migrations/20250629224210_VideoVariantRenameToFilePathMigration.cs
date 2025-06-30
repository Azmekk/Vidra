using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Vidra.Backend.Migrations
{
    /// <inheritdoc />
    public partial class VideoVariantRenameToFilePathMigration : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.RenameColumn(
                name: "Path",
                table: "VideoVariants",
                newName: "FilePath");
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.RenameColumn(
                name: "FilePath",
                table: "VideoVariants",
                newName: "Path");
        }
    }
}

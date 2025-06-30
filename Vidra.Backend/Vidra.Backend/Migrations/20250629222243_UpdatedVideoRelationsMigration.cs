using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Vidra.Backend.Migrations
{
    /// <inheritdoc />
    public partial class UpdatedVideoRelationsMigration : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.RenameColumn(
                name: "FormatId",
                table: "VideoVariants",
                newName: "Path");

            migrationBuilder.RenameColumn(
                name: "Thumbnail",
                table: "Videos",
                newName: "ThumbnailUrl");

            migrationBuilder.AddColumn<string>(
                name: "ExternalFormatId",
                table: "VideoVariants",
                type: "text",
                nullable: true);

            migrationBuilder.AddColumn<string>(
                name: "IsCustom",
                table: "VideoVariants",
                type: "text",
                nullable: false,
                defaultValue: "");

            migrationBuilder.AddColumn<bool>(
                name: "IsDefault",
                table: "VideoVariants",
                type: "boolean",
                nullable: false,
                defaultValue: false);

            migrationBuilder.AlterColumn<int>(
                name: "ViewCount",
                table: "Videos",
                type: "integer",
                nullable: true,
                oldClrType: typeof(int),
                oldType: "integer");

            migrationBuilder.AddColumn<string>(
                name: "ThumbnailPath",
                table: "Videos",
                type: "text",
                nullable: false,
                defaultValue: "");
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropColumn(
                name: "ExternalFormatId",
                table: "VideoVariants");

            migrationBuilder.DropColumn(
                name: "IsCustom",
                table: "VideoVariants");

            migrationBuilder.DropColumn(
                name: "IsDefault",
                table: "VideoVariants");

            migrationBuilder.DropColumn(
                name: "ThumbnailPath",
                table: "Videos");

            migrationBuilder.RenameColumn(
                name: "Path",
                table: "VideoVariants",
                newName: "FormatId");

            migrationBuilder.RenameColumn(
                name: "ThumbnailUrl",
                table: "Videos",
                newName: "Thumbnail");

            migrationBuilder.AlterColumn<int>(
                name: "ViewCount",
                table: "Videos",
                type: "integer",
                nullable: false,
                defaultValue: 0,
                oldClrType: typeof(int),
                oldType: "integer",
                oldNullable: true);
        }
    }
}

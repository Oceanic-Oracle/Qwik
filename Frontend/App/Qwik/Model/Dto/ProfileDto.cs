using System.Text.Json.Serialization;

namespace Qwik.Model.Dto
{
    public record class GetProfileRes
    {
        [JsonPropertyName("surname")]
        public string Surname { get; set; } = string.Empty;

        [JsonPropertyName("name")]
        public string Name { get; set; } = string.Empty;

        [JsonPropertyName("patronymic")]
        public string Patronymic { get; set; } = string.Empty;

        [JsonPropertyName("created_at")]
        public string CreatedAt { get; set; } = string.Empty;
    }
}

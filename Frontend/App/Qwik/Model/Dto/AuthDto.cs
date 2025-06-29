using System.Text.Json.Serialization;

namespace Qwik.Model.Dto
{
    public record AuthenticationReq
    {
        [JsonPropertyName("login")]
        public string Login { get; init; } = string.Empty;

        [JsonPropertyName("password")]
        public string Password { get; init; } = string.Empty;
    }

    public record AuthenticationRes
    {
        [JsonPropertyName("jwt")]
        public string Jwt { get; init; } = string.Empty;
    }
}

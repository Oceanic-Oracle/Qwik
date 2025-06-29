using System.Text.Json.Serialization;

namespace Qwik.Model.Dto
{
    public record class VerifyReq
    {
        [JsonPropertyName("email")]
        public string Email { get; set; } = string.Empty;
    }

    public record class VerifyRes
    {
        [JsonPropertyName("sessioncode")]
        public string Sessioncode { get; set; } = string.Empty;
    }

    public record class ConfirmRegReq
    {
        [JsonPropertyName("sessioncode")]
        public string SessionCode { get; set; } = string.Empty;

        [JsonPropertyName("email")]
        public string Email { get; set; } = string.Empty;

        [JsonPropertyName("login")]
        public string Login { get; set; } = string.Empty;

        [JsonPropertyName("password")]
        public string Password { get; set; } = string.Empty;

        [JsonPropertyName("verifycode")]
        public string VerifyCode { get; set; } = string.Empty;
    }

    public record class ConfirmRegRes
    {
        [JsonPropertyName("id")]
        public string Id { get; set; } = string.Empty;
    }
}

using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;

namespace Qwik.Model
{
    internal static class Api
    {
        public const string AuthenticationEndpoint = "user/v1/users/login";
        public const string VerifyCodeEndpoint = "user/v1/users/verify";
        public const string ConfirmRegEndpoint = "user/v1/users/registration";
        public static async Task<resDto> SendPostRequest<reqDto, resDto>(reqDto req, string enpoint)
        {
            try
            {
                string json = JsonSerializer.Serialize(req);
                var content = new StringContent(json, Encoding.UTF8, "application/json");

                var body = await Config.httpClient.PostAsync(enpoint, content);
                body.EnsureSuccessStatusCode();

                var responseBody = await body.Content.ReadAsStringAsync();

                return JsonSerializer.Deserialize<resDto>(responseBody)
                    ?? throw new InvalidOperationException("Invalid response from server");
            }
            catch (HttpRequestException ex)
            {
                System.Diagnostics.Debug.WriteLine(ex.Message);
                throw;
            }
            catch (Exception ex)
            {
                System.Diagnostics.Debug.WriteLine(ex.Message);
                throw;
            }
        }
    }
}

using System.Text;
using System.Text.Json;

namespace Qwik.Model
{
    internal static class Api
    {
        public const string AuthenticationEndpoint = "user/v1/users/login";
        public const string VerifyCodeEndpoint = "user/v1/users/verify";
        public const string ConfirmRegEndpoint = "user/v1/users/registration";
        public const string MyProfileEndpoint = "user/v1/profile/me";

        public static async Task<resDto> SendPostRequest<reqDto, resDto>(reqDto req, string endpoint)
        {
            try
            {
                var request = new HttpRequestMessage(HttpMethod.Post, endpoint);

                string json = JsonSerializer.Serialize(req);
                request.Content = new StringContent(json, Encoding.UTF8, "application/json");

                var body = await Config.httpClient.SendAsync(request);
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

        public static async Task<resDto> SendGetRequest<resDto>(Dictionary<string, string> header, string endpoint)
        {
            try
            {
                var request = new HttpRequestMessage(HttpMethod.Get, endpoint);
                
                foreach (var key in header)
                {
                    request.Headers.Add(key.Key, key.Value);
                }

                var body = await Config.httpClient.SendAsync(request);
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

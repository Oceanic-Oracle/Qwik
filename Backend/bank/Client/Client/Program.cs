using Grpc.Net.Client;
using bankapi.Protos;
using System.Net.Http;

await Task.Delay(5000); /// запуск сервера

var handler = new HttpClientHandler();
handler.ServerCertificateCustomValidationCallback =
    HttpClientHandler.DangerousAcceptAnyServerCertificateValidator;

var channel = GrpcChannel.ForAddress("https://bankapi:8080", new GrpcChannelOptions
{
    HttpHandler = handler
});

var client = new Auth.AuthClient(channel);

var req = new LoginRequest
{
    Username = "admin",
    Password = "admin"
};

try
{
    var response = await client.LoginAsync(req);
    Console.WriteLine($"Success: {response.Success}");
    Console.WriteLine($"Message: {response.Message}");  
    Console.ReadLine();// чтобы не завершалось
}
catch (Exception ex)
{
    Console.WriteLine($"Ошибка: {ex}");   
    Console.ReadLine();
}
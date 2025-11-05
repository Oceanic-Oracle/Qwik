using bankapi.Services;
using Microsoft.AspNetCore.Server.Kestrel.Core;
using Microsoft.AspNetCore.Server.Kestrel.Https;
using System.Security.Cryptography;
using System.Security.Cryptography.X509Certificates;

var builder = WebApplication.CreateBuilder(args);

builder.WebHost.ConfigureKestrel(options =>
{
    options.ListenAnyIP(8080, listenOptions =>
    {
        listenOptions.Protocols = HttpProtocols.Http2; 
        listenOptions.UseHttps(ConfigureHttps);
    });
});

builder.Services.AddGrpc();

var app = builder.Build();
app.MapGrpcService<AuthService>();
app.Run();

static void ConfigureHttps(HttpsConnectionAdapterOptions options)
{
 
    var cert = CreateDevelopmentCertificate();
    options.ServerCertificate = cert;
}

static X509Certificate2 CreateDevelopmentCertificate()
{
    using var rsa = RSA.Create(2048);
    var certRequest = new CertificateRequest(
        "CN=localhost",
        rsa,
        HashAlgorithmName.SHA256,
        RSASignaturePadding.Pkcs1);

    certRequest.CertificateExtensions.Add(
        new X509KeyUsageExtension(X509KeyUsageFlags.DigitalSignature, true));

    return certRequest.CreateSelfSigned(DateTimeOffset.Now, DateTimeOffset.Now.AddYears(1));
}
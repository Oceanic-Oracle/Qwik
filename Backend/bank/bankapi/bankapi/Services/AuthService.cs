using Grpc.Core;
using bankapi.Protos;
namespace bankapi.Services
   
{
    public class AuthService : Auth.AuthBase
    {
        private readonly ILogger<AuthService> _logger;

        public AuthService(ILogger<AuthService> logger)
        {
            _logger = logger;
        }

        public override Task<LoginResponse> Login(LoginRequest request, ServerCallContext context)
        {

            bool Validation = Check(request.Username, request.Password);
            var Response = new LoginResponse();
            Response.Success = Validation;
            if (Validation == true)
            {
                Response.Message = "Auth true";
            }
            else
                Response.Message = "Auth false";
            return Task.FromResult(Response);
            
        }

        private bool Check(string username, string password)// заглушка
        {
            if(username== "admin" && password =="admin" )
            {
                return true;
            }
            else
                return false;
        }

    }
}

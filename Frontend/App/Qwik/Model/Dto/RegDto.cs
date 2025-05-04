using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Qwik.Model.Dto
{
    public record class VerifyReq
    {
        public string email { get; set; } = string.Empty;
    }

    public record class VerifyRes
    {
        public string sessioncode { get; set; } = string.Empty;
    }

    public record class ConfirmRegReq
    {
        public string sessionCode { get; set; } = string.Empty;
        public string email { get; set; } = string.Empty;
        public string login { get; set; } = string.Empty;
        public string password { get; set; } = string.Empty;
        public string verifyCode { get; set; } = string.Empty;
    }

    public record class ConfirmRegRes
    {
        public string id { get; set; } = string.Empty;
    }
}

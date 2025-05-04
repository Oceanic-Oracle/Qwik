using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Qwik.Model.Dto
{
    public record AuthenticationReq
    {
        public string login { get; init; } = string.Empty;
        public string password { get; init; } = string.Empty;
    }

    public record AuthenticationRes
    {
        public string jwt { get; init; } = string.Empty;
    }
}

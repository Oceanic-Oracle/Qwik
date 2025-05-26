using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace DbNomad.Storage
{
    internal interface IStorage : IDisposable
    {
        public void Ping();
        public void InitializeHistoryTable();
        public void Migrate();
    }
}

using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace DbNomad.Utils
{
    static class FileReader
    {
        static public string ReadFile(string path)
        {
            try
            {
                string data = File.ReadAllText(path);
                
                return data;
            }
            catch (Exception)
            {
                throw;
            }
        }

        static public string[] ReadDirectory(string path) 
        {
            if (Directory.Exists("sql"))
            {
                string[] AllFiles = Directory.GetFiles("sql")
                    .Select(Path.GetFileName)
                    .Where(fileName => fileName != null)
                    .Select(fileName => fileName!)
                    .OrderBy(fileName => fileName)
                    .ToArray();
                return AllFiles;
            }
            else
            {
                throw new Exception("Отсутствует директория");
            }
        }
    }
}

using System.Security.Cryptography;
using System.Text;
using DbNomad.Utils;
using Npgsql;

namespace DbNomad.Storage.Postgres
{
    internal class Postgres : IStorage, IDisposable
    {
        NpgsqlConnection[] sqlConnection;
        public Postgres(string? urls)
        {
            if (urls == null) throw new ArgumentNullException(nameof(urls));
            string[] arrUrls = urls.Split(",", StringSplitOptions.RemoveEmptyEntries);
            sqlConnection = new NpgsqlConnection[arrUrls.Length];

            for (var i = 0; i < sqlConnection.Length; i++)
            {
                sqlConnection[i] = new NpgsqlConnection(Parse(arrUrls[i]));
                sqlConnection[i].Open();
            }
        }

        public void Ping()
        {
            try
            {
                foreach (var conn in sqlConnection)
                {
                    using (var cmd = new NpgsqlCommand("SELECT 1", conn))
                    {
                        using (var reader = cmd.ExecuteReader())
                        {
                            if (reader.HasRows)
                            {
                                Console.WriteLine("DEBUG: Подключение к {0} активно", conn.Host);
                            }
                            else
                            {
                                throw new InvalidOperationException($"FATAL: Не удалось проверить подключение к {conn.Host}");
                            }
                        }
                    }
                }
            }
            catch (Exception)
            {
                throw;
            }
        }

        public void InitializeHistoryTable()
        {
            try
            {
                string sql = FileReader.ReadFile("InitHistoryTable.sql");
                Console.WriteLine("DEBUG: " + sql);

                foreach (var conn in sqlConnection)
                {
                    try
                    {
                        using (var command = new NpgsqlCommand())
                        {
                            command.Connection = conn;
                            command.CommandText = sql;
                            command.ExecuteNonQuery();
                        }
                    }
                    catch (PostgresException ex) when (ex.SqlState == "42P07")
                    {
                        Console.WriteLine($"NOTICE: Таблица уже существует на {conn.Host}");
                    }
                    catch (Exception)
                    {
                        throw;
                    }
                }
            }
            catch (Exception)
            {
                throw;
            }
        }

        public void Migrate()
        {   
            try
            {
                string[] files = FileReader.ReadDirectory("sql");
                Console.WriteLine($"DEBUG: Найдено {files.Length} SQL-файлов");

                foreach (var file in files)
                {
                    string sql = FileReader.ReadFile(Path.Combine("sql", file));
                    Console.WriteLine(file + "\n" + sql);

                    foreach (var conn in sqlConnection)
                    {
                        try
                        {
                            Console.WriteLine($"DEBUG: Выполнение на {conn.Host}");

                            string hash = CalculateFileHash(sql);

                            using (var transaction = conn.BeginTransaction())
                            using (NpgsqlCommand command = new NpgsqlCommand())
                            {
                                command.Connection = conn;
                                command.Transaction = transaction;

                                command.CommandText = @"INSERT INTO __MigrationsHistory (fileName, hash)
                                                VALUES (@fileName, @hash);";
                                command.Parameters.AddWithValue("@fileName", file);
                                command.Parameters.AddWithValue("@hash", hash);
                                command.ExecuteNonQuery();

                                command.Parameters.Clear();

                                command.CommandText = sql;
                                command.ExecuteNonQuery();

                                transaction.Commit();
                            }

                            Console.WriteLine($"DEBUG: Успешно выполнено на {conn.Host}");
                        }
                        catch (Exception ex)
                        {
                            Console.WriteLine($"FATAL: Ошибка на {conn.Host}: {ex.Message}");
                            throw;
                        }
                    }
                }
            }
            catch (Exception)
            {
                throw;
            }  
        }

        public void Dispose()
        {
            foreach (var conn in sqlConnection)
            {
                conn.Close();
            }
        }

        private string CalculateFileHash(string content)
        {
            using (var sha = SHA256.Create())
            {
                byte[] hashBytes = sha.ComputeHash(Encoding.UTF8.GetBytes(content));
                return BitConverter.ToString(hashBytes).Replace("-", "").ToLower();
            }
        }

        private string Parse(string uri)
        {
            var pgUri = new Uri(uri);
            var builder = new NpgsqlConnectionStringBuilder
            {
                Host = pgUri.Host,
                Port = pgUri.Port == -1 ? 5432 : pgUri.Port,
                Username = pgUri.UserInfo.Split(':')[0],
                Password = pgUri.UserInfo.Split(':')[1],
                Database = pgUri.AbsolutePath.TrimStart('/')
            };

            var query = System.Web.HttpUtility.ParseQueryString(pgUri.Query);
            foreach (string key in query)
            {
                builder[key] = query[key];
            }

            return builder.ConnectionString;
        }
    }
}

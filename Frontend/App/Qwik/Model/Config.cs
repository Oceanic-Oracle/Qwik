using System.Text.Json;

namespace Qwik
{
    public class Config
    {
        private static readonly string _filePath = Path.Combine(FileSystem.AppDataDirectory, "userdata.json");
        static public HttpClient httpClient = new HttpClient { BaseAddress = new Uri("http://192.168.0.69:80/") };

        private record class StorageData
        {
            public string? JWT { get; set; }
        }

        private static StorageData storage = new StorageData();

        public static string? JWT
        {
            get => storage.JWT;
            set
            {
                storage.JWT = value;
                Save();
            }
        }

        private Config()
        {
            Load();
        }

        public static void Load()
        {
            try
            {
                if (File.Exists(_filePath))
                {
                    string json = File.ReadAllText(_filePath);
                    var loaded = JsonSerializer.Deserialize<StorageData>(json);
                    if (loaded != null)
                    {
                        storage = loaded;
                    }
                }
            }
            catch(Exception ex)
            {
                Console.WriteLine($"Ошибка загрузки конфига: {ex.Message}");
            }
        }

        public static void Save()
        {
            try
            {
                string json = JsonSerializer.Serialize(storage);
                File.WriteAllText(_filePath, json);
            }
            catch (Exception ex) 
            {
                Console.WriteLine($"Ошибка сохранения конфига {ex.Message}");
            }
        }
    }
}

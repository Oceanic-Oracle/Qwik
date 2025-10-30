using DbNomad.Storage;
using DbNomad.Storage.Postgres;

try
{
    string? typeDb = Environment.GetEnvironmentVariable("TYPE_DB");
    string? urls = Environment.GetEnvironmentVariable("URLS");

    Console.WriteLine($"DEBUG: TYPE_DB = {typeDb}");
    Console.WriteLine($"DEBUG: URLS = {urls}");

    using (IStorage storage = CreateStorage(typeDb, urls))
    {
        storage.Ping();
        storage.InitializeHistoryTable();
        storage.Migrate();
    }

    Environment.Exit(0);
}
catch (ArgumentNullException ex)
{
    Console.WriteLine($"FATAL: {ex.Message}");
}
catch (Exception ex)
{
    Console.WriteLine($"FATAL: {ex.Message}");
}

IStorage CreateStorage(string? typeDb, string? urls)
{
    return typeDb switch
    {
        "postgres" => new Postgres(urls),
        "none" => throw new InvalidOperationException("Укажите тип базы данных"),
        _ => throw new InvalidOperationException($"Несуществующий тип базы данных: {typeDb}")
    };
}

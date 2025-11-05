FROM mcr.microsoft.com/dotnet/sdk:9.0 AS build
WORKDIR /src


COPY Client/Client/Client.csproj ./Client/Client/
RUN dotnet restore Client/Client/Client.csproj


COPY proto/ ./proto/


COPY Client/Client/ ./Client/Client/
WORKDIR /src/Client/Client
RUN dotnet publish -c Release -o /app/publish

FROM mcr.microsoft.com/dotnet/runtime:9.0 AS final
WORKDIR /app
COPY --from=build /app/publish .
ENTRYPOINT ["dotnet", "Client.dll"]
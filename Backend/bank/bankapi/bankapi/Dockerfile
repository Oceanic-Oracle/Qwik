FROM mcr.microsoft.com/dotnet/sdk:9.0 AS build
WORKDIR /src


COPY proto/ ./proto/


COPY bankapi/bankapi/bankapi.csproj ./bankapi/bankapi/
RUN dotnet restore bankapi/bankapi/bankapi.csproj


COPY bankapi/bankapi/ ./bankapi/bankapi/
WORKDIR /src/bankapi/bankapi
RUN dotnet publish -c Release -o /app/publish

FROM mcr.microsoft.com/dotnet/aspnet:9.0 AS final
WORKDIR /app
COPY --from=build /app/publish .
EXPOSE 8080
ENTRYPOINT ["dotnet", "bankapi.dll"]
# Usamos a imagem do .NET 9 (sua versão) para o ambiente de execução.
FROM mcr.microsoft.com/dotnet/aspnet:9.0 AS base
WORKDIR /app
EXPOSE 8080

# Usamos a imagem do SDK do .NET 9 para o ambiente de build.
FROM mcr.microsoft.com/dotnet/sdk:9.0 AS build
WORKDIR /src
COPY ["Polls.API.csproj", "./"]
RUN dotnet restore "./Polls.API.csproj"
COPY . .
WORKDIR "/src/."
RUN dotnet build "Polls.API.csproj" -c Release -o /app/build

FROM build AS publish
RUN dotnet publish "Polls.API.csproj" -c Release -o /app/publish /p:UseAppHost=false

# Estágio final: copiamos os arquivos publicados para a imagem base.
FROM base AS final
WORKDIR /app
COPY --from=publish /app/publish .
ENTRYPOINT ["dotnet", "Polls.API.dll"]
// src/results-api/Program.cs
using MassTransit;
using Microsoft.AspNetCore.Http;
using Results.Api;
using StackExchange.Redis;
using RabbitMQ.Client;
using Microsoft.AspNetCore.SignalR;
using Prometheus;

// Variável para o nome da nossa política de CORS
var MyAllowSpecificOrigins = "_myAllowSpecificOrigins";

var builder = WebApplication.CreateBuilder(args);

// --- Configuração dos Serviços ---
builder.Services.AddCors(options =>
{
    options.AddPolicy(name: MyAllowSpecificOrigins,
        policy =>
        {
            policy.WithOrigins("http://localhost:8000") // Permite requisições da origem do nosso frontend
                  .AllowAnyHeader()
                  .AllowAnyMethod()
                  .AllowCredentials(); // Crucial para SignalR com autenticação/sessões
        });
});


builder.Services.AddSingleton<IConnectionMultiplexer>(sp => 
    ConnectionMultiplexer.Connect(builder.Configuration.GetConnectionString("Redis")!));
    
builder.Services.AddSignalR();

builder.Services.AddMassTransit(x =>
{
    x.AddConsumer<VoteConsumer>();
    x.UsingRabbitMq((context, cfg) =>
    {
        cfg.Host(builder.Configuration.GetConnectionString("RabbitMQ"));
        cfg.ReceiveEndpoint("results_queue", e =>
        {
            e.Bind("votes_exchange", x =>
            {
                x.ExchangeType = ExchangeType.Fanout;
                x.Durable = true;
            });
            e.UseRawJsonDeserializer();
            e.ConfigureConsumer<VoteConsumer>(context);
        });
    });
});

var app = builder.Build();

app.UseHttpMetrics();

// --- Configuração do Pipeline de Middlewares ---
app.UseCors(MyAllowSpecificOrigins);

// --- Definição dos Endpoints ---
app.MapGet("/results/{pollId}", async (string pollId, IConnectionMultiplexer redis) =>
{
    var db = redis.GetDatabase();
    var key = $"poll:{pollId}";
    var hashEntries = await db.HashGetAllAsync(key);
    if (hashEntries.Length == 0)
    {
        return Microsoft.AspNetCore.Http.Results.NotFound("Nenhum voto encontrado para esta enquete.");
    }
    var results = hashEntries.ToDictionary(
        entry => entry.Name.ToString(),
        entry => (int)entry.Value
    );
    return Microsoft.AspNetCore.Http.Results.Ok(results);
});

app.MapHub<ResultsHub>("/results-hub");

app.MapMetrics();

app.Run();
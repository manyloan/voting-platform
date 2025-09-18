using Microsoft.EntityFrameworkCore;
using Polls.Api;
using Prometheus;

var builder = WebApplication.CreateBuilder(args);

// --- Configuração dos Serviços ---

builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen();

var connectionString = builder.Configuration.GetConnectionString("DefaultConnection");

builder.Services.AddDbContext<PollsDbContext>(options =>
    options.UseNpgsql(connectionString));

// --- Construção da Aplicação ---
var app = builder.Build();

app.UseHttpMetrics();

// --- Configuração do Pipeline de Middlewares ---
if (app.Environment.IsDevelopment())
{
    app.UseSwagger();
    app.UseSwaggerUI();
}

using (var scope = app.Services.CreateScope())
{
    var db = scope.ServiceProvider.GetRequiredService<PollsDbContext>();
    db.Database.Migrate();
}

// --- Definição dos Endpoints ---
app.MapPost("/polls", async (Poll poll, PollsDbContext dbContext) =>
{
    dbContext.Polls.Add(poll);
    await dbContext.SaveChangesAsync();
    return Results.Created($"/polls/{poll.Id}", poll);
});

app.MapGet("/polls", async (PollsDbContext dbContext) =>
{
    var polls = await dbContext.Polls.ToListAsync();
    return Results.Ok(polls);
});

app.MapGet("/polls/{id}", async (Guid id, PollsDbContext dbContext) =>
{
    var poll = await dbContext.Polls.FindAsync(id);

    return poll is not null ? Results.Ok(poll) : Results.NotFound();
});

app.MapMetrics();


// --- Execução da Aplicação ---
app.Run();
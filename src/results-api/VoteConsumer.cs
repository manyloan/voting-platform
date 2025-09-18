using MassTransit;
using Microsoft.AspNetCore.SignalR;
using Microsoft.Extensions.Logging;
using StackExchange.Redis;

namespace Results.Api;

public class VoteConsumer(
    ILogger<VoteConsumer> logger, 
    IConnectionMultiplexer redis,
    IHubContext<ResultsHub> hubContext)
    : IConsumer<VoteContract>
{
    public async Task Consume(ConsumeContext<VoteContract> context)
    {
        var vote = context.Message;
        logger.LogInformation("Recebido voto para a enquete {PollId}, opção {OptionId}", vote.PollId, vote.OptionId);

        var db = redis.GetDatabase();
        var key = $"poll:{vote.PollId}";
        
        // 1. Incrementa o contador no Redis
        await db.HashIncrementAsync(key, vote.OptionId, 1);
        
        // 2. Lê todos os totais atualizados para esta enquete
        var hashEntries = await db.HashGetAllAsync(key);
        var results = hashEntries.ToDictionary(
            entry => entry.Name.ToString(),
            entry => (int)entry.Value
        );

        // 3. Envia os totais para o grupo de clientes interessados naquela enquete
        await hubContext.Clients.Group(vote.PollId).SendAsync("UpdateResults", results);
    }
}
using Microsoft.AspNetCore.SignalR;

namespace Results.Api;

public class ResultsHub : Hub
{
    // Cliente (navegador) vai chamar este método para se inscrever em atualizações de uma enquete específica.
    public async Task JoinPollGroup(string pollId)
    {
        // Adiciona a conexão atual a um grupo nomeado com o ID da enquete.
        // Assim, podemos enviar atualizações apenas para quem está interessado nessa enquete.
        await Groups.AddToGroupAsync(Context.ConnectionId, pollId);
    }
}
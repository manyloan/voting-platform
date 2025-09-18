namespace Polls.Api;

public class Poll
{
    public Guid Id { get; set; }
    public string Question { get; set; } = string.Empty;
}
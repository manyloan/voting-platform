using Microsoft.EntityFrameworkCore;

namespace Polls.Api;

public class PollsDbContext : DbContext
{
    public PollsDbContext(DbContextOptions<PollsDbContext> options) : base(options) { }
    
    public DbSet<Poll> Polls { get; set; }
}
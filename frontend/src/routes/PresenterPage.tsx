import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { apiPollResults } from '../api/polls';
import { apiListPolls } from '../api/polls';

function optionLabelFromPoll(poll: any, optionId: string): string {
  const opt = poll?.options?.find((o: any) => o.id === optionId);
  return opt?.label ?? optionId.slice(0, 8);
}

export function PresenterPage() {
  const { id } = useParams<{ id: string }>();

  const { data: results, isLoading } = useQuery({
    queryKey: ['poll', id, 'presenter'],
    queryFn: ({ signal }) => apiPollResults(id!, signal),
    enabled: !!id,
    refetchInterval: 2000,
  });

  const pollId = results?.pollId;
  const { data: polls = [] } = useQuery({
    queryKey: ['presenter-poll-list'],
    queryFn: ({ signal }) => apiListPolls('', signal),
    enabled: false,
  });

  const poll = polls.find((p) => p.id === pollId);

  if (isLoading) return <div className="presenter"><div>Loading…</div></div>;
  if (!results || !id) return <div className="presenter"><div>Poll not found</div></div>;

  const total = results.total;
  const entries = Object.entries(results.counts).sort((a, b) => b[1] - a[1]);
  const max = Math.max(1, ...entries.map(([, n]) => n));

  return (
    <div className="presenter">
      <Link to={`/event/${poll?.eventId ?? ''}`} className="presenter-exit">×</Link>
      <header className="presenter-head">
        <div className="presenter-question">{poll?.question ?? 'Live results'}</div>
        <div className={`presenter-status ${results.status.toLowerCase()}`}>
          {results.status}
        </div>
      </header>

      <div className="presenter-body">
        {entries.length === 0 && <div className="muted">No votes yet…</div>}
        {entries.map(([optId, count]) => {
          const pct = total > 0 ? Math.round((count / total) * 100) : 0;
          const height = Math.round((count / max) * 100);
          return (
            <div key={optId} className="presenter-row">
              <div className="presenter-label">
                {optionLabelFromPoll(poll, optId)}
              </div>
              <div className="presenter-bar-track">
                <div className="presenter-bar-fill" style={{ width: `${height}%` }} />
              </div>
              <div className="presenter-count">{count} · {pct}%</div>
            </div>
          );
        })}
      </div>

      <footer className="presenter-foot">
        {total} vote{total === 1 ? '' : 's'} · auto-refreshing
      </footer>
    </div>
  );
}
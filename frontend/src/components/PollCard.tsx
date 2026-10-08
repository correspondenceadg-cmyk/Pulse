import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { apiPollResults, apiVotePoll } from '../api/polls';
import type { Poll } from '../api/types';

interface Props {
  poll: Poll;
  isOwner: boolean;
}

export function PollCard({ poll, isOwner }: Props) {
  const qc = useQueryClient();

  const { data: results } = useQuery({
    queryKey: ['poll', poll.id, 'results'],
    queryFn: ({ signal }) => apiPollResults(poll.id, signal),
    refetchInterval: poll.status === 'OPEN' ? 3000 : false,
  });

  const vote = useMutation({
    mutationFn: (optionId: string) => apiVotePoll(poll.id, optionId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['poll', poll.id, 'results'] });
    },
  });

  if (!poll.options || poll.options.length === 0) {
    return null;
  }

  const total = results?.total ?? 0;
  const counts = results?.counts ?? {};
  const userOption = results?.userOptionId;

  return (
    <div className="poll-card">
      <header className="poll-head">
        <h3>{poll.question}</h3>
        <span className={`poll-status ${poll.status.toLowerCase()}`}>{poll.status}</span>
      </header>

      <div className="poll-options">
        {poll.options.map((opt) => {
          const c = counts[opt.id] ?? 0;
          const pct = total > 0 ? Math.round((c / total) * 100) : 0;
          const chosen = userOption === opt.id;
          const disabled = poll.status !== 'OPEN' || !!userOption || vote.isPending;

          return (
            <button
              key={opt.id}
              className={chosen ? 'poll-option chosen' : 'poll-option'}
              onClick={() => vote.mutate(opt.id)}
              disabled={disabled}
            >
              <div className="poll-bar" style={{ width: `${pct}%` }} />
              <span className="poll-label">{opt.label}</span>
              <span className="poll-count">
                {c} · {pct}%
              </span>
            </button>
          );
        })}
      </div>

      <footer className="poll-foot">
        <span className="muted">{total} vote{total === 1 ? '' : 's'}</span>
        {isOwner && (
          <Link to={`/poll/${poll.id}/present`} className="presenter-link">
            Present →
          </Link>
        )}
      </footer>
    </div>
  );
}
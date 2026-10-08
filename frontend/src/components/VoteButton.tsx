import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiVote } from '../api/events';

interface Props {
  eventId: string;
  up: number;
  down: number;
  score: number;
  userVote?: number;
}

export function VoteButton({ eventId, up, down, score, userVote }: Props) {
  const qc = useQueryClient();

  const m = useMutation({
    mutationFn: (value: -1 | 0 | 1) => apiVote(eventId, value),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['event', eventId, 'stats'] });
      qc.invalidateQueries({ queryKey: ['events'] });
    },
  });

  const send = (value: -1 | 1) => {
    if (userVote === value) {
      m.mutate(0);
    } else {
      m.mutate(value);
    }
  };

  return (
    <div className="vote-row">
      <button
        className={userVote === 1 ? 'vote-btn active up' : 'vote-btn up'}
        onClick={() => send(1)}
        disabled={m.isPending}
        aria-label="Upvote"
      >
        ▲ {up}
      </button>
      <div className="vote-score">{score}</div>
      <button
        className={userVote === -1 ? 'vote-btn active down' : 'vote-btn down'}
        onClick={() => send(-1)}
        disabled={m.isPending}
        aria-label="Downvote"
      >
        ▼ {down}
      </button>
    </div>
  );
}
import { useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { useAuth } from '../auth/useAuth';
import { apiGetEvent, apiStats } from '../api/events';
import { apiListPolls, apiClosePoll, apiOpenPoll } from '../api/polls';
import { VoteButton } from '../components/VoteButton';
import { PollCard } from '../components/PollCard';
import { PollCreateDialog } from '../components/PollCreateDialog';

function formatWhen(iso: string): string {
  try {
    return new Date(iso).toLocaleString(undefined, {
      weekday: 'short',
      month: 'short',
      day: 'numeric',
      hour: 'numeric',
      minute: '2-digit',
    });
  } catch {
    return iso;
  }
}

export function EventPage() {
  const { id } = useParams<{ id: string }>();
  const { user } = useAuth();
  const navigate = useNavigate();
  const [creating, setCreating] = useState(false);

  const { data: event, isLoading, error } = useQuery({
    queryKey: ['event', id],
    queryFn: ({ signal }) => apiGetEvent(id!, signal),
    enabled: !!id,
  });

  const { data: stats } = useQuery({
    queryKey: ['event', id, 'stats'],
    queryFn: ({ signal }) => apiStats(id!, signal),
    enabled: !!id,
  });

  const { data: polls = [] } = useQuery({
    queryKey: ['event', id, 'polls'],
    queryFn: ({ signal }) => apiListPolls(id!, signal),
    enabled: !!id,
  });

  if (isLoading) return <div className="page-loading">Loading…</div>;
  if (error || !event) {
    return (
      <div className="page">
        <header className="topbar">
          <button onClick={() => navigate('/')}>← Map</button>
          <h1>Not found</h1>
        </header>
        <main className="page-body">This event doesn't exist.</main>
      </div>
    );
  }

  const isOwner = user?.id === event.ownerId;

  return (
    <div className="page">
      <header className="topbar">
        <Link to="/" className="back-link">← Map</Link>
        <h1>{event.title}</h1>
        <div className="topbar-actions">
          <span className="muted">{user?.email}</span>
        </div>
      </header>

      <main className="page-body event-body">
        <div className="event-meta">
          <div className="event-when">{formatWhen(event.startsAt)}</div>
          {event.venue && <div className="event-venue">{event.venue}</div>}
          {event.address && <div className="muted">{event.address}</div>}
        </div>

        {event.description && <p className="event-desc">{event.description}</p>}

        <div className="event-vote">
          <VoteButton
            eventId={event.id}
            up={stats?.up ?? 0}
            down={stats?.down ?? 0}
            score={stats?.score ?? 0}
            userVote={stats?.userVote}
          />
        </div>

        <section className="polls-section">
          <header className="section-head">
            <h2>Polls</h2>
            {isOwner && (
              <button className="new-poll-btn" onClick={() => setCreating(true)}>
                + New poll
              </button>
            )}
          </header>

          {polls.length === 0 && (
            <p className="muted">No polls yet.</p>
          )}

          {polls.map((p) => (
            <div key={p.id}>
              <PollCard poll={p} isOwner={isOwner} />
              {isOwner && (
                <div className="poll-owner-actions">
                  {p.status === 'DRAFT' && (
                    <button onClick={() => apiOpenPoll(p.id).then(() => location.reload())}>
                      Open
                    </button>
                  )}
                  {p.status === 'OPEN' && (
                    <button onClick={() => apiClosePoll(p.id).then(() => location.reload())}>
                      Close
                    </button>
                  )}
                </div>
              )}
            </div>
          ))}
        </section>
      </main>

      {creating && id && <PollCreateDialog eventId={id} onClose={() => setCreating(false)} />}
    </div>
  );
}
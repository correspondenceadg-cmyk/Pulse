import { FormEvent, useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiCreatePoll } from '../api/polls';

interface Props {
  eventId: string;
  onClose: () => void;
}

export function PollCreateDialog({ eventId, onClose }: Props) {
  const qc = useQueryClient();
  const [question, setQuestion] = useState('');
  const [options, setOptions] = useState<string[]>(['', '']);
  const [error, setError] = useState<string | null>(null);

  const m = useMutation({
    mutationFn: () =>
      apiCreatePoll(
        eventId,
        question.trim(),
        options.map((o) => o.trim()).filter(Boolean),
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['event', eventId, 'polls'] });
      onClose();
    },
    onError: (err) => setError(err instanceof Error ? err.message : 'Failed'),
  });

  function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    const clean = options.map((o) => o.trim()).filter(Boolean);
    if (question.trim().length === 0) return setError('Question is required');
    if (clean.length < 2) return setError('At least 2 options');
    m.mutate();
  }

  function setOption(i: number, v: string) {
    setOptions((prev) => prev.map((o, idx) => (idx === i ? v : o)));
  }

  function addOption() {
    if (options.length < 10) setOptions((prev) => [...prev, '']);
  }

  function removeOption(i: number) {
    if (options.length <= 2) return;
    setOptions((prev) => prev.filter((_, idx) => idx !== i));
  }

  return (
    <div className="dialog-backdrop" onClick={onClose}>
      <form className="dialog" onClick={(e) => e.stopPropagation()} onSubmit={onSubmit}>
        <h2>New poll</h2>

        <label>
          <span>Question</span>
          <input
            type="text"
            value={question}
            onChange={(e) => setQuestion(e.target.value)}
            maxLength={300}
            autoFocus
          />
        </label>

        <div className="poll-options-edit">
          {options.map((o, i) => (
            <div key={i} className="poll-option-row">
              <input
                type="text"
                value={o}
                onChange={(e) => setOption(i, e.target.value)}
                placeholder={`Option ${i + 1}`}
                maxLength={120}
              />
              {options.length > 2 && (
                <button type="button" onClick={() => removeOption(i)} aria-label="Remove">
                  ×
                </button>
              )}
            </div>
          ))}
          {options.length < 10 && (
            <button type="button" className="add-option" onClick={addOption}>
              + Add option
            </button>
          )}
        </div>

        {error && <div className="form-error">{error}</div>}

        <div className="dialog-actions">
          <button type="button" onClick={onClose}>
            Cancel
          </button>
          <button type="submit" disabled={m.isPending}>
            {m.isPending ? 'Creating…' : 'Create'}
          </button>
        </div>
      </form>
    </div>
  );
}
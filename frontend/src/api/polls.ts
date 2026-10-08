import { request } from './client';
import type { Poll, PollResults } from './types';

export async function apiListPolls(eventId: string, signal?: AbortSignal): Promise<Poll[]> {
  const res = await request<{ items: Poll[] | null }>(`/api/events/${eventId}/polls`, { signal });
  return res.items ?? [];
}

export async function apiCreatePoll(
  eventId: string,
  question: string,
  options: string[],
): Promise<Poll> {
  return request<Poll>(`/api/events/${eventId}/polls`, {
    method: 'POST',
    body: { question, options },
  });
}

export async function apiOpenPoll(pollId: string): Promise<PollResults> {
  return request<PollResults>(`/api/polls/${pollId}/open`, { method: 'POST' });
}

export async function apiClosePoll(pollId: string): Promise<PollResults> {
  return request<PollResults>(`/api/polls/${pollId}/close`, { method: 'POST' });
}

export async function apiVotePoll(pollId: string, optionId: string): Promise<PollResults> {
  return request<PollResults>(`/api/polls/${pollId}/vote`, {
    method: 'POST',
    body: { optionId },
  });
}

export async function apiPollResults(pollId: string, signal?: AbortSignal): Promise<PollResults> {
  return request<PollResults>(`/api/polls/${pollId}/results`, { signal });
}
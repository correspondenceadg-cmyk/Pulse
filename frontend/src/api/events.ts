import { request } from './client';
import type { Event, EventStats } from './types';

interface ListResponse {
  items: Event[] | null;
}

export interface ListParams {
  bbox?: [number, number, number, number];
  when?: 'upcoming' | 'all';
  category?: string;
  limit?: number;
  cursor?: string;
}

export async function apiListEvents(
  params: ListParams = {},
  signal?: AbortSignal,
): Promise<Event[]> {
  const q = new URLSearchParams();
  if (params.bbox) q.set('bbox', params.bbox.join(','));
  if (params.when) q.set('when', params.when);
  if (params.category) q.set('category', params.category);
  if (params.limit) q.set('limit', String(params.limit));
  if (params.cursor) q.set('cursor', params.cursor);
  const qs = q.toString();
  const res = await request<ListResponse>(`/api/events/${qs ? `?${qs}` : ''}`, { signal });
  return res.items ?? [];
}

export async function apiGetEvent(id: string, signal?: AbortSignal): Promise<Event> {
  return request<Event>(`/api/events/${id}`, { signal });
}

export async function apiVote(id: string, value: -1 | 0 | 1): Promise<EventStats> {
  return request<EventStats>(`/api/events/${id}/vote`, {
    method: 'POST',
    body: { value },
  });
}

export async function apiStats(id: string, signal?: AbortSignal): Promise<EventStats> {
  return request<EventStats>(`/api/events/${id}/stats`, { signal });
}
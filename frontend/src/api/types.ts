export interface User {
  id: string;
  email: string;
  displayName: string;
  role: string;
}

export interface TokenResponse {
  accessToken: string;
  expiresIn: number;
  user: User;
}

export interface Event {
  id: string;
  ownerId: string;
  title: string;
  description: string;
  category: string;
  lat: number;
  lng: number;
  startsAt: string;
  endsAt?: string;
  venue: string;
  address: string;
  visibility: string;
  createdAt: string;
  updatedAt: string;
}

export interface EventStats {
  up: number;
  down: number;
  score: number;
  userVote?: number;
}

export interface PollOption {
  id: string;
  label: string;
  position: number;
}

export interface Poll {
  id: string;
  eventId: string;
  question: string;
  status: 'DRAFT' | 'OPEN' | 'CLOSED';
  createdAt: string;
  options?: PollOption[];
}

export interface PollResults {
  pollId: string;
  status: string;
  total: number;
  counts: Record<string, number>;
  userOptionId?: string;
}
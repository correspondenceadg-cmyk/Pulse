import { request } from './client';
import type { TokenResponse, User } from './types';

export async function apiRegister(email: string, password: string, displayName: string): Promise<User> {
  return request<User>('/api/auth/register', {
    method: 'POST',
    body: { email, password, displayName },
  });
}

export async function apiLogin(email: string, password: string): Promise<TokenResponse> {
  return request<TokenResponse>('/api/auth/login', {
    method: 'POST',
    body: { email, password },
  });
}

export async function apiRefresh(): Promise<TokenResponse> {
  return request<TokenResponse>('/api/auth/refresh', { method: 'POST' });
}

export async function apiLogout(): Promise<void> {
  return request<void>('/api/auth/logout', { method: 'POST' });
}
import { useAuth } from '../auth/AuthContext';

export function HomePage() {
  const { user, logout } = useAuth();
  return (
    <div className="page">
      <header className="topbar">
        <h1>Pulse</h1>
        <div className="topbar-actions">
          <span className="muted">{user?.email}</span>
          <button onClick={logout}>Log out</button>
        </div>
      </header>
      <main className="page-body">
        <p>Signed in. The map is next.</p>
      </main>
    </div>
  );
}
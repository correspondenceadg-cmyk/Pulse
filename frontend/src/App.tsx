import { Routes, Route, Link } from 'react-router-dom';

export default function App() {
  return (
    <div className="app">
      <header>
        <Link to="/">Pulse</Link>
      </header>
      <main>
        <Routes>
          <Route path="/" element={<div>Map coming soon.</div>} />
          <Route path="*" element={<div>Not found.</div>} />
        </Routes>
      </main>
    </div>
  );
}
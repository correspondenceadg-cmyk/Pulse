import { useEffect, useRef, useState } from 'react';
import maplibregl from 'maplibre-gl';
import 'maplibre-gl/dist/maplibre-gl.css';
import { useQuery } from '@tanstack/react-query';
import { useAuth } from '../auth/useAuth';
import { apiListEvents } from '../api/events';

export function MapPage() {
  const { user, logout } = useAuth();
  const containerRef = useRef<HTMLDivElement>(null);
  const mapRef = useRef<maplibregl.Map | null>(null);
  const [mapReady, setMapReady] = useState(false);

  const {
    data: events = [],
    isLoading,
    error,
  } = useQuery({
    queryKey: ['events', 'map'],
    queryFn: ({ signal }) => apiListEvents({ when: 'all', limit: 200 }, signal),
    staleTime: 30_000,
  });

  useEffect(() => {
    if (!containerRef.current || mapRef.current) return;

    const map = new maplibregl.Map({
      container: containerRef.current,
      style: 'https://demotiles.maplibre.org/style.json',
      center: [-98.5795, 39.8283],
      zoom: 3.5,
    });

    map.addControl(new maplibregl.NavigationControl(), 'top-right');
    map.on('load', () => setMapReady(true));
    mapRef.current = map;

    return () => {
      map.remove();
      mapRef.current = null;
      setMapReady(false);
    };
  }, []);

  useEffect(() => {
    if (!mapReady || !mapRef.current) return;
    const map = mapRef.current;
    const markers: maplibregl.Marker[] = [];

    for (const e of events) {
      const pin = document.createElement('button');
      pin.className = 'event-marker';
      pin.type = 'button';
      pin.setAttribute('aria-label', e.title);

      const content = document.createElement('div');
      content.className = 'event-popup';

      const title = document.createElement('strong');
      title.textContent = e.title;
      content.appendChild(title);

      if (e.venue) {
        content.appendChild(document.createElement('br'));
        const venue = document.createElement('span');
        venue.className = 'muted';
        venue.textContent = e.venue;
        content.appendChild(venue);
      }

      const popup = new maplibregl.Popup({ offset: 14, closeButton: true }).setDOMContent(content);

      const marker = new maplibregl.Marker({ element: pin })
        .setLngLat([e.lng, e.lat])
        .setPopup(popup)
        .addTo(map);

      markers.push(marker);
    }

    return () => {
      for (const m of markers) m.remove();
    };
  }, [events, mapReady]);

  return (
    <div className="map-page">
      <header className="topbar">
        <h1>Pulse</h1>
        <div className="topbar-actions">
          <span className="muted">{user?.email}</span>
          <button onClick={logout}>Log out</button>
        </div>
      </header>

      <div ref={containerRef} className="map-canvas" />

      {isLoading && <div className="map-overlay">Loading events…</div>}
      {error && <div className="map-overlay error">Failed to load events</div>}
      {!isLoading && !error && events.length === 0 && (
        <div className="map-overlay">
          No events yet. Add one from the API to see it here.
        </div>
      )}
    </div>
  );
}
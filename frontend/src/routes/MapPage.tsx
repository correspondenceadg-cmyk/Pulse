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

content.appendChild(document.createElement('br'));
const link = document.createElement('a');
link.href = `/event/${e.id}`;
link.className = 'event-popup-link';
link.textContent = 'View details →';
content.appendChild(link);
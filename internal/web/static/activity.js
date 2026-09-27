// Dates come from the server; the browser only supplies its calendar timezone.
(() => {
  try {
    const zone = Intl.DateTimeFormat().resolvedOptions().timeZone;
    if (!zone) return;
    const value = encodeURIComponent(zone);
    const previous = document.cookie.split('; ').find(c => c.startsWith('goland_timezone='));
    if (previous === `goland_timezone=${value}`) return;
    document.cookie = `goland_timezone=${value}; Path=/; Max-Age=31536000; SameSite=Lax`;
    // Refresh the calendar once after detection, only if cookies are accepted.
    if (document.querySelector('[data-activity]') &&
        document.cookie.split('; ').includes(`goland_timezone=${value}`)) {
      location.reload();
    }
  } catch {
    // UTC remains usable when timezone detection or cookies are unavailable.
  }
})();

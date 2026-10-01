# Original VoHive UI foundation

Version 0.1.1-personal-classic restores VoHive source at f240894e763cf7f0cf74c88562bb9b55f0d573b1 (giszh86/vohive), the HiDeck project's recorded original baseline.

The original 232px/52px sidebar, top bar, 767px mobile drawer, dashboard statistic cards and device-card grid, separate device header and tab card, SMS responsive panes, proxy tabs, login card, settings cards, and original global style sheet are restored. New call, task and command entries follow the existing sidebar items. The original light/dark preferences are retained and accidental navy values are migrated.

Upgraded auth, password management, device request isolation, SMS ICCID identity/read status/history/notification targets, and proxy configuration request guards remain in the personal edition. Telegram and QQ are the only notification settings displayed. Extra feature CSS variables inherit the original VoHive surfaces.

No database downgrade or overwrite is involved. Browser visual review was unavailable because browser access had been declined. Validation covers source layout comparison, Vue type checking, production compilation, functional unit tests, and router service/data checks; it does not claim pixel-by-pixel visual verification.

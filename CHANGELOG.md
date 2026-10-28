# Cambios

## 0.1.0 — 2026-10-28

Primera versión.

- Descargador de YouTube (video y audio) con cola, vista previa de playlists y ajustes de calidad y formato.
- Reproductor: búsqueda de video y música, mini-reproductor con cola, video y audio sincronizados, calidad
  automática a la baja si el equipo no da abasto, y descarga desde el reproductor.
- `yt-dlp`, `ffmpeg` y `deno` incluidos en los paquetes.
- Paquetes para Linux (`.deb`, `.rpm`) y Windows (instalador `.exe`).

Limitaciones conocidas: en Linux no se reproduce H.264 (se usa VP9/AV1); en Windows el reproductor no se ha
probado todavía.

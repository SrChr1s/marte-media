# Marte Desc

Descargador, conversor y reproductor de media de YouTube para escritorio (Linux y Windows).
Go + [Wails](https://wails.io) v2 + Svelte 5. Lleva `yt-dlp`, `ffmpeg` y `deno` incluidos.

## Qué hace
- **Reproducir**: busca video o música y reproduce sin descargar, con mini-reproductor, cola, enlaces y listas.
- **Descargar**: video o audio, listas completas, formatos y calidad a elegir.
- **Cola e Historial**: descargas concurrentes con progreso y cancelación.

## Desarrollo
```sh
scripts/fetch-deps.sh                      # yt-dlp, ffmpeg y deno en build/deps/
wails dev -tags webkit2_41                 # Linux
cd frontend && npm test                    # pruebas del frontend
go test -race ./internal/...               # pruebas del backend
go test -tags e2e ./internal/...           # contra YouTube real (necesita red)
```

## Paquetes
`scripts/package.sh linux` (`.deb` y `.rpm`, requiere nfpm) y `scripts/package.sh windows` (instalador, requiere makensis).

## Notas
- El reproductor sirve los flujos por un proxy local (`127.0.0.1`, secreto aleatorio). Las URL de YouTube tardan unos
  segundos en estar disponibles (`available_at`); el proxy espera.
- Estado y detalles del reproductor: `docs/reproductor.md`. Diseño: `DESIGN.md`. Licencias: `THIRD_PARTY.md`.

# Banco de pruebas del reproductor

Scripts para ejercitar la app real desde fuera (niri + `niri msg`), sin tocar el código de producción.

1. Crea la carpeta de trabajo (`S`, por defecto `/tmp/marte-testbed`) con `preal/` o `pdemo/`:
   `scenario.js` (copia de `scenario-real.js` / `scenario-offline.js`), `cfg/`, `shots/`, y `bin/fake-yt-dlp` (offline).
2. Compila una copia con el gancho de pruebas. En `main.go`, temporalmente, añade a `options.App`:
   `OnDomReady: func(ctx context.Context) { if js := os.Getenv("MARTE_JS1"); js != "" { runtime.WindowExecJS(ctx, js) } }`
   y `go build -tags desktop,production,webkit2_41 -o $S/realapp/marte-media .`; después restaura `main.go`.
3. Enlaza `yt-dlp`, `ffmpeg` y `deno` de `build/deps/linux-amd64/` junto a ese binario.
4. `S=... bash docs/player-testbed/run-real.sh` (red) o `run-offline.sh` (necesita `rangeserver.py` y medios de ffmpeg).

La app registra con `window.runtime.LogError('[real] ...')`; los marcadores `SHOT nombre` disparan capturas.

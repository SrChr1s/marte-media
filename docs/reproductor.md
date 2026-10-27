# Reproductor (buscar y reproducir video y música) — estado del trabajo

**Pausado el 27-oct-2026 a petición del usuario.** Este documento recoge lo hecho, lo medido, lo que falla y cómo
retomar. El plan aprobado está en `~/.claude/plans/vamos-a-hacer-un-foamy-phoenix.md`.

## Decisiones del usuario
- Reproducción por **stream directo propio**: yt-dlp resuelve los flujos y un servidor local en Go los sirve al
  `<audio>`/`<video>` (no iframe de YouTube, no descargar a caché).
- **Empaquetar deno** en los instaladores (yt-dlp necesita un motor JavaScript para YouTube).
- Alcance v1: búsqueda Video/Música, mini-reproductor con cola, descargar desde el reproductor, reproducir
  enlace o playlist pegados.
- **Reproducir es la primera vista** (primera en el rail y la que abre la app); Descargar queda debajo.

## Hecho

### Paso 0 — motor JavaScript de yt-dlp (también arregla el descargador)
- yt-dlp 2026.10.26 solo habilita `deno` por defecto; la app lo llamaba sin `--js-runtimes` (`JS runtimes: none`).
  Medido con YouTube real (Big Buck Bunny): **53 formatos / 10 pistas de audio sin motor JS, 61 / 14 con deno**.
- `internal/ytdlp/binary.go`: `Env{Bin, FFmpeg, JSArgs}`, `Binaries.Env()`, `JSRuntime()`; orden de búsqueda
  copia del usuario → junto al ejecutable → `PATH`. `Status` expone `jsRuntime`/`jsPath`. `Probe`, `Download` y la
  cola reciben `Env`.
- `scripts/fetch-deps.sh` baja **deno 2.9.8** (Linux y Windows) verificando su suma (el de Windows es salida de
  `Get-FileHash`); `DEPS=deno scripts/fetch-deps.sh` baja solo eso. `packaging/nfpm.yaml`, `project.nsi` y
  `THIRD_PARTY.md` actualizados. Coste: ~96 MB (Linux) / ~98 MB (Windows) sin comprimir.
- UI: fila «JavaScript» en Ajustes › Motor, aviso en Descargar si falta, el punto del rail exige yt-dlp+ffmpeg+JS.
- **Bug real encontrado y corregido** al probar una descarga por primera vez: `--print` activa `--quiet` y apagaba el
  progreso; la barra de la Cola nunca se habría movido. Ahora `BuildArgs` añade `--progress`.

### Paso 1 — backend (`internal/ytdlp`, `internal/player`, `app_player.go`)
- `search.go`: `Search` (video: `ytsearchN:`; música: `music.youtube.com/search`). La búsqueda de música mezcla
  canciones con álbumes/artistas (`YoutubeTab`, se descartan) y **solo trae el título** (sin artista/duración/
  carátula; la carátula se deriva de `i.ytimg.com`, el resto llega al resolver).
- `resolve.go`: `Resolve`, `ParseMedia`, `Pick(media, Prefs)`. Reglas aprendidas de datos reales: solo `https`
  (fuera HLS/mhtml), fuera variantes DRC, **solo audio estéreo** (itags 256/258/380/328 son 5.1), a igual altura
  prefiere el **menor fps**, el frontend manda la lista de códecs que decodifica.
- `internal/player/server.go`: proxy en `127.0.0.1:<puerto>` con secreto aleatorio en la ruta, comprobación de
  `Host`, `Range`, cabeceras de yt-dlp, renovación de URL si hay 403/410, vaciado inmediato de cabeceras
  (bug encontrado: el cliente esperaba al primer byte) y **troceo de rangos a 10 MiB** (ver hallazgos).
  `cache.go` (`MediaCache`, TTL 30 min o 10 min antes del `expire`), `stream.go` (`Resolver`, `StreamInfo`).
- `app_player.go`: `Search`, `CancelSearch`, `Resolve`; servidor iniciado en `startup`, cerrado en `shutdown`.
- Tests: fixtures reales recortados en `internal/ytdlp/testdata/`; `go test -race ./internal/...` pasa. Tests con
  red, etiquetados `e2e`: `go test -tags e2e ./internal/ytdlp ./internal/player -run E2E -v` (necesitan
  `build/deps/linux-amd64`). Verificados: búsqueda real, resolución real y bytes WebM reales a través del proxy.

### Pasos 2–4 — frontend
- `lib/player.svelte.ts` (almacén global: cola, repetir, aleatorio, volumen, calidad, MediaSession; los elementos
  `<audio>`/`<video>` viven aquí para que la música siga al cambiar de vista; **audio = reloj maestro**, video mudo
  que lo sigue), `lib/avsync.ts`, `lib/queue.ts` (+ 13 tests con `vitest`, nueva dependencia de desarrollo),
  `lib/search.svelte.ts`, `lib/downloads.ts`.
- Componentes: `Slider`, `Transport`, `MiniPlayer`, `Stage` (video, pantalla completa, calidad, velocidad, atajos),
  `QueueList`, `Cover` (miniatura con respaldo); vista `routes/Reproductor.svelte`; sección «Reproductor» en
  Ajustes; `Icon` con iconos nuevos; `Choice` con `onchange`.
- Gancho de diagnóstico opcional: con `localStorage['marte-media.debug']='1'` se expone `window.__marte`
  (elementos, controles, `flags()`, `trail` con tiempos y errores). Apagado por defecto.

## Medido en la app real
- **Offline** (medios VP9/Opus generados con ffmpeg + yt-dlp falso): deriva A/V máx **65–73 ms** en 12 s, 0/363
  fotogramas perdidos; saltar, pausar/reanudar y cambiar de calidad mantienen la sincronía y el audio no se
  interrumpe; música sigue al cambiar de vista; siguiente, playlist por enlace y «Descargar audio» (crea un trabajo
  en la Cola) funcionan. Capturas de cada estado revisadas.
- **YouTube real**: búsqueda de video 2,4 s (30 resultados), música 1,8 s (11 canciones); `yt-dlp -J` ≈ 3 s;
  deriva A/V máx **10–42 ms** en 15 s, 0 fotogramas perdidos; saltar +90 s y cambiar de calidad sincronizados.
- Bug de lógica real encontrado: el audio se quedaba parado con video. WebKit deja `readyState` en 2 mientras el
  video reproduce; ahora «listo» = no está buscando y (readyState ≥ 3 **o** el video avanza).
- Estado `loading` pegado tras saltos: WebKit no siempre emite `playing` tras `waiting`; ahora `timeupdate` lo corrige.
- `canPlayType` en este equipo dice «probably» para vp9, avc1, av01, opus y aac (un video de 144p se reprodujo
  como H.264, así que hay decodificador H.264 aunque `gst-inspect` no mostró `avdec_h264`).

## Problema del 403: RESUELTO (2026-10-27)
**Causa real (no era IPv6):** los formatos del cliente `TVHTML5` traen `available_at` ≈ 5 s en el futuro y la URL
responde 403 hasta ese instante (yt-dlp lo cumple con «Sleeping 5.00 seconds as required by the site»). Medido:
`curl` sobre la misma URL da 403, 403, 206 al pasar el instante; con `-4`/`-6`, UA o cabeceras no cambia nada.
Como la pista tardaba ≈3 s en resolverse y el elemento pedía enseguida, casi siempre caía en la ventana, y cada
renovación (otro `yt-dlp`) volvía a caer en ella.
**Arreglo:** `ytdlp.Format.AvailableAt` (de `available_at`), `player.Source.ReadyAt/NextReady` y `waitUntil` en
`fetch` (el proxy espera antes de pedir; también tras `Refresh`). Test `TestWaitsUntilReady`. Verificado con
YouTube real: 5 videos × audio/video × 4 rangos, sin ningún 403/502.
Prueba real en la app (2026-10-27, tras el arreglo): búsqueda 2,4 s; video arranca en 12,8 s (VP9 360p, deriva máx.
0,36 s, seek +90 s con deriva −0,03 s, cambio de calidad ok, 0 fotogramas perdidos); música arranca en 16,7 s con la
pista correcta y avanza. El arranque ≈ resolver (≈3,3 s) + espera de `available_at` (5 s) + carga; antes 22–78 s.
**H.264:** WebKitGTK dice «probably» para `avc1` pero no lo decodifica (video 0x0, `readyState` 0) → en Linux
`capabilities()` ya no ofrece `avc1`; Windows lo conserva (sin probar).
- Secundario: tras `setQuality` con la misma altura el video quedó una vez en `0x0` (sin `error`, `readyState=4` a
  los 5 s); revisar.

## Pendiente (no empezado)
- Repetir la prueba real en la app (video, música, enlace de playlist) tras el arreglo del 403.
- (hecho) Una pista que falla se reintenta una vez antes de saltar; el aviso muestra el motivo.
- Probar con clic real (todas las pruebas usaron clics/eventos programáticos) y el menú contextual de resultados.
- Pantalla completa, atajos de teclado y teclas multimedia sin probar. Windows (WebView2) sin probar.
- Reproducir recientes / historial en la vista inicial (hoy la vista vacía solo muestra sugerencias).
- `DESIGN.md` sin la sección del reproductor; README sin actualizar.
- Paquetes `.deb`/`.rpm`/instalador **sin regenerar** con nada de esto; `build/bin/marte-media` actual se compiló con
  el gancho de pruebas (inofensivo: solo actúa si existe `MARTE_JS1`), conviene recompilar antes de usarlo.
- No hay ningún commit en este repo; todo está sin versionar.

## Cómo retomar
1. `go test -race ./internal/... && (cd frontend && npm run check && npx vitest run)` debe pasar.
2. Banco de pruebas en `docs/player-testbed/` (escrito para una carpeta temporal: **ajustar las rutas** de los
   `run-*.sh`). Necesita medios de prueba, p. ej.:
   `ffmpeg -f lavfi -i testsrc2=size=1280x720:rate=30:duration=90 -c:v libvpx-vp9 -b:v 900k -deadline realtime -cpu-used 8 -g 30 -an video_720.webm`
   (igual `video_360.webm` a 640x360) y `ffmpeg -f lavfi -i sine=frequency=440:duration=90 -c:a libopus -b:a 96k audio.webm`,
   servidos con `rangeserver.py 18765`. Los scripts usan un gancho temporal en `main.go` (`OnDomReady` que ejecuta
   `os.Getenv("MARTE_JS1")` con `runtime.WindowExecJS` tras 1,5 s); **no está en el código** y debe retirarse al
   terminar. `run-real.sh` usa YouTube real con `PATH=/usr/bin:/bin` y los binarios incluidos junto a una copia del
   ejecutable.
3. Los scripts capturan la ventana con `niri msg action screenshot-window` (escribe en `~/Pictures/Screenshots`;
   los scripts mueven el archivo y no dejan nada allí).

## Archivos tocados o creados
Backend: `internal/ytdlp/{binary,probe,download,search,resolve}.go` + tests + `testdata/` + `e2e_test.go`;
`internal/player/{server,cache,stream}.go` + tests + `e2e_test.go`; `internal/queue/queue.go` (+test);
`app.go`, `app_player.go`, `main.go` (`OnShutdown`).
Frontend: `src/lib/{player.svelte.ts,search.svelte.ts,downloads.ts,avsync.ts,queue.ts,types.ts,format.ts,tokens.css}`
(+ `avsync.test.ts`, `queue.test.ts`); `src/lib/components/{Slider,Transport,MiniPlayer,Stage,QueueList,Cover,Icon,Choice}.svelte`;
`src/routes/{Reproductor,Settings,Download}.svelte`; `src/App.svelte`; `package.json` (vitest).
Empaquetado: `scripts/fetch-deps.sh`, `packaging/nfpm.yaml`, `build/windows/installer/project.nsi`, `THIRD_PARTY.md`.

## Hallazgos de entorno útiles
- Googlevideo estrangula los rangos abiertos (`bytes=N-`): **~40 KB/s frente a ~3 MB/s con trozos de 10 MiB**
  (medido). Por eso yt-dlp usa `http_chunk_size` 10 MiB y el proxy ahora parte cada petición.
- niri: `WAYLAND_DEBUG` y el volcado de protocolo ayudaron en el problema del teclado; no relevante aquí.
- El scratchpad de pruebas (`/tmp/claude-1000/...`) se borra al reiniciar el equipo.

## Actualización 2026-10-27 (tarde)
- Hecho: 403 resuelto, `avc1` fuera en Linux, reintento único por pista, secciones de `DESIGN.md` y `README.md`.
- Paquetes regenerados con todo el reproductor y deno: `.deb`/`.rpm` (146 MB, `deno` en `/usr/lib/marte-media/`) e
  instalador de Windows (125 MB). El binario ya no lleva el gancho de pruebas. Sin probar en Windows.
- Sigue sin probar: clic real, menú contextual de resultados, pantalla completa, atajos, teclas multimedia.

## Renombrado (2026-10-27)
La app pasó de `marte-desc` a **Marte Media** (`marte-media`), carpeta del repo incluida. Se migra `~/.config/marte-desc`
y la clave `marte-desc.player` de localStorage. `docs/player-testbed/README.md` explica cómo montar el banco de pruebas.

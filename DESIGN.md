---
name: YouTube Modern
colors:
  primary: "#FF0000"        # YouTube Red
  background: "#0f0f0f"     # Deep Black (Dark Mode)
  surface: "#272727"        # Card/Button Grey
  text: "#ffffff"           # Primary White
  text-muted: "#aaaaaa"     # Secondary Metadata
  # Marte extension: tints of the same palette, no new hues
  surface-low: "#171717"    # Panels sitting between background and surface
  border: "#2c2c2c"
  border-strong: "#3f3f3f"
  ok: "#2ba640"             # Completed / ready
  danger: "#ff6b6b"         # Error text
typography:
  fontFamily: "Roboto, Arial, sans-serif"
  displayFamily: "'Space Grotesk', Roboto, Arial, sans-serif"   # titles and the wordmark
  baseSize: "14px"
rounded:
  card: "12px"
  button: "18px"
  panel: "20px"
---
# YouTube Design System
- **Focus**: High information density with a "Cinematic" feel. 
- **Cards**: Video thumbnails use a 12px radius. Titles are limited to 2 lines (line-clamp).
- **Navigation**: Minimalist sidebar icons (24px) that collapse to a rail.

# Marte extension
marte-desc is named after the planet. The YouTube foundation above stays; Mars gives it a personality.
The palette does not change: red, near-black and greys. Red is the planet.

- **The planet**: a drawn Mars (red body, polar cap, canyon scar, craters, night side, tilted orbit ring) is the
  brand mark. It appears as the rail logo, the hero illustration on the download screen, the empty states and
  the app icon (`tools/genicon`). Keep it red/white only.
- **Horizon**: the stage behind the content carries a faint stars layer and the curve of a planet along the
  bottom edge, lit red from above. Content always sits on top and stays readable.
- **Grain**: a fixed film-grain overlay (5% opacity) breaks the flat black. Shadows are tinted red-black,
  never plain black, and share one light source (from above).
- **Type**: Space Grotesk 600/700 for titles, tight tracking (-0.02em). Roboto for everything else. Numbers
  that change (percent, size, speed, counts) use tabular figures.
- **Controls**: no native selects. One-of-many choices are a joined *segment* (2-3 options) or *chips*
  (more). Primary actions are red with a soft red glow; everything else is grey or ghost.
- **Active section**: a red notch on the rail's edge, not just a background change.
- **Progress**: a thin bar whose red fill glows; green when done. Status uses a dot plus a word, never colour alone.
- **Motion**: transform and opacity only, 180-260ms, staggered entry for lists, a small orbit instead of a
  spinner, skeleton blocks while analysing. Everything respects `prefers-reduced-motion`.
- **Performance**: decorative motion must stay on the GPU. The planet drifts with `translate` only, on its own
  layer, with no SVG filters; nothing animates `width`, `top` or shadows. Animations pause when the window is
  hidden or unfocused. If the webview paints on the CPU (X11, or `WEBKIT_DISABLE_DMABUF_RENDERER=1`) the UI
  enters low-graphics mode (`html[data-lowgfx]`): no grain, no shadows, no transitions. Never set that variable
  on Wayland: it costs about 4x the frame time.
- **States**: every list has an empty state (a small planet, a plain sentence, one action), loading has a
  skeleton, errors are inline text with an icon, and finished downloads raise a toast.
- **Copy**: sentence case, plain and specific, no exclamation marks.
- **Focus**: a 2px white outline on every focusable element.
- **z-index scale**: stage 0, content 1, toast 50, menu 100, grain 200 (see `tokens.css`).

# Reproductor
The first view of the app: search and play video or music without downloading.

- **Layout**: search bar plus a *Video | Música* segment. Video results are 16:9 cards, music results are rows with a
  square cover. A link or playlist pasted into the bar plays instead of searching. Skeleton blocks while searching;
  Esc cancels.
- **Mini player**: a bottom bar that survives view changes (cover, title, transport, progress, volume, queue popover,
  download). It hides while the full stage is open.
- **Stage**: the watching layout, video (or cover for music) with the queue beside it. Controls are our own: play, seek
  with buffer, volume, quality, speed, fullscreen. Nothing decorative animates over the picture.
- **Quality menu** only lists what the webview can decode. H.264 is not offered on Linux (WebKitGTK reports it but
  fails to render it).
- **z-index**: mini player 5, full stage 90 (`--z-player`, `--z-full`).
- **Failure**: a track that fails is retried once, then skipped when autoplay is on; the toast says why.

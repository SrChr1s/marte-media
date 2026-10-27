<script lang="ts">
  // A horizontal slider for the seek bar and the volume. Mouse, touch and keyboard.
  let {
    value,
    max,
    buffered = 0,
    label,
    step = 5,
    thick = false,
    onseek,
    onpreview,
  }: { value: number; max: number; buffered?: number; label: string; step?: number; thick?: boolean; onseek: (v: number) => void; onpreview?: (v: number | null) => void } = $props()

  let el = $state<HTMLDivElement>()
  let preview = $state<number | null>(null)
  const shown = $derived(preview ?? value)
  const ratio = (v: number) => (max > 0 ? Math.max(0, Math.min(1, v / max)) : 0)

  function at(e: PointerEvent): number {
    const r = el!.getBoundingClientRect()
    return Math.max(0, Math.min(1, (e.clientX - r.left) / r.width)) * max
  }
  function down(e: PointerEvent) {
    el!.setPointerCapture(e.pointerId)
    preview = at(e)
    onpreview?.(preview)
  }
  function move(e: PointerEvent) {
    if (preview === null) return
    preview = at(e)
    onpreview?.(preview)
  }
  function up() {
    if (preview === null) return
    onseek(preview)
    preview = null
    onpreview?.(null)
  }
  function key(e: KeyboardEvent) {
    const big = max / 10
    const to: Record<string, number> = {
      ArrowRight: shown + step,
      ArrowUp: shown + step,
      ArrowLeft: shown - step,
      ArrowDown: shown - step,
      PageUp: shown + big,
      PageDown: shown - big,
      Home: 0,
      End: max,
    }
    if (!(e.key in to)) return
    e.preventDefault()
    onseek(Math.max(0, Math.min(max, to[e.key])))
  }
</script>

<div
  bind:this={el}
  class="slider"
  class:thick
  class:active={preview !== null}
  role="slider"
  tabindex="0"
  aria-label={label}
  aria-valuemin={0}
  aria-valuemax={Math.round(max)}
  aria-valuenow={Math.round(shown)}
  onpointerdown={down}
  onpointermove={move}
  onpointerup={up}
  onpointercancel={up}
  onkeydown={key}
>
  <div class="track">
    <div class="buf" style="transform: scaleX({ratio(buffered)})"></div>
    <div class="fill" style="transform: scaleX({ratio(shown)})"></div>
  </div>
  <div class="thumb" style="left: {ratio(shown) * 100}%"></div>
</div>

<style>
  .slider {
    position: relative;
    flex: 1;
    min-width: 0;
    height: 20px;
    display: flex;
    align-items: center;
    cursor: pointer;
    touch-action: none;
  }
  .track {
    position: relative;
    width: 100%;
    height: 4px;
    border-radius: 2px;
    background: rgba(255, 255, 255, 0.16);
    overflow: hidden;
    transition: height 120ms var(--ease);
  }
  .thick .track {
    height: 5px;
  }
  .slider:hover .track,
  .slider.active .track,
  .slider:focus-visible .track {
    height: 6px;
  }
  .buf,
  .fill {
    position: absolute;
    inset: 0;
    transform-origin: left;
  }
  .buf {
    background: rgba(255, 255, 255, 0.22);
  }
  .fill {
    background: var(--primary);
  }
  .thumb {
    position: absolute;
    width: 12px;
    height: 12px;
    margin-left: -6px;
    border-radius: 50%;
    background: #fff;
    box-shadow: 0 0 8px rgba(255, 0, 0, 0.6);
    transform: scale(0);
    transition: transform 120ms var(--ease);
    pointer-events: none;
  }
  .slider:hover .thumb,
  .slider.active .thumb,
  .slider:focus-visible .thumb {
    transform: scale(1);
  }
</style>

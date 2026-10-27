<script lang="ts">
  // Small planet mark for the rail: body, canyon scar and a tilted ring.
  let { size = 36 }: { size?: number } = $props()
  const id = Math.random().toString(36).slice(2, 8)
</script>

<svg width={size} height={size} viewBox="0 0 48 48" role="img" aria-label="marte-desc">
  <defs>
    <radialGradient id="b-{id}" cx="34%" cy="30%" r="80%">
      <stop offset="0" stop-color="#ff3a2c" />
      <stop offset="0.5" stop-color="#d40a00" />
      <stop offset="1" stop-color="#520000" />
    </radialGradient>
    <clipPath id="c-{id}"><circle cx="24" cy="24" r="14" /></clipPath>
  </defs>
  <g transform="rotate(-18 24 24)"><path d="M3 24 A21 6.5 0 0 1 45 24" fill="none" stroke="#fff" stroke-opacity="0.4" stroke-width="1" /></g>
  <circle cx="24" cy="24" r="14" fill="url(#b-{id})" />
  <g clip-path="url(#c-{id})">
    <ellipse cx="24" cy="12" rx="6" ry="2" fill="#f6efee" opacity="0.85" />
    <path d="M12 26 C17 24.5 22 27.5 28 26 C31 25.3 33 26 35 27" fill="none" stroke="#3a0000" stroke-width="1.6" stroke-linecap="round" opacity="0.6" />
    <circle cx="30" cy="19" r="2" fill="#6a0000" opacity="0.5" />
  </g>
  <g transform="rotate(-18 24 24)"><path d="M3 24 A21 6.5 0 0 0 45 24" fill="none" stroke="#fff" stroke-opacity="0.55" stroke-width="1.1" /></g>
</svg>

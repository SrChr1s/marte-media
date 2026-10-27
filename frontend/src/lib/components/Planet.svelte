<script lang="ts">
  // Mars drawn in SVG: red body, polar cap, a canyon scar, craters, a night side and a tilted orbit ring.
  let { size = 320, ring = true, float = false }: { size?: number; ring?: boolean; float?: boolean } = $props()
  const id = Math.random().toString(36).slice(2, 8)
</script>

<svg class:float width={size} height={size} viewBox="0 0 400 400" role="img" aria-label="Planeta Marte">
  <defs>
    <radialGradient id="body-{id}" cx="34%" cy="30%" r="82%">
      <stop offset="0" stop-color="#ff3a2c" />
      <stop offset="0.42" stop-color="#d40a00" />
      <stop offset="1" stop-color="#4e0000" />
    </radialGradient>
    <radialGradient id="night-{id}" cx="30%" cy="26%" r="88%">
      <stop offset="0.55" stop-color="#000" stop-opacity="0" />
      <stop offset="1" stop-color="#000" stop-opacity="0.78" />
    </radialGradient>
    <radialGradient id="halo-{id}" cx="50%" cy="50%" r="50%">
      <stop offset="0.62" stop-color="#ff1a0a" stop-opacity="0.3" />
      <stop offset="1" stop-color="#ff1a0a" stop-opacity="0" />
    </radialGradient>
    <linearGradient id="ring-{id}" x1="0" x2="1">
      <stop offset="0" stop-color="#fff" stop-opacity="0" />
      <stop offset="0.5" stop-color="#fff" stop-opacity="0.42" />
      <stop offset="1" stop-color="#fff" stop-opacity="0" />
    </linearGradient>
    <clipPath id="clip-{id}"><circle cx="200" cy="200" r="120" /></clipPath>
    <radialGradient id="light-{id}">
      <stop offset="0" stop-color="#ff5a48" stop-opacity="0.34" />
      <stop offset="1" stop-color="#ff5a48" stop-opacity="0" />
    </radialGradient>
    <radialGradient id="dark-{id}">
      <stop offset="0" stop-color="#2a0000" stop-opacity="0.5" />
      <stop offset="1" stop-color="#2a0000" stop-opacity="0" />
    </radialGradient>
    <radialGradient id="cap-{id}">
      <stop offset="0.45" stop-color="#f6efee" stop-opacity="0.95" />
      <stop offset="1" stop-color="#f6efee" stop-opacity="0" />
    </radialGradient>
    <linearGradient id="scar-{id}" gradientUnits="userSpaceOnUse" x1="104" x2="304">
      <stop offset="0" stop-color="#3a0000" stop-opacity="0" />
      <stop offset="0.25" stop-color="#3a0000" stop-opacity="0.6" />
      <stop offset="0.75" stop-color="#3a0000" stop-opacity="0.6" />
      <stop offset="1" stop-color="#3a0000" stop-opacity="0" />
    </linearGradient>
  </defs>

  <circle cx="200" cy="200" r="190" fill="url(#halo-{id})" />

  {#if ring}
    <g transform="rotate(-18 200 200)"><path d="M10 200 A190 52 0 0 1 390 200" fill="none" stroke="url(#ring-{id})" stroke-width="1.4" /></g>
  {/if}

  <circle cx="200" cy="200" r="120" fill="url(#body-{id})" />
  <g clip-path="url(#clip-{id})">
    <ellipse cx="150" cy="150" rx="104" ry="34" fill="url(#light-{id})" transform="rotate(-24 150 150)" />
    <ellipse cx="260" cy="262" rx="112" ry="38" fill="url(#dark-{id})" transform="rotate(-14 260 262)" />
    <ellipse cx="202" cy="88" rx="54" ry="19" fill="url(#cap-{id})" />
    <path d="M104 214 C146 202 188 226 238 212 C268 204 288 212 304 222" fill="none" stroke="url(#scar-{id})" stroke-width="15" stroke-linecap="round" opacity="0.35" />
    <path d="M104 214 C146 202 188 226 238 212 C268 204 288 212 304 222" fill="none" stroke="url(#scar-{id})" stroke-width="8" stroke-linecap="round" />
    <path d="M118 224 C158 214 196 236 240 224" fill="none" stroke="#ff6a58" stroke-width="1.6" stroke-linecap="round" opacity="0.3" />
    <ellipse cx="254" cy="156" rx="17" ry="11" fill="#6a0000" opacity="0.5" />
    <ellipse cx="254" cy="153" rx="17" ry="11" fill="none" stroke="#ff7a68" stroke-opacity="0.35" />
    <ellipse cx="148" cy="262" rx="12" ry="7" fill="#6a0000" opacity="0.5" />
    <ellipse cx="148" cy="260" rx="12" ry="7" fill="none" stroke="#ff7a68" stroke-opacity="0.3" />
    <ellipse cx="170" cy="130" rx="7" ry="4.5" fill="#6a0000" opacity="0.45" />
    <circle cx="200" cy="200" r="120" fill="url(#night-{id})" />
  </g>

  {#if ring}
    <g transform="rotate(-18 200 200)">
      <path d="M10 200 A190 52 0 0 0 390 200" fill="none" stroke="url(#ring-{id})" stroke-width="1.6" />
      <circle cx="318" cy="238" r="6" fill="#fff" />
      <circle cx="318" cy="238" r="14" fill="#fff" opacity="0.14" />
    </g>
  {/if}
</svg>

<style>
  svg {
    display: block;
    overflow: visible;
  }
  svg.float {
    animation: drift 9s ease-in-out infinite;
    will-change: translate;
  }
  @keyframes drift {
    50% {
      translate: 0 -6px;
    }
  }
</style>

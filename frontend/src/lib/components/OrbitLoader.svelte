<script lang="ts">
  // A dot circling a small planet; stands in for a generic spinner.
  let { size = 28 }: { size?: number } = $props()
</script>

<span class="loader" style="--s: {size}px" role="status" aria-label="Cargando">
  <span class="planet"></span>
  <span class="orbit"><span class="dot"></span></span>
</span>

<style>
  .loader {
    position: relative;
    display: inline-block;
    width: var(--s);
    height: var(--s);
  }
  .planet {
    position: absolute;
    inset: 30%;
    border-radius: 50%;
    background: radial-gradient(circle at 35% 30%, #ff3a2c, #b00800 60%, #4e0000);
  }
  .orbit {
    position: absolute;
    inset: 0;
    border-radius: 50%;
    border: 1px solid rgba(255, 255, 255, 0.14);
    animation: spin 1.5s linear infinite;
  }
  .dot {
    position: absolute;
    top: -2.5px;
    left: calc(50% - 2.5px);
    width: 5px;
    height: 5px;
    border-radius: 50%;
    background: #fff;
    box-shadow: 0 0 8px rgba(255, 255, 255, 0.8);
  }
  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>

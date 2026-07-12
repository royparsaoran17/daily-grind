<script setup lang="ts">
// Faux device status bar to match the design's phone-frame chrome.
const time = ref('07.30')
onMounted(() => {
  const set = () => {
    time.value = new Date().toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' }).replace(':', '.')
  }
  set()
})
</script>

<template>
  <div class="status">
    <span>{{ time }}</span>
    <span class="ri">
      <i class="ph-fill ph-cell-signal-full" />
      <i class="ph-fill ph-wifi-high" />
      <i class="ph-fill ph-battery-full" />
    </span>
  </div>
</template>

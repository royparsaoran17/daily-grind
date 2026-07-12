<script setup lang="ts">
const { toasts } = useToast()
</script>

<template>
  <ClientOnly>
    <TransitionGroup name="fade" tag="div">
      <div v-for="t in toasts" :key="t.id" class="toast">
        <i :class="t.icon" />{{ t.text }}
      </div>
    </TransitionGroup>
  </ClientOnly>
</template>

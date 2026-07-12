interface Toast { id: number; text: string; icon: string }

const toasts = ref<Toast[]>([])
let seq = 0

/** Lightweight transient toast, used for reward feedback (+EXP, level up). */
export function useToast() {
  function show(text: string, icon = 'ph-fill ph-lightning') {
    const id = ++seq
    toasts.value = [...toasts.value, { id, text, icon }]
    setTimeout(() => {
      toasts.value = toasts.value.filter((t) => t.id !== id)
    }, 2200)
  }
  return { toasts, show }
}

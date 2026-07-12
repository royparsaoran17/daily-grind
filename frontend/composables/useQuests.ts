import type { User } from '~/stores/auth'
import { ATTR_ICON } from '~/utils/format'

export interface Quest {
  id: string
  name: string
  categoryId: string
  category: string
  icon: string
  attribute: string
  frequency: string
  difficulty: string
  expReward: number
  coinReward: number
  reminder?: string
  weekday?: number | null
  dayOfMonth?: number | null
  schedule: string
  streak: number
  done: boolean
  dueToday: boolean
}

const quests = ref<Quest[]>([])
const loaded = ref(false)

/** Shared quest list + optimistic completion toggling. */
export function useQuests() {
  const auth = useAuthStore()
  const { show } = useToast()

  async function load(force = false) {
    if (loaded.value && !force) return
    quests.value = await useApi()<Quest[]>('/quests')
    loaded.value = true
  }

  async function toggle(quest: Quest) {
    const target = quests.value.find((q) => q.id === quest.id)
    if (!target) return
    const wasDone = target.done

    // Optimistic UI
    target.done = !wasDone
    target.streak += wasDone ? -1 : 1
    if (target.streak < 0) target.streak = 0

    try {
      const user = await useApi()<User>(`/quests/${quest.id}/complete`, {
        method: wasDone ? 'DELETE' : 'POST',
      })
      const prevLevel = auth.user?.level ?? user.level
      auth.setUser(user)
      if (!wasDone) {
        if (user.level > prevLevel) {
          show(`Naik ke Level ${user.level}! 🎉`, 'ph-fill ph-arrow-fat-up')
        } else {
          show(`+${quest.expReward} EXP · ${quest.attribute.toUpperCase()}`)
        }
      }
    } catch {
      // Roll back on failure
      target.done = wasDone
      target.streak += wasDone ? 1 : -1
      show('Gagal menyimpan. Coba lagi.', 'ph-fill ph-warning')
    }
  }

  async function create(payload: {
    name: string; categoryId: string; frequency: string; difficulty: string; reminder: string
    weekday?: number | null; dayOfMonth?: number | null
  }) {
    await useApi()('/quests', { method: 'POST', body: payload })
    await load(true)
  }

  const doneCount = computed(() => quests.value.filter((q) => q.done).length)
  const totalExpToday = computed(() =>
    quests.value.filter((q) => q.done).reduce((s, q) => s + q.expReward, 0),
  )

  return { quests, load, toggle, create, doneCount, totalExpToday, ATTR_ICON }
}

import { computed, ref, watch } from 'vue'
import { defineStore } from 'pinia'

import type { SavingsGoal } from '@/entities/savings-goal/model/types'
import { createId } from '@/shared/lib/id'
import { demoSavingsGoals } from '@/shared/mocks/demo'
import { readStorage, writeStorage } from '@/shared/lib/storage'

const SAVINGS_STORAGE_KEY = 'plos-savings-goals'
const SAVINGS_FILTERS_STORAGE_KEY = 'plos-savings-filters'

function cloneGoals() {
  return demoSavingsGoals.map((goal) => ({ ...goal }))
}

export const useSavingsStore = defineStore('savings', () => {
  const goals = ref<SavingsGoal[]>(cloneGoals())
  const showCompleted = ref(false)
  const hydrated = ref(false)

  function bootstrap() {
    goals.value = readStorage<SavingsGoal[]>(SAVINGS_STORAGE_KEY, cloneGoals())
    showCompleted.value = readStorage<boolean>(SAVINGS_FILTERS_STORAGE_KEY, false)
    hydrated.value = true
  }

  const visibleGoals = computed(() =>
    goals.value.filter((goal) => showCompleted.value || !goal.isCompleted),
  )

  const progressRatio = computed(() => {
    const target = goals.value.reduce((sum, goal) => sum + goal.targetMinor, 0)
    const saved = goals.value.reduce((sum, goal) => sum + goal.savedMinor, 0)

    if (!target) {
      return 0
    }

    return saved / target
  })

  function addGoal(input: { name: string; targetMinor: number; targetDate: string | null }) {
    const goal: SavingsGoal = {
      id: createId('goal'),
      name: input.name.trim(),
      targetMinor: input.targetMinor,
      savedMinor: 0,
      currency: 'RUB',
      targetDate: input.targetDate,
      isCompleted: false,
    }

    goals.value = [goal, ...goals.value]
    return goal
  }

  function toggleShowCompleted() {
    showCompleted.value = !showCompleted.value
  }

  function reset() {
    goals.value = cloneGoals()
    showCompleted.value = false
  }

  watch(
    goals,
    (nextGoals) => {
      if (hydrated.value) {
        writeStorage(SAVINGS_STORAGE_KEY, nextGoals)
      }
    },
    {
      deep: true,
    },
  )
  watch(showCompleted, (value) => {
    if (hydrated.value) {
      writeStorage(SAVINGS_FILTERS_STORAGE_KEY, value)
    }
  })

  return {
    addGoal,
    bootstrap,
    goals,
    progressRatio,
    reset,
    showCompleted,
    toggleShowCompleted,
    visibleGoals,
  }
})

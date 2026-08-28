<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import { normalizeApiError } from '@/api/http'
import { updateSystemConfig, type SystemUpdateData } from '@/api/services/system'
import SignupForm from '@/components/blocks/signup-02/components/SignupForm.vue'
import { useAuthStore } from '@/stores/auth'
import { useSystemInitStore } from '@/stores/system-init'

const router = useRouter()
const auth = useAuthStore()
const systemInit = useSystemInitStore()
const pending = ref(false)
const error = ref<string | null>(null)

async function submit(settings: SystemUpdateData): Promise<void> {
  pending.value = true
  error.value = null
  try {
    await updateSystemConfig(settings)
    await systemInit.initialize()
    if (!systemInit.isAvailable) {
      await router.replace({ name: 'server-unavailable' })
      return
    }
    await auth.restoreSession()
    await router.replace({ name: auth.isAuthenticated ? 'home' : 'login' })
  } catch (cause) {
    error.value = normalizeApiError(cause).message
  } finally {
    pending.value = false
  }
}
</script>

<template>
  <div class="grid min-h-svh lg:grid-cols-[minmax(0,1.25fr)_minmax(22rem,.75fr)]">
    <div class="flex flex-col gap-6 p-6 md:p-10">
      <div class="flex items-center gap-2 font-medium">
        <span class="grid size-7 place-items-center rounded-md bg-primary text-xs text-primary-foreground">OC</span>
        OCServ Dashboard
      </div>
      <div class="mx-auto flex w-full max-w-2xl flex-1 items-center py-6">
        <SignupForm class="w-full" :error="error" :pending="pending" @submit="submit" />
      </div>
    </div>
    <aside class="relative hidden overflow-hidden bg-zinc-950 text-white lg:flex lg:flex-col lg:justify-end lg:p-12">
      <div class="absolute inset-0 bg-[radial-gradient(circle_at_20%_20%,oklch(0.55_0.19_260/.55),transparent_40%),radial-gradient(circle_at_80%_70%,oklch(0.62_0.18_185/.3),transparent_35%)]" />
      <div class="relative">
        <p class="text-2xl font-semibold">A clean start for your private network.</p>
        <p class="mt-3 max-w-md text-sm leading-6 text-white/65">These settings are sent directly to the generated panel system API.</p>
      </div>
    </aside>
  </div>
</template>

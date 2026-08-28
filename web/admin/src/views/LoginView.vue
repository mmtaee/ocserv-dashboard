<script setup lang="ts">
import { useRouter } from 'vue-router'

import type { GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemLoginData } from '@/api/generated'
import LoginForm from '@/components/blocks/login-02/components/LoginForm.vue'
import { useAuthStore } from '@/stores/auth'
import { useSystemInitStore } from '@/stores/system-init'

const router = useRouter()
const auth = useAuthStore()
const systemInit = useSystemInitStore()

async function submit(credentials: GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemLoginData): Promise<void> {
  if (await auth.signIn(credentials)) await router.replace({ name: 'home' })
}
</script>

<template>
  <div class="grid min-h-svh lg:grid-cols-2">
    <div class="flex flex-col gap-4 p-6 md:p-10">
      <div class="flex items-center gap-2 font-medium">
        <span class="grid size-7 place-items-center rounded-md bg-primary text-xs text-primary-foreground">OC</span>
        OCServ Dashboard
      </div>
      <div class="flex flex-1 items-center justify-center">
        <div class="w-full max-w-sm">
          <LoginForm :captcha-site-key="systemInit.captchaSiteKey" :error="auth.error" :pending="auth.isLoading" @submit="submit" />
        </div>
      </div>
    </div>
    <div class="relative hidden overflow-hidden bg-zinc-950 lg:block">
      <div class="absolute inset-0 bg-[radial-gradient(circle_at_30%_20%,oklch(0.55_0.19_260/.55),transparent_42%),radial-gradient(circle_at_75%_75%,oklch(0.62_0.18_185/.35),transparent_38%)]" />
      <div class="absolute inset-x-12 bottom-12 rounded-xl border border-white/10 bg-white/5 p-6 text-white backdrop-blur">
        <p class="text-lg font-medium">Secure infrastructure administration</p>
        <p class="mt-2 text-sm text-white/65">Monitor sessions, users, agents, and system health from one panel.</p>
      </div>
    </div>
  </div>
</template>

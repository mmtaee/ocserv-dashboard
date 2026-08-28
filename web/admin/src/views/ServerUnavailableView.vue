<script setup lang="ts">
import { ServerCrash } from '@lucide/vue'
import { useRouter } from 'vue-router'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Spinner } from '@/components/ui/spinner'
import { useAuthStore } from '@/stores/auth'
import { useSystemInitStore } from '@/stores/system-init'

const router = useRouter()
const auth = useAuthStore()
const systemInit = useSystemInitStore()

async function retry(): Promise<void> {
  if (!(await systemInit.initialize())) return
  if (!systemInit.isInitialized) {
    await router.replace({ name: 'system-setup' })
    return
  }
  await auth.restoreSession()
  await router.replace({ name: auth.isAuthenticated ? 'home' : 'login' })
}
</script>

<template>
  <main class="grid min-h-svh place-items-center bg-muted/40 p-6">
    <Card class="w-full max-w-lg">
      <CardHeader>
        <div class="mb-2 grid size-11 place-items-center rounded-full bg-destructive/10 text-destructive"><ServerCrash aria-hidden="true" /></div>
        <CardTitle>Server unavailable</CardTitle>
        <CardDescription>The dashboard could not load the panel initialization config.</CardDescription>
      </CardHeader>
      <CardContent class="grid gap-4">
        <Alert v-if="systemInit.error" variant="destructive">
          <AlertTitle>Connection failed</AlertTitle>
          <AlertDescription>{{ systemInit.error }}</AlertDescription>
        </Alert>
        <Button :disabled="systemInit.isLoading" @click="retry">
          <Spinner v-if="systemInit.isLoading" data-icon="inline-start" />
          Try again
        </Button>
      </CardContent>
    </Card>
  </main>
</template>

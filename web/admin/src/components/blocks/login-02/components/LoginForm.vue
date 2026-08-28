<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { reactive } from 'vue'
import { useI18n } from 'vue-i18n'

import type { GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemLoginData } from '@/api/generated'
import CaptchaField from '@/components/atoms/CaptchaField.vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Field, FieldDescription, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Spinner } from '@/components/ui/spinner'
import { cn } from '@/lib/utils'

const props = defineProps<{
  captchaSiteKey?: string
  class?: HTMLAttributes['class']
  error?: string | null
  pending?: boolean
}>()
const { t } = useI18n({ useScope: 'global' })

const emit = defineEmits<{
  submit: [credentials: GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemLoginData]
}>()

const form = reactive<GithubComMmtaeeOcservDashboardBackendInternalServicesAdminApiSystemLoginData>({
  password: '', remember_me: false, token: undefined, username: '',
})

function submit(): void {
  if (props.captchaSiteKey && !form.token) return
  emit('submit', { ...form })
}
</script>

<template>
  <form :class="cn('flex flex-col gap-6', props.class)" @submit.prevent="submit">
    <FieldGroup>
      <div class="flex flex-col items-center gap-1 text-center">
        <h1 class="text-2xl font-bold">{{ t('auth.title') }}</h1>
        <p class="text-balance text-sm text-muted-foreground">{{ t('auth.subtitle') }}</p>
      </div>

      <Alert v-if="error" variant="destructive">
        <AlertTitle>{{ t('auth.failure') }}</AlertTitle>
        <AlertDescription>{{ error }}</AlertDescription>
      </Alert>

      <Field>
        <FieldLabel for="username">{{ t('auth.username') }}</FieldLabel>
        <Input id="username" v-model="form.username" name="username" autocomplete="username" :placeholder="t('auth.usernamePlaceholder')" required />
      </Field>
      <Field>
        <FieldLabel for="password">{{ t('auth.password') }}</FieldLabel>
        <Input id="password" v-model="form.password" name="password" type="password" autocomplete="current-password" required />
      </Field>

      <Field v-if="captchaSiteKey">
        <FieldLabel>{{ t('auth.verification') }}</FieldLabel>
        <CaptchaField v-model="form.token" :site-key="captchaSiteKey" />
        <FieldDescription>{{ t('auth.captchaPrompt') }}</FieldDescription>
      </Field>

      <Field orientation="horizontal">
        <Checkbox id="remember-me" :model-value="form.remember_me" @update:model-value="form.remember_me = $event === true" />
        <FieldLabel for="remember-me" class="font-normal">{{ t('auth.rememberMe') }}</FieldLabel>
      </Field>

      <Field>
        <Button type="submit" :disabled="pending || Boolean(captchaSiteKey && !form.token)">
          <Spinner v-if="pending" data-icon="inline-start" />
          {{ t('auth.submit') }}
        </Button>
      </Field>
    </FieldGroup>
  </form>
</template>

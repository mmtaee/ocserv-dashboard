<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { type PropType, ref, watch } from 'vue';
import type { SystemGetSystemResponse, SystemPatchSystemUpdateData } from '@/api';

const props = defineProps<{
    data: SystemGetSystemResponse;
    loading: boolean;
}>();

const emit = defineEmits(['updateSystem', 'cancel']);

const { t } = useI18n();

const systemData = ref<SystemPatchSystemUpdateData>({
    auto_delete_inactive_users: false,
    google_captcha_secret_key: '',
    google_captcha_site_key: '',
    keep_inactive_user_days: 0
});

watch(
    () => props.data,
    (newData) => {
        if (!newData) return;
        systemData.value = {
            auto_delete_inactive_users: newData.auto_delete_inactive_users ?? false,
            google_captcha_secret_key: newData.google_captcha_secret_key ?? '',
            google_captcha_site_key: newData.google_captcha_site_key ?? '',
            keep_inactive_user_days: newData.keep_inactive_user_days ?? 0
        };
    },
    { immediate: true, deep: true }
);
</script>

<template>
    <v-row align="start" justify="center">
        <v-col cols="12">
            <v-label class="font-weight-bold mb-1">
                {{ t('GOOGLE_CAPTCHA_SITE_KEY') }}
                <span class="text-subtitle-1">
                    <a
                        class="text-info mx-2"
                        href="https://www.google.com/recaptcha/admin/create"
                        style="text-decoration: none"
                    >
                        (recaptcha {{ t('HELP') }})
                    </a>
                </span>
            </v-label>
            <v-text-field
                v-model="systemData.google_captcha_site_key"
                color="primary"
                hide-details
                type="text"
                variant="outlined"
            />
        </v-col>
        <v-col cols="12">
            <v-label class="font-weight-bold mb-1">{{ t('GOOGLE_CAPTCHA_SECRET_KEY') }}</v-label>
            <v-text-field
                v-model="systemData.google_captcha_secret_key"
                color="primary"
                hide-details
                type="text"
                variant="outlined"
            />
        </v-col>

        <v-col cols="12">
            <v-row align="center">
                <v-col cols="12" lg="6" class="ma-0 pa-0">
                    <v-checkbox
                        v-model="systemData.auto_delete_inactive_users"
                        color="primary"
                        hide-details
                        class="text-capitalize mt-md-6"
                    >
                        <template v-slot:label class="text-body-1">{{ t('AUTO_DELETE_INACTIVE_USERS') }}</template>
                    </v-checkbox>
                </v-col>

                <v-col cols="12" lg="6">
                    <v-label class="font-weight-bold mb-1">
                        {{ t('KEEP_INACTIVE_USER_DAYS') }}
                    </v-label>
                    <v-text-field
                        :disabled="!systemData.auto_delete_inactive_users"
                        v-model.number="systemData.keep_inactive_user_days"
                        color="primary"
                        hide-details
                        type="number"
                        variant="outlined"
                    />
                </v-col>
            </v-row>
        </v-col>
    </v-row>
    <v-row align="center" justify="end" class="mt-md-5 ms-sm-15">
        <v-btn color="muted" variant="text" @click="emit('cancel')">
            {{ t('CANCEL') }}
        </v-btn>
        <v-btn :loading="loading" class="ms-2 me-1" color="primary" variant="flat" @click='emit("updateSystem", systemData)'>
            {{ t('UPDATE') }}
        </v-btn>
    </v-row>
</template>

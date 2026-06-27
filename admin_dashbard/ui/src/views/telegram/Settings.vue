<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { languages } from '@/plugins/i18n';
import { TelegramAPI, type TelegramSettings } from '@/api/telegram';
import { useSnackbarStore } from '@/stores/snackbar';
import UiParentCard from '@/components/shared/UiParentCard.vue';

const { t } = useI18n();
const snackbar = useSnackbarStore();

const loading = ref(false);
const showToken = ref(false);
const settings = ref<TelegramSettings>({
    enabled: false,
    bot_token: '',
    bot_username: '',
    admin_chat_id: 0,
    low_quota_threshold_mb: 200,
    default_language: 'en',
    ocserv_host: '',
    card_number: '',
    card_holder: '',
    support_username: ''
});

const load = async () => {
    loading.value = true;
    try {
        const res = await TelegramAPI.getSettings();
        settings.value = res.data;
    } finally {
        loading.value = false;
    }
};

const save = async () => {
    loading.value = true;
    try {
        const res = await TelegramAPI.updateSettings({
            enabled: settings.value.enabled,
            bot_token: settings.value.bot_token,
            admin_chat_id: Number(settings.value.admin_chat_id) || 0,
            low_quota_threshold_mb: Number(settings.value.low_quota_threshold_mb) || 200,
            default_language: settings.value.default_language,
            ocserv_host: settings.value.ocserv_host,
            card_number: settings.value.card_number,
            card_holder: settings.value.card_holder,
            support_username: (settings.value.support_username || '').replace(/^@+/, '').trim()
        });
        settings.value = res.data;
        snackbar.show({
            id: 1,
            message: t('TELEGRAM_SETTINGS_SAVED'),
            color: 'success',
            timeout: 3000
        });
    } finally {
        loading.value = false;
    }
};

const sendTest = async () => {
    loading.value = true;
    try {
        await TelegramAPI.test('Test message from Ocserv Dashboard');
        snackbar.show({
            id: 2,
            message: t('TELEGRAM_TEST_SENT'),
            color: 'success',
            timeout: 3000
        });
    } finally {
        loading.value = false;
    }
};

onMounted(load);
</script>

<template>
    <v-row>
        <v-col cols="12">
            <UiParentCard :title="t('TELEGRAM_SETTINGS')">
                <v-form @submit.prevent="save" class="pa-4">
                    <v-row dense>
                        <v-col cols="12" md="6">
                            <v-switch
                                v-model="settings.enabled"
                                :label="t('TELEGRAM_ENABLED')"
                                color="primary"
                                hide-details
                            />
                        </v-col>
                        <v-col cols="12" md="6">
                            <v-text-field
                                v-model="settings.bot_username"
                                :label="t('TELEGRAM_BOT_USERNAME')"
                                readonly
                                variant="outlined"
                                density="comfortable"
                                hide-details
                            />
                        </v-col>
                    </v-row>

                    <v-row dense class="mt-2">
                        <v-col cols="12">
                            <v-card variant="outlined" class="pa-4 mb-2">
                                <v-text-field
                                    v-model="settings.bot_token"
                                    class="telegram-token-input"
                                    :class="{ 'telegram-token-input--masked': !showToken }"
                                    :label="t('TELEGRAM_BOT_TOKEN')"
                                    type="text"
                                    name="telegram_bot_token"
                                    autocomplete="off"
                                    autocorrect="off"
                                    autocapitalize="off"
                                    spellcheck="false"
                                    inputmode="text"
                                    data-lpignore="true"
                                    data-1p-ignore="true"
                                    data-form-type="other"
                                    :append-inner-icon="showToken ? 'mdi-eye-off' : 'mdi-eye'"
                                    @click:append-inner="showToken = !showToken"
                                    variant="outlined"
                                    density="comfortable"
                                    :hint="t('TELEGRAM_BOT_TOKEN_HINT')"
                                    persistent-hint
                                />
                            </v-card>
                        </v-col>
                    </v-row>

                    <v-row dense>
                        <v-col cols="12" md="6">
                            <v-text-field
                                v-model.number="settings.admin_chat_id"
                                :label="t('TELEGRAM_ADMIN_CHAT_ID')"
                                type="number"
                                variant="outlined"
                                density="comfortable"
                                :hint="t('TELEGRAM_ADMIN_CHAT_ID_HINT')"
                            />
                        </v-col>
                        <v-col cols="12" md="6">
                            <v-text-field
                                v-model.number="settings.low_quota_threshold_mb"
                                :label="t('TELEGRAM_LOW_QUOTA_THRESHOLD_MB')"
                                type="number"
                                min="10"
                                max="10240"
                                variant="outlined"
                                density="comfortable"
                            />
                        </v-col>
                        <v-col cols="12" md="6">
                            <v-select
                                v-model="settings.default_language"
                                :label="t('TELEGRAM_DEFAULT_LANGUAGE')"
                                :items="languages.map((l) => ({ value: l.code, title: l.label }))"
                                variant="outlined"
                                density="comfortable"
                            />
                        </v-col>
                        <v-col cols="12" md="6">
                            <v-text-field
                                v-model="settings.ocserv_host"
                                :label="t('TELEGRAM_OCSERV_HOST')"
                                variant="outlined"
                                density="comfortable"
                                :hint="t('TELEGRAM_OCSERV_HOST_HINT')"
                            />
                        </v-col>
                        <v-col cols="12" md="6">
                            <v-text-field
                                v-model="settings.card_number"
                                :label="t('TELEGRAM_CARD_NUMBER')"
                                variant="outlined"
                                density="comfortable"
                                :hint="t('TELEGRAM_CARD_NUMBER_HINT')"
                                persistent-hint
                            />
                        </v-col>
                        <v-col cols="12" md="6">
                            <v-text-field
                                v-model="settings.card_holder"
                                :label="t('TELEGRAM_CARD_HOLDER')"
                                variant="outlined"
                                density="comfortable"
                                :hint="t('TELEGRAM_CARD_HOLDER_HINT')"
                                persistent-hint
                            />
                        </v-col>
                        <v-col cols="12" md="6">
                            <v-text-field
                                v-model="settings.support_username"
                                :label="t('TELEGRAM_SUPPORT_USERNAME')"
                                placeholder="masniper"
                                prefix="@"
                                variant="outlined"
                                density="comfortable"
                                :hint="t('TELEGRAM_SUPPORT_USERNAME_HINT')"
                                persistent-hint
                            />
                        </v-col>
                    </v-row>

                    <v-row class="mt-2">
                        <v-col class="d-flex flex-wrap gap-3">
                            <v-btn type="submit" color="primary" :loading="loading">
                                {{ t('SAVE') }}
                            </v-btn>
                            <v-btn color="secondary" variant="outlined" :loading="loading" @click="sendTest">
                                {{ t('TELEGRAM_SEND_TEST') }}
                            </v-btn>
                            <v-btn variant="text" :loading="loading" @click="load">
                                {{ t('RELOAD') }}
                            </v-btn>
                        </v-col>
                    </v-row>
                </v-form>
            </UiParentCard>
        </v-col>
    </v-row>
</template>

<style scoped>
.telegram-token-input--masked :deep(input) {
    -webkit-text-security: disc;
}
</style>

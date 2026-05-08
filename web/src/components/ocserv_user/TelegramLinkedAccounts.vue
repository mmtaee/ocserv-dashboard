<script lang="ts" setup>
import { onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { TelegramAPI, type TelegramAccount } from '@/api/telegram';

const props = defineProps<{ uid: string }>();

const { t } = useI18n();
const accounts = ref<TelegramAccount[]>([]);
const loading = ref(false);

const load = async () => {
    if (!props.uid) return;
    loading.value = true;
    try {
        const res = await TelegramAPI.accountsForUser(props.uid);
        accounts.value = res.data || [];
    } finally {
        loading.value = false;
    }
};

const remove = async (id: number) => {
    if (!confirm(t('CONFIRM_DELETE'))) return;
    loading.value = true;
    try {
        await TelegramAPI.deleteAccount(id);
        await load();
    } finally {
        loading.value = false;
    }
};

watch(() => props.uid, load);
onMounted(load);
</script>

<template>
    <v-card variant="outlined" class="mt-4">
        <v-card-title class="text-h6 d-flex align-center">
            <v-icon class="me-2">mdi-robot</v-icon>
            {{ t('TELEGRAM_LINKED_ACCOUNTS') }}
        </v-card-title>
        <v-card-text>
            <div v-if="!accounts.length" class="text-grey">
                {{ t('TELEGRAM_NO_LINKED_ACCOUNTS') }}
            </div>
            <v-table v-else density="compact">
                <thead>
                    <tr>
                        <th>{{ t('CHAT_ID') }}</th>
                        <th>{{ t('TELEGRAM_USERNAME') }}</th>
                        <th>{{ t('LANGUAGE') }}</th>
                        <th>{{ t('CREATED_AT') }}</th>
                        <th>{{ t('ACTION') }}</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="a in accounts" :key="a.id">
                        <td>{{ a.chat_id }}</td>
                        <td>{{ a.telegram_username || '—' }}</td>
                        <td>{{ a.language }}</td>
                        <td>{{ new Date(a.created_at).toLocaleString() }}</td>
                        <td>
                            <v-btn
                                icon="mdi-link-variant-off"
                                color="error"
                                size="small"
                                variant="text"
                                :loading="loading"
                                @click="remove(a.id)"
                            />
                        </td>
                    </tr>
                </tbody>
            </v-table>
        </v-card-text>
    </v-card>
</template>

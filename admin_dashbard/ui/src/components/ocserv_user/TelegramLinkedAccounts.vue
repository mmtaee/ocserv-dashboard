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
    <div class="bg-surface shadow rounded-lg p-4 my-10">
        <h2 class="text-lg font-semibold mb-3 text-capitalize d-flex align-center">
            <v-icon class="me-2" color="primary">mdi-robot</v-icon>
            {{ t('TELEGRAM_LINKED_ACCOUNTS') }}
        </h2>

        <v-progress-linear :active="loading" indeterminate color="primary" class="mb-2" />

        <div v-if="!loading && !accounts.length" class="text-grey ms-5 text-capitalize">
            {{ t('TELEGRAM_NO_LINKED_ACCOUNTS') }}
        </div>

        <v-table v-else-if="accounts.length" density="compact" class="mx-3">
            <thead>
                <tr class="text-capitalize bg-lightprimary">
                    <th class="text-left">{{ t('CHAT_ID') }}</th>
                    <th class="text-left">{{ t('TELEGRAM_USERNAME') }}</th>
                    <th class="text-left">{{ t('LANGUAGE') }}</th>
                    <th class="text-left">{{ t('CREATED_AT') }}</th>
                    <th class="text-left">{{ t('ACTION') }}</th>
                </tr>
            </thead>
            <tbody>
                <tr v-for="a in accounts" :key="a.id">
                    <td>
                        <code>{{ a.chat_id }}</code>
                    </td>
                    <td>
                        <a
                            v-if="a.telegram_username"
                            :href="`https://t.me/${a.telegram_username}`"
                            target="_blank"
                            rel="noopener noreferrer"
                            class="text-primary"
                        >
                            @{{ a.telegram_username }}
                        </a>
                        <span v-else class="text-grey text-caption">{{ t('TELEGRAM_USERNAME_NOT_PUBLIC') }}</span>
                    </td>
                    <td class="text-uppercase">{{ a.language }}</td>
                    <td>{{ new Date(a.created_at).toLocaleString() }}</td>
                    <td>
                        <v-tooltip :text="t('TELEGRAM_UNLINK_ACCOUNT')">
                            <template #activator="{ props: tip }">
                                <v-btn
                                    v-bind="tip"
                                    icon="mdi-link-variant-off"
                                    color="error"
                                    size="small"
                                    variant="text"
                                    :loading="loading"
                                    @click="remove(a.id)"
                                />
                            </template>
                        </v-tooltip>
                    </td>
                </tr>
            </tbody>
        </v-table>
    </div>
</template>

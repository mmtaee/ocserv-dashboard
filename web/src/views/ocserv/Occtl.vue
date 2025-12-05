<script lang="ts" setup>
import UiParentCard from '@/components/shared/UiParentCard.vue';
import { useI18n } from 'vue-i18n';
import { computed, reactive, ref } from 'vue';
import { OCCTLApi } from '@/api';
import { getAuthorization } from '@/utils/request';
import { useServerStore } from '@/stores/config';
import UiChildCard from '@/components/shared/UiChildCard.vue';
import { useSnackbarStore } from '@/stores/snackbar';

const { t } = useI18n();
const data = reactive({
    action: null,
    value: ''
});
const result = ref<string>('');
const isExecuted = ref(false);
const commands = [
    { id: 1, command: 'Show users', description: t('SHOW_USERS_DESC') },
    { id: 2, command: 'Show user by username', description: t('SHOW_USER_DETAIL'), value: t('OCSERV_USERNAME') },
    { id: 3, command: 'Show user by id', description: t('SHOW_ID_DESC'), value: t('OCSERV_USER_ID') },
    { id: 4, command: 'Disconnect user', description: t('DISCONNECT_USER_DESC'), value: t('OCSERV_USERNAME') },
    { id: 5, command: 'Show sessions all', description: t('SHOW_SESSIONS_ALL_DESC') },
    { id: 6, command: 'Show sessions valid', description: t('SHOW_SESSIONS_VALID_DESC') },
    { id: 7, command: 'Show session by SID', description: t('SHOW_SESSION_DESC'), value: 'SID' },
    { id: 8, command: 'Show ip ban points', description: t('SHOW_IP_BANS_DESC') },
    { id: 9, command: 'Unban ip', description: t('UNBAN_IP_DESC'), value: 'IP' },
    { id: 10, command: 'Show status', description: t('SHOW_STATUS_DESC') },
    { id: 11, command: 'Show events', description: t('SHOW_EVENTS_DESC') },
    { id: 12, command: 'Show iroutes', description: t('SHOW_IROUTES_DESC') },
    { id: 13, command: 'Reload', description: t('RELOAD_DESC') }
];

const serverStore = useServerStore();
const checkDisableValue = computed(() => {
    if (!data.action) return true;

    const selected = commands.find((c) => c.id === data.action);
    if (selected?.value) {
        return false;
    }
    data.value = '';
    return true;
});

const checkDisableBtn = computed(() => {
    if (!data.action) return true;

    const selected = commands.find((c) => c.id === data.action);
    return !!(selected?.value && data.value == '');
});

const actionHint = computed(() => {
    if (data.action) {
        const selected = commands.find((c) => c.id === data.action);
        return selected ? selected.description : '';
    }
    return t('OCCTL_ACTION_COMMAND');
});

const valueHint = computed(() => {
    if (data.action) {
        const selected = commands.find((c) => c.id === data.action);
        return selected ? selected.value : '';
    }
    return t('OCCTL_ACTION_VALUE');
});

const reloadDisabled = ref(false);
let reloadCooldownEnd = 0;

const execute = () => {
    const now = Date.now();

    if (data.action == 13) {
        let blockTime = 120000;

        if (reloadDisabled.value) {
            const remainingSeconds = Math.ceil((reloadCooldownEnd - now) / 1000);
            const snackbar = useSnackbarStore();
            snackbar.show({
                id: 1,
                message: t('RELOADING_DISABLED_DESC') + ` (${remainingSeconds} ` + t('SECONDS_REMAINING') + ')',
                color: 'warning',
                timeout: 3000
            });
            return;
        }

        reloadDisabled.value = true;
        reloadCooldownEnd = now + blockTime;

        setTimeout(() => {
            reloadDisabled.value = false;
        }, blockTime);
    }

    const api = new OCCTLApi();
    api.occtlCommandsGet({
        ...getAuthorization(),
        action: data.action || 0,
        value: data.value || ''
    })
        .then((res) => {
            if (res.data == 'null') return;
            try {
                result.value = JSON.stringify(JSON.parse(res.data), null, 4);
            } catch (e) {
                console.error(e);
            }
        })
        .finally(() => {
            isExecuted.value = true;
        });
};
</script>
<template>
    <v-row>
        <v-col cols="12" md="12">
            <UiParentCard :title="`OCCTL ${t('HANDLER')}`">
                <div class="pa-7 pt-1">
                    <p class="text-body-1 text-muted" v-html="serverStore.occtlVersionInfo"></p>
                </div>

                <UiChildCard>
                    <v-row align="start" class="mb-1" justify="center">
                        <v-col cols="12" md="3">
                            <v-label class="font-weight-bold mb-1 text-capitalize">{{ t('ACTION') }}</v-label>
                            <v-select
                                v-model="data.action"
                                :hint="actionHint"
                                :items="commands"
                                color="primary"
                                item-title="command"
                                item-value="id"
                                persistent-hint
                                variant="outlined"
                                @update:model-value="data.value = ''"
                            />
                        </v-col>
                        <v-col cols="12" md="3">
                            <v-label class="font-weight-bold mb-1 text-capitalize">{{ t('VALUE') }}</v-label>
                            <v-text-field
                                v-model="data.value"
                                :disabled="checkDisableValue"
                                :hint="valueHint"
                                color="primary"
                                persistent-hint
                                variant="outlined"
                            />
                        </v-col>
                        <v-col align-self="center" cols="12" md="2">
                            <v-btn :disabled="checkDisableBtn" color="primary" @click="execute">
                                {{ t('EXECUTE') }}
                            </v-btn>
                        </v-col>
                    </v-row>
                </UiChildCard>

                <UiChildCard>
                    <v-sheet
                        v-if="isExecuted"
                        class="text-pre-wrap pa-10 bg-lightsecondary"
                        style="overflow: auto; white-space: pre; height: 500px"
                    >
                        {{ result }}
                    </v-sheet>
                </UiChildCard>
            </UiParentCard>
        </v-col>
    </v-row>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { onBeforeUnmount, onMounted, ref, watch, type PropType } from 'vue';
import { SystemdApi } from '@/api';
import { getAuthorization } from '@/utils/request';
import { useSnackbarStore } from '@/stores/snackbar';

type ActionType = 'restart' | 'enable' | 'disable' | null;
type StatusEmit = 'restarting' | 'enabling' | 'disabling' | null;

type CurrentAction = {
    value: ActionType;
    expiresAt: number;
};

const props = defineProps({
    state: {
        type: String as PropType<'active' | 'inactive' | 'failed' | 'activating' | 'deactivating'>,
        default: 'inactive'
    }
});

const emit = defineEmits<{
    (e: 'getState'): void;
    (e: 'currentStatus', value: StatusEmit): void;
}>();

const { t } = useI18n();
const api = new SystemdApi();
const snackbar = useSnackbarStore();

const currentAction = ref<ActionType>(null);
const remainingTime = ref<number>(0);

const storageKey = 'current_action';
const TTL = 2 * 60 * 1000; // 2 minutes

let countdownInterval: ReturnType<typeof setInterval> | null = null;
let hasExpired = false;

const status = () => {
    emit('getState');
};

// ✅ single source of truth for expiration
const handleExpire = () => {
    localStorage.removeItem(storageKey);
    currentAction.value = null;
    remainingTime.value = 0;

    if (!hasExpired) {
        hasExpired = true;
        emit('currentStatus', null);
        emit('getState');
    }
};

const validateStorage = () => {
    const raw = localStorage.getItem(storageKey);

    if (!raw) {
        handleExpire();
        return;
    }

    try {
        const data: CurrentAction = JSON.parse(raw);
        const diff = data.expiresAt - Date.now();

        if (diff <= 0) {
            handleExpire();
            return;
        }

        currentAction.value = data.value;
        remainingTime.value = Math.ceil(diff / 1000);
    } catch {
        handleExpire();
    }
};

const setCurrentAction = (value: ActionType) => {
    const payload: CurrentAction = {
        value,
        expiresAt: Date.now() + TTL
    };

    localStorage.setItem(storageKey, JSON.stringify(payload));
    currentAction.value = value;
    hasExpired = false;
};

const restart = () => {
    if (props.state !== 'active') return;

    api.systemdRestartPost({
        ...getAuthorization()
    }).then((res) => {
        snackbar.show({
            id: 1,
            message: res.data.message,
            color: 'info',
            timeout: 5000
        });

        emit('currentStatus', 'restarting');
        setCurrentAction('restart');
    });
};

const enable = () => {
    if (props.state !== 'inactive') return;

    api.systemdEnablePost({
        ...getAuthorization()
    }).then((res) => {
        snackbar.show({
            id: 1,
            message: res.data.message,
            color: 'info',
            timeout: 5000
        });

        emit('currentStatus', 'enabling');
        setCurrentAction('enable');
    });
};

const disable = () => {
    if (props.state !== 'active') return;

    api.systemdDisablePost({
        ...getAuthorization()
    }).then((res) => {
        snackbar.show({
            id: 1,
            message: res.data.message,
            color: 'info',
            timeout: 5000
        });

        emit('currentStatus', 'disabling');
        setCurrentAction('disable');
    });
};

onMounted(() => {
    validateStorage();
});

watch(
    currentAction,
    (val) => {
        if (countdownInterval) clearInterval(countdownInterval);
        countdownInterval = null;

        if (!val) {
            hasExpired = false;
            return;
        }

        countdownInterval = setInterval(() => {
            const raw = localStorage.getItem(storageKey);

            if (!raw) {
                handleExpire();
                return;
            }

            try {
                const data: CurrentAction = JSON.parse(raw);
                const diff = data.expiresAt - Date.now();

                if (diff <= 0) {
                    handleExpire();
                    return;
                }

                remainingTime.value = Math.ceil(diff / 1000);
            } catch {
                handleExpire();
            }
        }, 1000);
    },
    { immediate: true }
);

onBeforeUnmount(() => {
    if (countdownInterval) clearInterval(countdownInterval);
});
</script>

<template>
    <v-row align="center" justify="center" v-if="currentAction == null">
        <v-col cols="12" md="auto" v-if="props.state == 'active'">
            <v-btn @click="status" color="info"> {{ t('STATUS') }} </v-btn>
        </v-col>

        <v-col cols="12" md="auto" v-if="props.state == 'active'">
            <v-btn @click="restart" color="primary">
                {{ t('RESTART') }}
            </v-btn>
        </v-col>

        <v-col cols="12" md="auto" v-if="state == 'inactive'">
            <v-btn @click="enable" color="success">
                {{ t('ENABLE') }}
            </v-btn>
        </v-col>

        <v-col cols="12" md="auto" v-if="state == 'active'">
            <v-btn @click="disable" color="error">
                {{ t('DISABLE') }}
            </v-btn>
        </v-col>
    </v-row>

    <div v-else class="text-center">
        <div class="text-h6 text-primary text-capitalize">{{ currentAction }}ing ...</div>
        <div v-if="remainingTime > 0" class="text-h6 my-3" style="color: #888888">
            {{ t('PLEASE_WAIT_UNTIL') }}: {{ remainingTime }}s
        </div>
    </div>
</template>

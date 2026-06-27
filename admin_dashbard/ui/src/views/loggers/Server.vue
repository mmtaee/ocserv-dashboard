<script lang="ts" setup>
import UiParentCard from '@/components/shared/UiParentCard.vue';
import { useI18n } from 'vue-i18n';
import UiChildCard from '@/components/shared/UiChildCard.vue';
import { nextTick, onMounted, onUnmounted, ref, watch } from 'vue';

const { t } = useI18n();
const isConnected = ref(false);
const btnDisable = ref(false);
const logs = ref<string[]>([]);
const logContainer = ref<HTMLElement | null>(null);

let eventSource: EventSource | null = null;

const host = window.location.host; // includes hostname:port
const protocol = window.location.protocol;
const SSE_URL = import.meta.env.VITE_LOG_SOCKET_URL || `${protocol}//${host}/ws/logs`;

const connected = ref(false);
const containerHeight = ref(window.innerHeight);
const maxLogs = ref(0);

const updateMaxLogs = () => {
    containerHeight.value = window.innerHeight - 325;
    maxLogs.value = Math.floor(containerHeight.value / 25);
    if (logs.value.length > maxLogs.value) {
        logs.value.splice(0, logs.value.length - maxLogs.value);
    }
};

const addLog = async (newLog: string) => {
    if (logs.value.length >= maxLogs.value) {
        logs.value.splice(0, 2);
    }
    logs.value.push(newLog);
};

const connect = () => {
    if (connected.value) return;
    btnDisable.value = true;
    logs.value = [];
    addLog(t('START_CONNECTING') + '...');

    setTimeout(() => {
        eventSource = new EventSource(SSE_URL);
        eventSource.onmessage = (event) => {
            addLog(event.data);
        };
        eventSource.onerror = (error) => {
            console.error('EventSource error:', error);
            disconnect();
        };
        isConnected.value = true;
        addLog(t('SERVER_CONNECTED_MSG'));
        btnDisable.value = false;
    }, 3000);
};

const disconnect = () => {
    btnDisable.value = true;
    addLog(t('SERVER_DISCONNECTING') + '...');
    eventSource?.close();
    eventSource = null;

    setTimeout(() => {
        logs.value = [];
        addLog(t('SERVER_DISCONNECTED_MSG'));
        isConnected.value = false;
        btnDisable.value = false;
    }, 3000);
};

onMounted(() => {
    addLog(t('SERVER_DISCONNECTED_MSG'));
    updateMaxLogs();
    window.addEventListener('resize', updateMaxLogs);
});

onUnmounted(() => {
    window.removeEventListener('resize', updateMaxLogs);
    disconnect();
});

watch(
    () => logs.value.length,
    async () => {
        await nextTick();
        if (logContainer.value) {
            logContainer.value.scrollTop = logContainer.value.scrollHeight;
        }
    }
);
</script>

<template>
    <v-row>
        <v-col cols="12" md="12">
            <UiParentCard :title="t('LIVE_SERVER_LOG')">
                <UiChildCard>
                    <v-row align="center" justify="start">
                        <v-col cols="12" md="auto">
                            <span class="text-capitalize">{{ t('LIVE_SERVER_TEXT_HELP') }}</span>
                        </v-col>
                        <v-col cols="12" md="2">
                            <v-btn
                                v-if="!isConnected"
                                :disabled="btnDisable"
                                class="me-lg-5"
                                color="primary"
                                size="small"
                                variant="flat"
                                @click="connect"
                            >
                                {{ t('CONNECT') }}
                            </v-btn>
                            <v-btn
                                v-if="isConnected"
                                :disabled="btnDisable"
                                class="me-lg-5"
                                color="error"
                                size="small"
                                variant="flat"
                                @click="disconnect"
                            >
                                {{ t('DISCONNECT') }}
                            </v-btn>
                        </v-col>
                    </v-row>
                </UiChildCard>

                <UiChildCard>
                    <v-card
                        v-if="logs.length > 0"
                        ref="logContainer"
                        class="pa-4 overflow-auto font-mono bg-muted text-white"
                        height="520"
                    >
                        <div v-for="(log, index) in logs" :key="index">
                            {{ log }}
                        </div>
                    </v-card>

                    <v-card v-else height="520"></v-card>
                </UiChildCard>
            </UiParentCard>
        </v-col>
    </v-row>
</template>

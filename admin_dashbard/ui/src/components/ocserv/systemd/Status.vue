<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { SystemdApi, type SystemdOcservSystemdStatus } from '@/api';
import { getAuthorization } from '@/utils/request';
import { computed, onMounted, type PropType, ref, watch } from 'vue';

const props = defineProps({
    currentStatus: {
        type: String as PropType<'enabling' | 'disabling' | 'restarting' | null>,
        default: null
    }
});

const emit = defineEmits(['state']);

const { t } = useI18n();

const service = ref<SystemdOcservSystemdStatus>({});

const getStatus = () => {
    const api = new SystemdApi();

    api.systemdStatusGet({
        ...getAuthorization()
    }).then((res) => {
        Object.assign(service.value, res.data);
        emit('state', res.data.active_state);
    });
};

onMounted(() => {
    getStatus();
});

defineExpose({
    getStatus
});

// status color
const statusColor = computed(() => {
    const active = service.value.active_state;
    const unit = service.value.unit_file_state;

    if (active === 'active') return 'success';
    if (active === 'activating' || active === 'restarting') return 'info';
    if (active === 'deactivating') return 'warning';
    if (active === 'failed') return 'error';
    if (active === 'inactive' && unit === 'disabled') return 'grey';
    if (active === 'enabling' || active === 'disabling') return 'info';
    if (unit === 'masked') return 'black';

    return 'secondary';
});

const displayState = computed(() => {
    if (props.currentStatus === 'restarting') return 'restarting';
    if (props.currentStatus === 'enabling') return 'enabling';
    if (props.currentStatus === 'disabling') return 'disabling';

    return service.value.active_state;
});

const formatMemory = (bytes?: number) => {
    if (!bytes) return '-';

    const mb = bytes / 1024 / 1024;
    return `${mb.toFixed(1)} MB`;
};

const formatCPU = (ns?: number) => {
    if (!ns) return '-';

    const sec = ns / 1e9;
    return `${sec.toFixed(2)} s`;
};

watch(
    () => props.currentStatus,
    (val, oldVal) => {
        // when action completes → reload real state
        if (oldVal && val === null) {
            getStatus();
        }
    }
);
</script>

<template>
    <v-row>
        <v-col cols="12" md="12">
            <!-- HEADER -->
            <v-row align="center" justify="space-between" class="mt-2 text-capitalize">
                <v-col class="ma-0 pa-0 ms-6">
                    {{ service.description }}
                </v-col>
                <v-col class="ma-0 pa-0 me-5">
                    <v-chip :color="statusColor">
                        {{ displayState }}
                    </v-chip>
                </v-col>
            </v-row>

            <v-divider class="my-3" />

            <!-- GRID INFO -->
            <v-row dense>
                <v-col cols="12" md="6">
                    <v-list>
                        <v-list-item class="mb-3">
                            <v-list-item-title>ID</v-list-item-title>
                            <v-list-item-subtitle>{{ service.id }}</v-list-item-subtitle>
                        </v-list-item>

                        <v-list-item class="mb-3">
                            <v-list-item-title class="text-capitalize">sub state</v-list-item-title>
                            <v-list-item-subtitle>{{ service.sub_state }}</v-list-item-subtitle>
                        </v-list-item>

                        <v-list-item class="mb-3">
                            <v-list-item-title class="text-capitalize">unit state</v-list-item-title>
                            <v-list-item-subtitle>{{ service.unit_file_state }}</v-list-item-subtitle>
                        </v-list-item>

                        <v-list-item class="mb-3">
                            <v-list-item-title class="text-capitalize">start time</v-list-item-title>
                            <v-list-item-subtitle>{{ service.start_time }}</v-list-item-subtitle>
                        </v-list-item>
                    </v-list>
                </v-col>

                <v-col cols="12" md="6">
                    <v-list>
                        <v-list-item class="mb-3">
                            <v-list-item-title>PID</v-list-item-title>
                            <v-list-item-subtitle>{{ service.main_pid }}</v-list-item-subtitle>
                        </v-list-item>

                        <v-list-item class="mb-3">
                            <v-list-item-title class="text-capitalize">memory</v-list-item-title>
                            <v-list-item-subtitle>
                                {{ formatMemory(service.memory) }}
                            </v-list-item-subtitle>
                        </v-list-item>

                        <v-list-item class="mb-3">
                            <v-list-item-title class="text-capitalize">cpu usage</v-list-item-title>
                            <v-list-item-subtitle>
                                {{ formatCPU(service.cpu_usage_nsec) }}
                            </v-list-item-subtitle>
                        </v-list-item>

                        <v-list-item class="mb-3">
                            <v-list-item-title class="text-capitalize">tasks</v-list-item-title>
                            <v-list-item-subtitle>{{ service.tasks }}</v-list-item-subtitle>
                        </v-list-item>
                    </v-list>
                </v-col>
            </v-row>
        </v-col>
    </v-row>
</template>

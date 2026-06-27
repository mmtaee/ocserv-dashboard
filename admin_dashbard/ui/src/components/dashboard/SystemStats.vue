<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { onMounted, onUnmounted, ref } from 'vue';
import { HomeApi, type HomeDockerService, type HomeServerStatusResponse } from '@/api';
import { getAuthorization } from '@/utils/request';
import { useConfigStore } from '@/stores/config';

const { t } = useI18n();

const configStore = useConfigStore();

const api = new HomeApi();

const isDockerMod = import.meta.env.VITE_SYSTEMD != 'true';
const systemUsage = ref<HomeServerStatusResponse>({});
const dockerUsage = ref<HomeDockerService>(<HomeDockerService>{
    log_stream: {},
    ocserv: {},
    postgres: {},
    user_expiry: {},
    web: {}
});

const getSystemUsage = () => {
    api.homeSystemStatsGet(
        {
            ...getAuthorization()
        },
        {
            headers: {
                'x-skip-loading': 'true'
            }
        }
    ).then((res) => {
        Object.assign(systemUsage.value, res.data);
    });
};

const getDockerUsage = () => {
    api.homeContainerStatsGet(
        {
            ...getAuthorization()
        },
        {
            headers: {
                'x-skip-loading': 'true'
            }
        }
    ).then((res) => {
        Object.assign(dockerUsage.value, res.data);
    });
};

let intervalId: ReturnType<typeof setInterval>;

onMounted(() => {
    // run immediately
    getSystemUsage();

    if (isDockerMod) {
        getDockerUsage();
    }

    // run every 10 seconds
    intervalId = setInterval(() => {
        getSystemUsage();

        if (isDockerMod) {
            getDockerUsage();
        }
    }, 10000);
});

onUnmounted(() => {
    clearInterval(intervalId);
});
</script>

<template>
    <!-- SYSTEM STATS -->
    <v-card elevation="10">
        <v-card-item>
            <v-card-title class="text-h5 text-capitalize">{{ t('SYSTEM_USAGE') }}</v-card-title>
            <v-card-text>
                <v-row align="center" justify="center" class="mt-3">
                    <!-- CPU USAGE -->
                    <v-col cols="12" lg="2" sm="6">
                        <v-row class="text-center">
                            <v-col cols="12">
                                <v-progress-circular
                                    :model-value="systemUsage.cpu?.avg_percent"
                                    :size="100"
                                    :width="12"
                                    bg-color="grey100"
                                    color="primary"
                                    reveal
                                    rounded
                                >
                                    <v-avatar color="surface-light" size="70">
                                        {{ systemUsage.cpu?.avg_percent }}%
                                    </v-avatar>
                                </v-progress-circular>
                            </v-col>
                            <v-col cols="12">
                                <div class="text-subtitle-2 text-capitalize">{{ t('CPU_USAGE') }}</div>
                                <div class="text-subtitle-2" dir="LTR">
                                    <span class="me-1">{{ systemUsage.cpu?.used_units }}</span>
                                    /
                                    <span class="ms-1">{{ systemUsage.cpu?.total }} </span>
                                    {{ t('UNITS') }}
                                </div>
                            </v-col>
                        </v-row>
                    </v-col>

                    <!-- RAM USAGE -->
                    <v-col cols="12" lg="2" sm="6">
                        <v-row class="text-center">
                            <v-col cols="12">
                                <v-progress-circular
                                    :model-value="systemUsage.ram?.used_percent"
                                    :size="100"
                                    :width="12"
                                    bg-color="grey100"
                                    color="primary"
                                    reveal
                                    rounded
                                >
                                    <v-avatar color="surface-light" size="70">
                                        {{ systemUsage.ram?.used_percent }}%
                                    </v-avatar>
                                </v-progress-circular>
                            </v-col>
                            <v-col cols="12">
                                <div class="text-subtitle-2 text-capitalize">{{ t('RAM_USAGE') }}</div>
                                <div class="text-subtitle-2" dir="LTR">
                                    <span class="me-1">{{ systemUsage.ram?.used }}</span>
                                    /
                                    <span class="ms-1">{{ systemUsage.ram?.total }} </span>
                                    GB
                                </div>
                            </v-col>
                        </v-row>
                    </v-col>

                    <!-- SWAP USAGE -->
                    <v-col cols="12" lg="2" sm="6">
                        <v-row class="text-center">
                            <v-col cols="12">
                                <v-progress-circular
                                    :model-value="systemUsage.swap?.used_percent"
                                    :size="100"
                                    :width="12"
                                    bg-color="grey100"
                                    color="primary"
                                    reveal
                                    rounded
                                >
                                    <v-avatar color="surface-light" size="70">
                                        {{ systemUsage.swap?.used_percent }}%
                                    </v-avatar>
                                </v-progress-circular>
                            </v-col>
                            <v-col cols="12">
                                <div class="text-subtitle-2 text-capitalize">{{ t('SWAP_USAGE') }}</div>
                                <div class="text-subtitle-2" dir="LTR">
                                    <span class="me-1">{{ systemUsage.swap?.used }}</span>
                                    /
                                    <span class="ms-1">{{ systemUsage.swap?.total }} </span>
                                    GB
                                </div>
                            </v-col>
                        </v-row>
                    </v-col>

                    <!-- DISK USAGE -->
                    <v-col cols="12" lg="2" sm="6">
                        <v-row class="text-center">
                            <v-col cols="12">
                                <v-progress-circular
                                    :model-value="systemUsage.disk?.used_percent"
                                    :size="100"
                                    :width="12"
                                    bg-color="grey100"
                                    color="primary"
                                    reveal
                                    rounded
                                >
                                    <v-avatar color="surface-light" size="70">
                                        {{ systemUsage.disk?.used_percent }}%
                                    </v-avatar>
                                </v-progress-circular>
                            </v-col>
                            <v-col cols="12">
                                <div class="text-subtitle-2 text-capitalize">{{ t('DISK_USAGE') }}</div>
                                <div class="text-subtitle-2" dir="LTR">
                                    <span class="me-1">{{ systemUsage.disk?.used }}</span>
                                    /
                                    <span class="ms-1">{{ systemUsage.disk?.total }} </span>
                                    GB
                                </div>
                            </v-col>
                        </v-row>
                    </v-col>
                </v-row>
            </v-card-text>
        </v-card-item>
    </v-card>

    <!-- DOCKER CONTAINERS STATS -->
    <v-card elevation="10" class="mt-5" v-if="isDockerMod">
        <v-card-item>
            <v-card-title class="my-lg-4 text-capitalize text-h5">{{ t('CONTAINERS_USAGE') }}</v-card-title>
            <v-row>
                <!-- OCSERV CONTAINER -->
                <v-col cols="12" lg="4" class="mb-lg-4">
                    <v-card-subtitle class="text-h5 text-center"> Ocserv </v-card-subtitle>
                    <v-row align="center" justify="center">
                        <!-- CPU USAGE -->
                        <v-col cols="12" lg="4" sm="6">
                            <v-row class="text-center">
                                <v-col cols="12">
                                    <v-progress-circular
                                        :model-value="dockerUsage.ocserv?.cpu?.avg_percent"
                                        :size="100"
                                        :width="12"
                                        bg-color="grey100"
                                        color="primary"
                                        reveal
                                        rounded
                                    >
                                        <v-avatar color="surface-light" size="70">
                                            {{ dockerUsage.ocserv?.cpu?.avg_percent }}%
                                        </v-avatar>
                                    </v-progress-circular>
                                </v-col>
                                <v-col cols="12">
                                    <div class="text-subtitle-2 text-capitalize">{{ t('CPU_USAGE') }}</div>
                                    <div class="text-subtitle-2" dir="LTR">
                                        <span class="me-1">{{ dockerUsage.ocserv?.cpu?.used_units }}</span>
                                        /
                                        <span class="ms-1">{{ dockerUsage.ocserv?.cpu?.total }} </span>
                                        {{ t('UNITS') }}
                                    </div>
                                </v-col>
                            </v-row>
                        </v-col>

                        <!-- RAM USAGE -->
                        <v-col cols="12" lg="4" sm="6">
                            <v-row class="text-center">
                                <v-col cols="12">
                                    <v-progress-circular
                                        :model-value="dockerUsage.ocserv?.ram?.used_percent"
                                        :size="100"
                                        :width="12"
                                        bg-color="grey100"
                                        color="primary"
                                        reveal
                                        rounded
                                    >
                                        <v-avatar color="surface-light" size="70">
                                            {{ dockerUsage.ocserv?.ram?.used_percent }}%
                                        </v-avatar>
                                    </v-progress-circular>
                                </v-col>
                                <v-col cols="12">
                                    <div class="text-subtitle-2 text-capitalize">{{ t('RAM_USAGE') }}</div>
                                    <div class="text-subtitle-2" dir="LTR">
                                        <span class="me-1">{{ dockerUsage.ocserv?.ram?.used }}</span>
                                        /
                                        <span class="ms-1">{{ dockerUsage.ocserv?.ram?.total }} </span>
                                        GB
                                    </div>
                                </v-col>
                            </v-row>
                        </v-col>
                    </v-row>
                </v-col>

                <!-- LOG STREAM CONTAINER -->
                <v-col cols="12" lg="4" class="mb-lg-4">
                    <v-card-subtitle class="text-h5 text-center"> Log stream </v-card-subtitle>
                    <v-row align="center" justify="center">
                        <!-- CPU USAGE -->
                        <v-col cols="12" lg="4" sm="6">
                            <v-row class="text-center">
                                <v-col cols="12">
                                    <v-progress-circular
                                        :model-value="dockerUsage.log_stream?.cpu?.avg_percent"
                                        :size="100"
                                        :width="12"
                                        bg-color="grey100"
                                        color="primary"
                                        reveal
                                        rounded
                                    >
                                        <v-avatar color="surface-light" size="70">
                                            {{ dockerUsage.log_stream?.cpu?.avg_percent }}%
                                        </v-avatar>
                                    </v-progress-circular>
                                </v-col>
                                <v-col cols="12">
                                    <div class="text-subtitle-2 text-capitalize">{{ t('CPU_USAGE') }}</div>
                                    <div class="text-subtitle-2" dir="LTR">
                                        <span class="me-1">{{ dockerUsage.log_stream?.cpu?.used_units }}</span>
                                        /
                                        <span class="ms-1">{{ dockerUsage.log_stream?.cpu?.total }} </span>
                                        {{ t('UNITS') }}
                                    </div>
                                </v-col>
                            </v-row>
                        </v-col>

                        <!-- RAM USAGE -->
                        <v-col cols="12" lg="4" sm="6">
                            <v-row class="text-center">
                                <v-col cols="12">
                                    <v-progress-circular
                                        :model-value="dockerUsage.log_stream?.ram?.used_percent"
                                        :size="100"
                                        :width="12"
                                        bg-color="grey100"
                                        color="primary"
                                        reveal
                                        rounded
                                    >
                                        <v-avatar color="surface-light" size="70">
                                            {{ dockerUsage.log_stream?.ram?.used_percent }}%
                                        </v-avatar>
                                    </v-progress-circular>
                                </v-col>
                                <v-col cols="12">
                                    <div class="text-subtitle-2 text-capitalize">{{ t('RAM_USAGE') }}</div>
                                    <div class="text-subtitle-2" dir="LTR">
                                        <span class="me-1">{{ dockerUsage.log_stream?.ram?.used }}</span>
                                        /
                                        <span class="ms-1">{{ dockerUsage.log_stream?.ram?.total }} </span>
                                        GB
                                    </div>
                                </v-col>
                            </v-row>
                        </v-col>
                    </v-row>
                </v-col>

                <!-- USER EXPIRY CONTAINER -->
                <v-col cols="12" lg="4" class="mb-lg-4">
                    <v-card-subtitle class="text-h5 text-center"> User Expiry </v-card-subtitle>
                    <v-row align="center" justify="center">
                        <!-- CPU USAGE -->
                        <v-col cols="12" lg="4" sm="6">
                            <v-row class="text-center">
                                <v-col cols="12">
                                    <v-progress-circular
                                        :model-value="dockerUsage.user_expiry?.cpu?.avg_percent"
                                        :size="100"
                                        :width="12"
                                        bg-color="grey100"
                                        color="primary"
                                        reveal
                                        rounded
                                    >
                                        <v-avatar color="surface-light" size="70">
                                            {{ dockerUsage.user_expiry?.cpu?.avg_percent }}%
                                        </v-avatar>
                                    </v-progress-circular>
                                </v-col>
                                <v-col cols="12">
                                    <div class="text-subtitle-2 text-capitalize">{{ t('CPU_USAGE') }}</div>
                                    <div class="text-subtitle-2" dir="LTR">
                                        <span class="me-1">{{ dockerUsage.user_expiry?.cpu?.used_units }}</span>
                                        /
                                        <span class="ms-1">{{ dockerUsage.user_expiry?.cpu?.total }} </span>
                                        {{ t('UNITS') }}
                                    </div>
                                </v-col>
                            </v-row>
                        </v-col>

                        <!-- RAM USAGE -->
                        <v-col cols="12" lg="4" sm="6">
                            <v-row class="text-center">
                                <v-col cols="12">
                                    <v-progress-circular
                                        :model-value="dockerUsage.user_expiry?.ram?.used_percent"
                                        :size="100"
                                        :width="12"
                                        bg-color="grey100"
                                        color="primary"
                                        reveal
                                        rounded
                                    >
                                        <v-avatar color="surface-light" size="70">
                                            {{ dockerUsage.user_expiry?.ram?.used_percent }}%
                                        </v-avatar>
                                    </v-progress-circular>
                                </v-col>
                                <v-col cols="12">
                                    <div class="text-subtitle-2 text-capitalize">{{ t('RAM_USAGE') }}</div>
                                    <div class="text-subtitle-2" dir="LTR">
                                        <span class="me-1">{{ dockerUsage.user_expiry?.ram?.used }}</span>
                                        /
                                        <span class="ms-1">{{ dockerUsage.user_expiry?.ram?.total }} </span>
                                        GB
                                    </div>
                                </v-col>
                            </v-row>
                        </v-col>
                    </v-row>
                </v-col>

                <!-- TELEGRAM BOT CONTAINER -->
                <v-col cols="12" lg="4" class="mb-lg-4" v-if="configStore.telegramBotEnabled">
                    <v-card-subtitle class="text-h5 text-center"> Telegram Bot </v-card-subtitle>
                    <v-row align="center" justify="center">
                        <!-- CPU USAGE -->
                        <v-col cols="12" lg="4" sm="6">
                            <v-row class="text-center">
                                <v-col cols="12">
                                    <v-progress-circular
                                        :model-value="dockerUsage.telegram_bot?.cpu?.avg_percent"
                                        :size="100"
                                        :width="12"
                                        bg-color="grey100"
                                        color="primary"
                                        reveal
                                        rounded
                                    >
                                        <v-avatar color="surface-light" size="70">
                                            {{ dockerUsage.telegram_bot?.cpu?.avg_percent }}%
                                        </v-avatar>
                                    </v-progress-circular>
                                </v-col>
                                <v-col cols="12">
                                    <div class="text-subtitle-2 text-capitalize">{{ t('CPU_USAGE') }}</div>
                                    <div class="text-subtitle-2" dir="LTR">
                                        <span class="me-1">{{ dockerUsage.telegram_bot?.cpu?.used_units }}</span>
                                        /
                                        <span class="ms-1">{{ dockerUsage.telegram_bot?.cpu?.total }} </span>
                                        {{ t('UNITS') }}
                                    </div>
                                </v-col>
                            </v-row>
                        </v-col>

                        <!-- RAM USAGE -->
                        <v-col cols="12" lg="4" sm="6">
                            <v-row class="text-center">
                                <v-col cols="12">
                                    <v-progress-circular
                                        :model-value="dockerUsage.user_expiry?.ram?.used_percent"
                                        :size="100"
                                        :width="12"
                                        bg-color="grey100"
                                        color="primary"
                                        reveal
                                        rounded
                                    >
                                        <v-avatar color="surface-light" size="70">
                                            {{ dockerUsage.user_expiry?.ram?.used_percent }}%
                                        </v-avatar>
                                    </v-progress-circular>
                                </v-col>
                                <v-col cols="12">
                                    <div class="text-subtitle-2 text-capitalize">{{ t('RAM_USAGE') }}</div>
                                    <div class="text-subtitle-2" dir="LTR">
                                        <span class="me-1">{{ dockerUsage.user_expiry?.ram?.used }}</span>
                                        /
                                        <span class="ms-1">{{ dockerUsage.user_expiry?.ram?.total }} </span>
                                        GB
                                    </div>
                                </v-col>
                            </v-row>
                        </v-col>
                    </v-row>
                </v-col>

                <!-- POSTGRES CONTAINER -->
                <v-col cols="12" lg="4" class="mb-lg-4">
                    <v-card-subtitle class="text-h5 text-center"> PostgreSQL </v-card-subtitle>
                    <v-row align="center" justify="center">
                        <!-- CPU USAGE -->
                        <v-col cols="12" lg="4" sm="6">
                            <v-row class="text-center">
                                <v-col cols="12">
                                    <v-progress-circular
                                        :model-value="dockerUsage.postgres?.cpu?.avg_percent"
                                        :size="100"
                                        :width="12"
                                        bg-color="grey100"
                                        color="primary"
                                        reveal
                                        rounded
                                    >
                                        <v-avatar color="surface-light" size="70">
                                            {{ dockerUsage.postgres?.cpu?.avg_percent }}%
                                        </v-avatar>
                                    </v-progress-circular>
                                </v-col>
                                <v-col cols="12">
                                    <div class="text-subtitle-2 text-capitalize">{{ t('CPU_USAGE') }}</div>
                                    <div class="text-subtitle-2" dir="LTR">
                                        <span class="me-1">{{ dockerUsage.postgres?.cpu?.used_units }}</span>
                                        /
                                        <span class="ms-1">{{ dockerUsage.postgres?.cpu?.total }} </span>
                                        {{ t('UNITS') }}
                                    </div>
                                </v-col>
                            </v-row>
                        </v-col>

                        <!-- RAM USAGE -->
                        <v-col cols="12" lg="4" sm="6">
                            <v-row class="text-center">
                                <v-col cols="12">
                                    <v-progress-circular
                                        :model-value="dockerUsage.postgres?.ram?.used_percent"
                                        :size="100"
                                        :width="12"
                                        bg-color="grey100"
                                        color="primary"
                                        reveal
                                        rounded
                                    >
                                        <v-avatar color="surface-light" size="70">
                                            {{ dockerUsage.postgres?.ram?.used_percent }}%
                                        </v-avatar>
                                    </v-progress-circular>
                                </v-col>
                                <v-col cols="12">
                                    <div class="text-subtitle-2 text-capitalize">{{ t('RAM_USAGE') }}</div>
                                    <div class="text-subtitle-2" dir="LTR">
                                        <span class="me-1">{{ dockerUsage.postgres?.ram?.used }}</span>
                                        /
                                        <span class="ms-1">{{ dockerUsage.postgres?.ram?.total }} </span>
                                        GB
                                    </div>
                                </v-col>
                            </v-row>
                        </v-col>
                    </v-row>
                </v-col>

                <!-- WEB CONTAINER -->
                <v-col cols="12" lg="4">
                    <v-card-subtitle class="text-h5 text-center"> Web Server (Nginx) </v-card-subtitle>
                    <v-row align="center" justify="center">
                        <!-- CPU USAGE -->
                        <v-col cols="12" lg="4" sm="6">
                            <v-row class="text-center">
                                <v-col cols="12">
                                    <v-progress-circular
                                        :model-value="dockerUsage.web?.cpu?.avg_percent"
                                        :size="100"
                                        :width="12"
                                        bg-color="grey100"
                                        color="primary"
                                        reveal
                                        rounded
                                    >
                                        <v-avatar color="surface-light" size="70">
                                            {{ dockerUsage.web?.cpu?.avg_percent }}%
                                        </v-avatar>
                                    </v-progress-circular>
                                </v-col>
                                <v-col cols="12">
                                    <div class="text-subtitle-2 text-capitalize">{{ t('CPU_USAGE') }}</div>
                                    <div class="text-subtitle-2" dir="LTR">
                                        <span class="me-1">{{ dockerUsage.web?.cpu?.used_units }}</span>
                                        /
                                        <span class="ms-1">{{ dockerUsage.web?.cpu?.total }} </span>
                                        {{ t('UNITS') }}
                                    </div>
                                </v-col>
                            </v-row>
                        </v-col>

                        <!-- RAM USAGE -->
                        <v-col cols="12" lg="4" sm="6">
                            <v-row class="text-center">
                                <v-col cols="12">
                                    <v-progress-circular
                                        :model-value="dockerUsage.web?.ram?.used_percent"
                                        :size="100"
                                        :width="12"
                                        bg-color="grey100"
                                        color="primary"
                                        reveal
                                        rounded
                                    >
                                        <v-avatar color="surface-light" size="70">
                                            {{ dockerUsage.web?.ram?.used_percent }}%
                                        </v-avatar>
                                    </v-progress-circular>
                                </v-col>
                                <v-col cols="12">
                                    <div class="text-subtitle-2 text-capitalize">{{ t('RAM_USAGE') }}</div>
                                    <div class="text-subtitle-2" dir="LTR">
                                        <span class="me-1">{{ dockerUsage.web?.ram?.used }}</span>
                                        /
                                        <span class="ms-1">{{ dockerUsage.web?.ram?.total }} </span>
                                        GB
                                    </div>
                                </v-col>
                            </v-row>
                        </v-col>
                    </v-row>
                </v-col>
            </v-row>
        </v-card-item>
    </v-card>
</template>

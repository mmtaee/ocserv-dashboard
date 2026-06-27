<script lang="ts" setup>
import {
    HomeApi,
    type HomeGetHomeUser,
    type HomeTelegramServiceStatus,
    type ModelsDailyTraffic,
    type ModelsIPBanPoints,
    type RepositoryTopBandwidthUsers,
    type RepositoryTotalBandwidths
} from '@/api';
import { onMounted, ref } from 'vue';
import { getAuthorization } from '@/utils/request';

import OcservStatsOverview from '@/components/dashboard/OcservStatsOverview.vue';
import OnlineUsersOverview from '@/components/dashboard/OnlineUsersOverview.vue';
import IpBansPointOverview from '@/components/dashboard/IpBansPointOverview.vue';
import TopBandwidthUsers from '@/components/dashboard/TopBandwidthUsers.vue';
import UsersOverview from '@/components/dashboard/UsersOverview.vue';
import RxTxDonutOverview from '@/components/dashboard/RxTxDonutOverview.vue';
import RxTxChartOverview from '@/components/dashboard/RxTxChartOverview.vue';
import UiParentCard from '@/components/shared/UiParentCard.vue';
import SystemStats from '@/components/dashboard/SystemStats.vue';
import TelegramStatusOverview from '@/components/dashboard/TelegramStatusOverview.vue';
import { useConfigStore } from '@/stores/config';

const configStore = useConfigStore();

const trafficData = ref<ModelsDailyTraffic[]>([]);
const users = ref<HomeGetHomeUser>({});
const ipBanPoints = ref<ModelsIPBanPoints[]>([]);
const topUsers = ref<RepositoryTopBandwidthUsers>({});
const totalBandwidths = ref<RepositoryTotalBandwidths>({ rx: 0, tx: 0 });
const telegramService = ref<HomeTelegramServiceStatus>({});

onMounted(() => {
    const api = new HomeApi();
    api.homeGet(getAuthorization()).then((res) => {
        users.value = res.data?.users || {};
        trafficData.value = res.data?.statistics || [];
        ipBanPoints.value = res.data?.ip_bans || [];
        topUsers.value = res.data?.top_bandwidth_user || {};
        totalBandwidths.value = res.data?.total_bandwidth || { rx: 0, tx: 0 };
        telegramService.value = res.data?.telegram_service || {};
    });
});
</script>

<template>
    <v-row>
        <v-col cols="12">
            <UiParentCard>
                <v-row>
                    <v-col cols="12" lg="12" v-if="configStore.telegramBotEnabled">
                        <TelegramStatusOverview
                            :enabled="telegramService.enabled"
                            :has-bot-token="telegramService.has_bot_token"
                            :bot-username="telegramService.bot_username"
                        />
                    </v-col>

                    <!-- System Stats Usage overview -->
                    <v-col cols="12" lg="12">
                        <SystemStats />
                    </v-col>

                    <!-- OcservStatsOverview Overview -->
                    <v-col cols="12" lg="12" sm="6">
                        <OcservStatsOverview />
                    </v-col>

                    <!-- Rx Tx overview -->
                    <v-col cols="12" lg="8">
                        <RxTxChartOverview :data="trafficData" />
                    </v-col>

                    <!-- User Overview / Rx Tx overview -->
                    <v-col cols="12" lg="4">
                        <div class="mb-6">
                            <UsersOverview :users="users" />
                        </div>
                        <div>
                            <RxTxDonutOverview :totalBandwidths="totalBandwidths" />
                        </div>
                    </v-col>

                    <!-- Online Users OverView -->
                    <v-col cols="12" lg="8">
                        <OnlineUsersOverview :sessions="users?.online_users_session || []" />
                    </v-col>

                    <!-- IP Ban Points OverView -->
                    <v-col cols="12" lg="4">
                        <IpBansPointOverview :ipBanPoints="ipBanPoints" />
                    </v-col>

                    <!-- Top Bandwidth Users OverView -->
                    <v-col cols="12">
                        <TopBandwidthUsers :topUsers="topUsers" />
                    </v-col>
                </v-row>
            </UiParentCard>
        </v-col>
    </v-row>
</template>

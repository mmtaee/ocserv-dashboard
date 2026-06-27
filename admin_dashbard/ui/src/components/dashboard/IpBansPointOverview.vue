<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import type { ModelsIPBanPoints } from '@/api';

defineProps<{
    ipBanPoints: ModelsIPBanPoints[];
}>();
const { t } = useI18n();
</script>
<template>
    <v-card elevation="10" height="450px">
        <v-card-item class="pb-0">
            <v-card-title class="text-h5 pt-sm-2 text-capitalize">{{ t('LAST_IP_BAN_POINTS_OVERVIEW') }}</v-card-title>
            <div v-if="ipBanPoints.length > 0" class="recent-transaction mt-10 px-3">
                <div v-for="(list, index) in ipBanPoints.slice(0, 5)" :key="`${index}-ip-ban-points-overview`">
                    <v-row class="d-flex mb-4">
                        <v-col class="px-0 pt-0 pb-1 d-flex align-start" cols="4" lg="3" md="auto" sm="auto">
                            <h6 class="text-body-1 textSecondary text-no-wrap">
                                {{ list._Since }}
                            </h6>
                        </v-col>

                        <v-col class="px-0 text-center pt-0 pb-1" cols="1" sm="1">
                            <v-icon color="primary" size="13"> mdi-circle-outline </v-icon>
                            <div v-if="index != 4" class="line mx-auto bg-primary"></div>
                        </v-col>

                        <v-col class="pt-0" cols="7" sm="8">
                            <h6 class="text-body-2 font-weight-bold">
                                {{ list.IP }}
                            </h6>
                            <h6 class="text-body-2 text-primary my-1">{{ t('SCORE') }}: {{ list.Score }}</h6>
                            <div class="text-subtitle-2">
                                {{ list.Since }}
                            </div>
                        </v-col>
                    </v-row>
                </div>
            </div>
            <div v-else class="mt-8 text-capitalize">{{ t('NO_IP_BAN_POINTS_OVERVIEW_FOUND') }}!</div>
        </v-card-item>
    </v-card>
</template>
<style lang="scss">
.recent-transaction {
    .line {
        width: 2px;
        height: 35px;
    }
}
</style>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import UiChildCard from '@/components/shared/UiChildCard.vue';
import DateRange from '@/components/shared/DateRange.vue';
import BarChart from '@/components/shared/BarChart.vue';
import { reactive, ref } from 'vue';
import {
    type ModelsOcservUserSessionLog,
    ModelsOcservUserTrafficTypeEnum,
    OcservUsersApi,
    type OcservUserStatisticsResponse,
    ReportApi
} from '@/api';
import { getAuthorization } from '@/utils/request';
import type { Meta } from '@/types/metaTypes/MetaType';
import { bytesToGB, formatDate, formatDateTime, trafficTypesTransformer } from '@/utils/convertors';
import Pagination from '@/components/shared/Pagination.vue';
import UiParentCard from '@/components/shared/UiParentCard.vue';

const emits = defineEmits(['close']);

const { t } = useI18n();

const loading = ref(false);
const logs = ref<ModelsOcservUserSessionLog[]>([]);

const meta = reactive<Meta>({
    page: 1,
    size: 10,
    sort: 'ASC',
    total_records: 0
});

const today = new Date();
const last30Days = new Date();
last30Days.setDate(today.getDate() - 31);

const initData = {
    dateStart: last30Days,
    dateEnd: today
};

const search = (dateStart: string, dateEnd: string) => {
    const api = new ReportApi();
    api.reportsSessionLogsGet({
        ...getAuthorization(),
        dateStart,
        dateEnd,
        ...meta
    })
        .then((res) => {
            logs.value = res.data.result ?? [];
            Object.assign(meta, res.data.meta);
        })
        .finally(() => {
            loading.value = false;
        });
};

const updateMeta = (newMeta: Meta) => {
    Object.assign(meta, newMeta);
    search(formatDate(initData.dateStart ?? ''), formatDate(initData.dateEnd ?? ''));
};
</script>

<template>
    <v-row>
        <v-col cols="12" md="12">
            <UiParentCard :title="t('OCSERV_USERS_SESSION_LOGS')">
                <v-progress-linear :active="loading" indeterminate></v-progress-linear>

                <div v-if="!loading">
                    <div class="mb-3">
                        <v-row align="center" class="px-md-15 mb-3 text-capitalize" justify="start">
                            <v-col cols="12" md="12" sm="12">
                                <DateRange
                                    :initDate="initData"
                                    :loading="loading"
                                    @search="search"
                                    :disable-more30-days="false"
                                    :pre-hook="false"
                                />
                            </v-col>
                        </v-row>
                    </div>

                    <v-table v-if="logs.length > 0" class="px-md-15">
                        <thead>
                            <tr class="text-capitalize bg-lightprimary">
                                <th class="text-left">{{ t('USERNAME') }}</th>
                                <th class="text-left">{{ t('CREATED_AT') }}</th>
                                <th class="text-left">IP</th>
                                <th class="text-left">{{ t('EVENT') }}</th>
                                <th class="text-left">{{ t('MESSAGE') }}</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="(item, index) in logs" :key="index">
                                <td>{{ item.username }}</td>
                                <td>{{ formatDateTime(item.created_at, '') }}</td>
                                <td>{{ item.ip }}</td>
                                <td>{{ item.event }}</td>
                                <td>{{ item.message }}</td>
                            </tr>
                        </tbody>
                    </v-table>
                </div>

                <div v-if="loading || logs.length == 0" class="ms-md-5 mb-md-5 text-capitalize">
                    {{ t('NO_OCSERV_USER_SESSION_LOG_FOUND') }}
                </div>

                <Pagination :totalRecords="meta.total_records" @update="updateMeta" />
            </UiParentCard>
        </v-col>
    </v-row>
</template>

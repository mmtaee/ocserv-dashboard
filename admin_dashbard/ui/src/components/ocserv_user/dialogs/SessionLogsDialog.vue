<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import DateRange from '@/components/shared/DateRange.vue';
import { reactive, ref } from 'vue';
import { type ModelsOcservUserSessionLog, OcservUsersApi } from '@/api';
import { getAuthorization } from '@/utils/request';
import type { Meta } from '@/types/metaTypes/MetaType';
import { formatDate, formatDateTime } from '@/utils/convertors';
import Pagination from '@/components/shared/Pagination.vue';

const props = defineProps({
    show: {
        type: Boolean,
        default: false
    },
    username: {
        type: String,
        required: true
    },
    uid: {
        type: String,
        required: true
    }
});

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
    const api = new OcservUsersApi();
    api.ocservUsersUidSessionLogsGet({
        ...getAuthorization(),
        dateStart,
        dateEnd,
        ...meta,
        uid: props.uid
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
    <v-dialog v-model="props.show" height="980" max-width="1200">
        <v-card>
            <v-card-title class="bg-info text-capitalize">
                <v-row align="end" class="no-gutters" justify="space-between">
                    <v-col md="auto"> {{ t('SESSION_LOGS') }} ({{ username }}) </v-col>
                    <v-col md="auto">
                        <v-icon @click="emits('close')">mdi-close</v-icon>
                    </v-col>
                </v-row>
            </v-card-title>

            <v-card-text>
                <v-row>
                    <v-col cols="12" md="12">
                        <v-progress-linear :active="loading" indeterminate></v-progress-linear>

                        <div v-if="!loading">
                            <div class="mb-3">
                                <v-row align="center" class="px-md-15 mb-3 text-capitalize" justify="start">
                                    <v-col cols="12" md="12" sm="12">
                                        <DateRange
                                            :disable-more30-days="false"
                                            :initDate="initData"
                                            :loading="loading"
                                            :pre-hook="false"
                                            @search="search"
                                        />
                                    </v-col>
                                </v-row>
                            </div>

                            <v-table v-if="logs.length > 0" class="px-md-15">
                                <thead>
                                    <tr class="text-capitalize bg-lightprimary">
                                        <th class="text-left">{{ t('CREATED_AT') }}</th>
                                        <th class="text-left">IP</th>
                                        <th class="text-left">{{ t('EVENT') }}</th>
                                        <th class="text-left">{{ t('MESSAGE') }}</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <tr v-for="(item, index) in logs" :key="index">
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
                    </v-col>
                </v-row>
            </v-card-text>
        </v-card>
    </v-dialog>
</template>

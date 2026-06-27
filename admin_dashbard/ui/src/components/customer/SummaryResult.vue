<script lang="ts" setup>
import type { CustomerSummaryResponse } from '@/api';
import { bytesToGB, bytesToTrafficSize, formatDate, numberToFixer, trafficTypesTransformer } from '@/utils/convertors';
import UiChildCard from '@/components/shared/UiChildCard.vue';
import { useI18n } from 'vue-i18n';

defineProps<{
    result: CustomerSummaryResponse;
}>();

const emit = defineEmits(['newSummary', 'disconnect', 'downloadCertificate', 'openCiscoSetup']);

const { t } = useI18n();
</script>

<template>
    <v-col cols="12" sm="10" md="8" lg="6">
        <UiChildCard class="px-2 px-sm-3">
            <template #title-header>
                <v-row align="center" justify="start" class="ga-2">
                    <v-col cols="12" md="8">
                        <span class="text-capitalize text-primary text-h4 text-sm-h3">
                            <span>{{ result.ocserv_user.username }}</span>
                            <span class="text-muted mx-1 text-capitalize"> ({{ t('ACCOUNT_AND_USAGE_SUMMARY') }}) </span>
                        </span>
                    </v-col>

                    <v-col cols="12" sm="auto">
                        <v-btn block color="primary" flat size="small" @click="emit('disconnect')">
                            {{ t('DISCONNECT_ALL_SESSIONS') }}
                        </v-btn>
                    </v-col>

                    <v-col v-if="result.ocserv_user.certificate_available" cols="12" sm="auto">
                        <v-btn block color="info" size="small" variant="outlined" @click="emit('downloadCertificate')">
                            {{ t('DOWNLOAD_CERTIFICATE') }}
                        </v-btn>
                    </v-col>

                    <v-col cols="12" sm="auto">
                        <v-btn block color="primary" size="small" variant="outlined" @click="emit('openCiscoSetup')">
                            {{ t('CISCO_SETUP') }}
                        </v-btn>
                    </v-col>
                </v-row>
            </template>

            <template #action>
                <v-spacer />
                <v-btn color="primary" @click="emit('newSummary')">
                    {{ t('CHECK_ANOTHER') }}
                </v-btn>
            </template>

            <div class="space-y-4 mt-8 px-0 px-sm-1">
                <!-- General info -->
                <div class="bg-surface shadow rounded-lg pa-3 pa-sm-4">
                    <h4 class="text-lg font-semibold my-4">{{ t('DETAILS') }}</h4>

                    <div class="mx-1 mx-sm-5">
                        <v-row align="center" justify="start">
                            <v-col cols="12" md="6">
                                <span class="font-medium text-gray-600 text-capitalize"> {{ t('TRAFFIC_TYPE') }}: </span>
                                <span class="ms-1 text-capitalize">
                                    {{ trafficTypesTransformer(result.ocserv_user.traffic_type) }}
                                </span>
                            </v-col>

                            <v-col cols="12" md="6">
                                <span class="font-medium text-gray-600 text-capitalize"> {{ t('TRAFFIC_SIZE') }}: </span>
                                <span class="ms-1">{{ bytesToTrafficSize(result.ocserv_user.traffic_size) }}</span>
                            </v-col>
                        </v-row>

                        <v-row align="center" justify="start">
                            <v-col cols="12" md="6">
                                <span class="font-medium text-gray-600 text-capitalize"> {{ t('CERTIFICATE') }}: </span>
                                <span
                                    :class="result.ocserv_user.certificate_enabled ? 'text-success' : 'text-warning'"
                                    class="ms-1 text-capitalize"
                                >
                                    {{ result.ocserv_user.certificate_enabled ? t('ENABLED') : t('DISABLED') }}
                                </span>
                            </v-col>

                            <v-col cols="12" md="6">
                                <span class="font-medium text-gray-600 text-capitalize"> {{ t('TOTAL') }}: </span>
                                <span class="ms-1">
                                    {{ bytesToGB(result.ocserv_user.rx + result.ocserv_user.tx) }} GB
                                </span>
                            </v-col>
                        </v-row>

                        <v-row align="center" justify="start">
                            <v-col cols="12" md="6">
                                <span class="font-medium text-gray-600 text-capitalize"> RX ({{ t('TOTAL') }}): </span>
                                <span class="ms-1">{{ bytesToGB(result.ocserv_user.rx) }} GB</span>
                            </v-col>

                            <v-col cols="12" md="6">
                                <span class="font-medium text-gray-600 text-capitalize"> TX ({{ t('TOTAL') }}): </span>
                                <span class="ms-1">{{ bytesToGB(result.ocserv_user.tx) }} GB</span>
                            </v-col>
                        </v-row>

                        <v-row align="center" justify="start">
                            <v-col cols="12" md="6">
                                <span class="font-medium text-gray-600 text-capitalize"> {{ t('EXPIRE_AT') }}: </span>
                                <span v-if="result.ocserv_user.expire_at" class="ms-1">
                                    {{ formatDate(result.ocserv_user.expire_at) }}
                                </span>
                                <span v-else class="ms-1 text-warning italic">{{ t('NOT_SET') }}</span>
                            </v-col>

                            <v-col cols="12" md="6">
                                <span class="font-medium text-gray-600 text-capitalize">
                                    {{ t('DEACTIVATED_AT') }}:
                                </span>
                                <span v-if="result.ocserv_user.deactivated_at" class="ms-1">
                                    {{ formatDate(result.ocserv_user.deactivated_at) }}
                                </span>
                                <span v-else class="ms-1 text-warning italic">{{ t('NOT_SET') }}</span>
                            </v-col>
                        </v-row>
                    </div>

                    <h4 class="text-lg font-semibold my-4">{{ t('MONTHLY_BANDWIDTHS') }}</h4>

                    <div class="mx-1 mx-sm-5">
                        <v-row align="center" justify="start">
                            <v-col cols="12" md="6">
                                <span class="font-medium text-gray-600 text-capitalize"> {{ t('DATE_START') }}: </span>
                                <span class="ms-1">{{ formatDate(result.usage.date_start) }}</span>
                            </v-col>

                            <v-col cols="12" md="6">
                                <span class="font-medium text-gray-600 text-capitalize"> {{ t('DATE_END') }}: </span>
                                <span class="ms-1">{{ formatDate(result.usage.date_end) }}</span>
                            </v-col>
                        </v-row>
                        <v-row align="center" justify="start">
                            <v-col cols="12" md="6">
                                <span class="font-medium text-gray-600 text-capitalize"> RX: </span>
                                <span class="ms-1">{{ numberToFixer(result.usage.bandwidths.rx, 6) }} GB</span>
                            </v-col>

                            <v-col cols="12" md="6">
                                <span class="font-medium text-gray-600 text-capitalize"> TX: </span>
                                <span class="ms-1">{{ numberToFixer(result.usage.bandwidths.tx, 6) }} GB</span>
                            </v-col>
                        </v-row>
                    </div>
                </div>
            </div>
        </UiChildCard>
    </v-col>
</template>

<script setup lang="ts">
import { type PropType, ref } from 'vue';
import type { ModelsOcservGroupConfig } from '@/api';
import { useI18n } from 'vue-i18n';

defineProps({
    resultArrayObj: {
        type: Object as PropType<ModelsOcservGroupConfig>,
        required: true
    },
    resultOther: {
        type: Object as PropType<ModelsOcservGroupConfig>,
        required: true
    }
});

const { t } = useI18n();
</script>

<template>
    <div class="bg-white shadow rounded-lg p-4">
        <h2 class="text-lg font-semibold my-4 text-capitalize">{{ t('CONFIGURATION') }}</h2>

        <v-row class="mx-3">
            <v-col class="text-h6 text-capitalize" cols="12">
                {{ t('NETWORK_CONFIGURATION') }}
            </v-col>
            <v-col v-for="(val, key, index) in resultOther" :key="`config-${index}`" class="pa-3" cols="12" md="4">
                <span v-if="!Array.isArray(val)">
                    <span class="w-40 font-medium text-gray-600">{{ key }}: </span>
                    <span v-if="val" class="text-primary">
                        {{ val }}
                    </span>
                    <span v-else class="text-warning italic">{{ t('NOT_SET') }}</span>
                </span>
            </v-col>
        </v-row>

        <v-row class="mx-3">
            <v-col class="text-h6 text-capitalize" cols="12">
                {{ t('ROUTES') }}
            </v-col>
            <v-col
                v-for="(val, key, index) in resultArrayObj"
                :key="`config-array-obj-${index}`"
                class="pa-3"
                cols="12"
                md="3"
            >
                <v-card class="overflow-y-auto" elevation="1" height="200" variant="text">
                    <v-card-title class="text-subtitle-1 pa-2"> {{ key }}:</v-card-title>
                    <v-card-text>
                        <span v-for="(v, index) in val" v-if="Array.isArray(val)" :key="index" class="mx-1 text-primary">
                            {{ v }} <br />
                        </span>
                    </v-card-text>
                </v-card>
            </v-col>
        </v-row>
    </div>
</template>

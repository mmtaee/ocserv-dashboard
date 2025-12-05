<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { computed, onMounted, ref } from 'vue';
import {
    ModelsOcservUserTrafficTypeEnum,
    type OcservUserSyncOcpasswdRequest,
    OcservUserSyncOcpasswdRequestTrafficTypeEnum
} from '@/api';
import { formatDate } from '@/utils/convertors';
import { requiredRule } from '@/utils/rules';

const props = defineProps({
    show: {
        type: Boolean,
        default: false
    },
    loading: {
        type: Boolean,
        default: false
    }
});

const emits = defineEmits(['saveToDB', 'close']);

const { t } = useI18n();
const valid = ref(true);
const showDateMenu = ref(false);
const config = ref<OcservUserSyncOcpasswdRequest>({
    config: {},
    description: '',
    expire_at: '',
    traffic_size: 0,
    traffic_type: OcservUserSyncOcpasswdRequestTrafficTypeEnum.FREE,
    users: []
});

const trafficTypes = ref([
    {
        label: t('FREE'),
        value: ModelsOcservUserTrafficTypeEnum.FREE
    },
    {
        label: t('MONTHLY_TRANSMIT'),
        value: ModelsOcservUserTrafficTypeEnum.MONTHLY_TRANSMIT
    },
    {
        label: t('MONTHLY_RECEIVE'),
        value: ModelsOcservUserTrafficTypeEnum.MONTHLY_RECEIVE
    },
    {
        label: t('TOTALLY_RECEIVE'),
        value: ModelsOcservUserTrafficTypeEnum.TOTALLY_RECEIVE
    },
    {
        label: t('TOTALLY_TRANSMIT'),
        value: ModelsOcservUserTrafficTypeEnum.TOTALLY_TRANSMIT
    }
]);

const rules = {
    required: (v: string) => requiredRule(v, t)
};

const expireAtDate = computed<Date>({
    get: () => {
        return config.value.expire_at ? new Date(config.value.expire_at) : new Date();
    },
    set: (val: Date) => {
        config.value.expire_at = formatDate(val);
    }
});

const save = () => {
    emits('saveToDB', config.value);
    config.value = {
        config: {},
        description: '',
        expire_at: '',
        traffic_size: 0,
        traffic_type: OcservUserSyncOcpasswdRequestTrafficTypeEnum.FREE,
        users: []
    };
};
</script>

<template>
    <template>
        <v-dialog v-model="props.show" max-width="450">
            <v-card>
                <v-card-title class="bg-primary text-capitalize">
                    {{ t('SAVE_DB_DIALOG_TITLE') }}
                </v-card-title>

                <v-card-text>
                    <span class="text-capitalize"> {{ t('SAVE_DB_DIALOG_TEXT_HELP') }} </span>

                    <v-form v-model="valid" class="mt-5">
                        <v-row align="center" justify="start">
                            <v-col cols="12">
                                <v-label class="font-weight-bold mb-1 text-capitalize">{{ t('TRAFFIC_TYPE') }}</v-label>
                                <v-select
                                    v-model="config.traffic_type"
                                    :items="trafficTypes"
                                    :rules="[rules.required]"
                                    color="primary"
                                    hide-details
                                    item-title="label"
                                    item-value="value"
                                    variant="outlined"
                                    @update:modelValue="
                                        (v) =>
                                            v == ModelsOcservUserTrafficTypeEnum.FREE ? (config.traffic_size = 0) : false
                                    "
                                />
                            </v-col>
                            <v-col cols="12">
                                <v-label class="font-weight-bold mb-1 text-capitalize">{{ t('TRAFFIC_SIZE') }}</v-label>
                                <v-text-field
                                    v-model.number="config.traffic_size"
                                    :disabled="config.traffic_type == ModelsOcservUserTrafficTypeEnum.FREE"
                                    :rules="
                                        config.traffic_type == ModelsOcservUserTrafficTypeEnum.FREE
                                            ? []
                                            : [rules.required]
                                    "
                                    color="primary"
                                    hide-details
                                    suffix="GB"
                                    type="number"
                                    variant="outlined"
                                    @update:modelValue="
                                        (val: any) => {
                                            config.traffic_size = Boolean(val) ? (Number(val) as any) : null;
                                        }
                                    "
                                />
                            </v-col>
                            <v-col cols="12">
                                <v-menu
                                    v-model="showDateMenu"
                                    :close-on-content-click="false"
                                    transition="scale-transition"
                                >
                                    <template #activator="{ props }">
                                        <v-label class="font-weight-bold mb-1 text-capitalize">
                                            {{ t('EXPIRE_AT') }}
                                        </v-label>
                                        <v-text-field
                                            :model-value="config.expire_at ? formatDate(config.expire_at) : ''"
                                            color="primary"
                                            hide-details
                                            readonly
                                            v-bind="props"
                                            variant="outlined"
                                            clearable
                                            @click:clear="config.expire_at = ''"
                                        />
                                    </template>
                                    <v-date-picker
                                        v-model="expireAtDate"
                                        :header="t('EXPIRE_AT')"
                                        elevation="24"
                                        title=""
                                        @update:model-value="() => (showDateMenu = false)"
                                    />
                                </v-menu>
                            </v-col>
                        </v-row>
                    </v-form>
                </v-card-text>

                <v-card-actions class="mx-2 my-1">
                    <v-spacer></v-spacer>

                    <v-btn color="grey" variant="tonal" @click="emits('close')">
                        {{ t('CANCEL') }}
                    </v-btn>
                    <v-btn color="primary" variant="flat" @click="save" :loading="loading">
                        <v-icon color="white" start class="ms-1">mdi-database-arrow-left-outline</v-icon> {{ t('SYNC') }}
                    </v-btn>
                </v-card-actions>
            </v-card>
        </v-dialog>
    </template>
</template>

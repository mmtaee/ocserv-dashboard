<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { onMounted, ref } from 'vue';
import {
    type ModelsOcservUser,
    type ModelsOcservUserConfig,
    ModelsOcservUserTrafficTypeEnum,
    OcservUsersApi
} from '@/api';
import { getAuthorization } from '@/utils/request';
import { router } from '@/router';
import UiParentCard from '@/components/shared/UiParentCard.vue';
import UiChildCard from '@/components/shared/UiChildCard.vue';
import {
    formatDate,
    formatDateTimeWithRelative,
    formatDateWithRelative,
    trafficTypesTransformer
} from '@/utils/convertors';

const props = defineProps<{ uid: string }>();

const { t } = useI18n();
const result = ref<ModelsOcservUser>({
    created_at: '',
    group: '',
    is_locked: false,
    is_online: false,
    owner: '',
    password: '',
    rx: 0,
    traffic_size: 0,
    traffic_type: ModelsOcservUserTrafficTypeEnum.FREE,
    tx: 0,
    uid: '',
    username: ''
});

const configArrayKeys = ['route', 'no-route', 'dns', 'split-dns'];
const resultArrayObj = ref<ModelsOcservUserConfig>({});
const resultOther = ref<ModelsOcservUserConfig>({});
const showPassword = ref(false);

const getUser = () => {
    if (props.uid == undefined) {
        return;
    }
    const api = new OcservUsersApi();
    api.ocservUsersUidGet({
        ...getAuthorization(),
        uid: props.uid
    }).then((res) => {
        console.log(res.data);
        result.value = res.data;
        resultArrayObj.value = Object.entries(res.data?.config || {})
            .filter(([key]) => configArrayKeys.includes(key))
            .reduce<ModelsOcservUserConfig>((acc, [key, val]) => {
                (acc as any)[key] = val;
                return acc;
            }, {} as ModelsOcservUserConfig);

        // keep keys NOT in configArrayKeys
        resultOther.value = Object.entries(res.data?.config || {})
            .filter(([key]) => !configArrayKeys.includes(key))
            .reduce<ModelsOcservUserConfig>((acc, [key, val]) => {
                (acc as any)[key] = val;
                return acc;
            }, {} as ModelsOcservUserConfig);
    });
};

onMounted(() => {
    getUser();
});
</script>

<template>
    <v-row>
        <v-col cols="12" md="12">
            <UiParentCard :title="t('OCSERV_USER_DETAIL_TITLE')">
                <template #header-prepend>
                    <v-tooltip :text="t('GO_BACK_TO_USERS')">
                        <template #activator="{ props }">
                            <v-icon start v-bind="props" @click.stop="router.push({ name: 'Ocserv Users' })">
                                mdi-arrow-left-top
                            </v-icon>
                        </template>
                    </v-tooltip>
                </template>
                <template #action>
                    <v-btn
                        class="me-lg-5"
                        color="grey"
                        size="small"
                        variant="flat"
                        @click="router.push({ name: 'Ocserv User Update', params: { uid: props.uid } })"
                    >
                        {{ t('UPDATE') }}
                    </v-btn>
                </template>
                <UiChildCard class="px-3">
                    <div class="space-y-4">
                        <!-- General info -->
                        <div class="bg-white shadow rounded-lg p-4">
                            <h2 class="text-lg font-semibold mb-3 text-capitalize">{{ t('DETAILS') }}</h2>

                            <div class="grid grid-cols-2 gap-4 mx-5">
                                <v-row align="center" justify="start">
                                    <v-col cols="12" md="4">
                                        <span class="font-medium text-gray-600">UID:</span>
                                        <span class="ms-1 text-primary">{{ result.uid }}</span>
                                    </v-col>

                                    <v-col cols="12" md="4">
                                        <span class="font-medium text-gray-600 text-capitalize"> {{ t('OWNER') }}: </span>
                                        <span class="ms-1 text-capitalize text-primary">{{ result.owner }}</span>
                                    </v-col>

                                    <v-col cols="12" md="4">
                                        <span class="font-medium text-gray-600 text-capitalize">
                                            {{ t('USERNAME') }}:
                                        </span>
                                        <span class="ms-1 text-primary">{{ result.username }}</span>
                                    </v-col>

                                    <v-col cols="12" md="4">
                                        <span class="font-medium text-gray-600 text-capitalize">
                                            {{ t('PASSWORD') }}:
                                        </span>
                                        <span v-if="showPassword" class="ms-1 text-primary">{{ result.password }}</span>
                                        <span v-else class="ms-1">
                                            {{ '*'.repeat(result.password?.length || 0) }}
                                            <v-icon class="mx-1" @click="showPassword = true">mdi-eye-outline</v-icon>
                                        </span>
                                    </v-col>

                                    <v-col cols="12" md="4">
                                        <span class="font-medium text-gray-600 text-capitalize">{{ t('GROUP') }}:</span>
                                        <span class="ms-1 text-primary">{{ result.group }}</span>
                                    </v-col>

                                    <v-col cols="12" md="4">
                                        <span class="font-medium text-gray-600 text-capitalize">
                                            {{ t('TRAFFIC_TYPE') }}:
                                        </span>
                                        <span class="ms-1 text-capitalize text-primary">
                                            {{ trafficTypesTransformer(result.traffic_type) }}
                                        </span>
                                    </v-col>

                                    <v-col cols="12" md="4">
                                        <span class="font-medium text-gray-600 text-capitalize">
                                            {{ t('TRAFFIC_SIZE') }}:
                                        </span>
                                        <span class="ms-1 text-primary">
                                            {{
                                                result.traffic_size &&
                                                result.traffic_type !== ModelsOcservUserTrafficTypeEnum.FREE
                                                    ? result.traffic_size + ' GB'
                                                    : t('FREE')
                                            }}
                                        </span>
                                    </v-col>

                                    <v-col cols="12" md="4">
                                        <span class="font-medium text-gray-600 text-capitalize">
                                            RX {{ t('USED') }}:
                                        </span>
                                        <span class="ms-1 text-primary">{{ result.rx }} GB</span>
                                    </v-col>

                                    <v-col cols="12" md="4">
                                        <span class="font-medium text-gray-600 text-capitalize">
                                            TX {{ t('USED') }}:
                                        </span>
                                        <span class="ms-1 text-primary">{{ result.tx }} GB</span>
                                    </v-col>

                                    <v-col cols="12" md="4">
                                        <span class="font-medium text-gray-600 text-capitalize">
                                            {{ t('CREATED_AT') }}:
                                        </span>
                                        <span v-if="result.created_at" class="ms-1 text-primary">
                                            {{ formatDateWithRelative(result.created_at, '') }}
                                        </span>
                                        <span v-else class="ms-1 text-warning italic">{{ t('NOT_SET') }}</span>
                                    </v-col>

                                    <v-col cols="12" md="4">
                                        <span class="font-medium text-gray-600 text-capitalize">
                                            {{ t('EXPIRE_AT') }}:
                                        </span>
                                        <span v-if="result.expire_at" class="ms-1 text-primary">
                                            {{ formatDateWithRelative(result.expire_at, '') }}
                                        </span>
                                        <span v-else class="ms-1 text-warning italic">{{ t('UNLIMITED') }}</span>
                                    </v-col>

                                    <v-col cols="12" md="4">
                                        <span class="font-medium text-gray-600 text-capitalize">
                                            {{ t('DEACTIVATED_AT') }}:
                                        </span>
                                        <span v-if="result.deactivated_at" class="ms-1 text-primary">
                                            {{ formatDateWithRelative(result.deactivated_at, '') }}
                                        </span>
                                        <span v-else class="ms-1 text-info italic text-capitalize">
                                            {{ t('USER_IS_ACTIVE_NOW') }}
                                        </span>
                                    </v-col>

                                    <v-col cols="12" md="4">
                                        <span class="font-medium text-gray-600 text-capitalize">
                                            {{ t('DESCRIPTION') }}:
                                        </span>
                                        <span v-if="result.description" class="ms-1 text-primary">
                                            {{ result.description }}
                                        </span>
                                        <span v-else class="ms-1 text-warning italic">{{ t('NOT_SET') }}</span>
                                    </v-col>
                                </v-row>
                                <div></div>
                            </div>
                        </div>

                        <!-- Config section -->
                        <div class="bg-white shadow rounded-lg p-4">
                            <h2 class="text-lg font-semibold my-4 text-capitalize">{{ t('CONFIGURATION') }}</h2>

                            <v-row class="mx-3">
                                <v-col class="text-h6 text-capitalize" cols="12">
                                    {{ t('NETWORK_CONFIGURATION') }}
                                </v-col>
                                <v-col
                                    v-for="(val, key, index) in resultOther"
                                    :key="`config-${index}`"
                                    class="pa-3"
                                    cols="12"
                                    md="4"
                                >
                                    <span v-if="!Array.isArray(val)">
                                        <span class="w-40 font-medium text-gray-600">{{ key }}: </span>
                                        <span v-if="val" class="text-primary">{{ val }}</span>
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
                                            <span
                                                v-for="(v, index) in val"
                                                v-if="Array.isArray(val)"
                                                :key="index"
                                                class="mx-1 text-primary"
                                            >
                                                {{ v }} <br />
                                            </span>
                                        </v-card-text>
                                    </v-card>
                                </v-col>
                            </v-row>
                        </div>
                    </div>
                </UiChildCard>
            </UiParentCard>
        </v-col>
    </v-row>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { ModelsOcservUserTrafficTypeEnum } from '@/api';
import { TelegramAPI, type TelegramPackage } from '@/api/telegram';
import { useSnackbarStore } from '@/stores/snackbar';
import UiParentCard from '@/components/shared/UiParentCard.vue';
import { trafficTypesTransformer } from '@/utils/convertors';

const { t } = useI18n();
const snackbar = useSnackbarStore();

const loading = ref(false);
const items = ref<TelegramPackage[]>([]);
const dialog = ref(false);
const editing = ref<TelegramPackage>(emptyPackage());
const isNew = ref(true);

const trafficTypes = [
    {
        value: ModelsOcservUserTrafficTypeEnum.TOTALLY_TRANSMIT,
        title: trafficTypesTransformer(ModelsOcservUserTrafficTypeEnum.TOTALLY_TRANSMIT)
    },
    {
        value: ModelsOcservUserTrafficTypeEnum.TOTALLY_RECEIVE,
        title: trafficTypesTransformer(ModelsOcservUserTrafficTypeEnum.TOTALLY_RECEIVE)
    },
    {
        value: ModelsOcservUserTrafficTypeEnum.TOTALLY_RX_TX,
        title: trafficTypesTransformer(ModelsOcservUserTrafficTypeEnum.TOTALLY_RX_TX)
    },
    {
        value: ModelsOcservUserTrafficTypeEnum.MONTHLY_TRANSMIT,
        title: trafficTypesTransformer(ModelsOcservUserTrafficTypeEnum.MONTHLY_TRANSMIT)
    },
    {
        value: ModelsOcservUserTrafficTypeEnum.MONTHLY_RECEIVE,
        title: trafficTypesTransformer(ModelsOcservUserTrafficTypeEnum.MONTHLY_RECEIVE)
    },
    {
        value: ModelsOcservUserTrafficTypeEnum.MONTHLY_RX_TX,
        title: trafficTypesTransformer(ModelsOcservUserTrafficTypeEnum.MONTHLY_RX_TX)
    },
    {
        value: ModelsOcservUserTrafficTypeEnum.FREE,
        title: trafficTypesTransformer(ModelsOcservUserTrafficTypeEnum.FREE)
    }
];

function emptyPackage(): TelegramPackage {
    return {
        id: 0,
        title: '',
        days: 30,
        traffic_size_gb: 30,
        traffic_type: ModelsOcservUserTrafficTypeEnum.TOTALLY_TRANSMIT,
        price_text: '',
        is_active: true
    };
}

const formatTrafficType = (trafficType: string): string => {
    return trafficTypesTransformer(trafficType as ModelsOcservUserTrafficTypeEnum);
};

const load = async () => {
    loading.value = true;
    try {
        const res = await TelegramAPI.listPackages(true);
        items.value = res.data;
    } finally {
        loading.value = false;
    }
};

const openCreate = () => {
    editing.value = emptyPackage();
    isNew.value = true;
    dialog.value = true;
};

const openEdit = (pkg: TelegramPackage) => {
    editing.value = { ...pkg };
    isNew.value = false;
    dialog.value = true;
};

const save = async () => {
    loading.value = true;
    try {
        if (isNew.value) {
            await TelegramAPI.createPackage({
                title: editing.value.title,
                days: editing.value.days,
                traffic_size_gb: editing.value.traffic_size_gb,
                traffic_type: editing.value.traffic_type,
                price_text: editing.value.price_text,
                is_active: editing.value.is_active
            });
        } else {
            await TelegramAPI.updatePackage(editing.value.id, {
                title: editing.value.title,
                days: editing.value.days,
                traffic_size_gb: editing.value.traffic_size_gb,
                traffic_type: editing.value.traffic_type,
                price_text: editing.value.price_text,
                is_active: editing.value.is_active
            });
        }
        dialog.value = false;
        await load();
        snackbar.show({
            id: 1,
            message: t('TELEGRAM_PACKAGE_SAVED'),
            color: 'success',
            timeout: 3000
        });
    } finally {
        loading.value = false;
    }
};

const remove = async (pkg: TelegramPackage) => {
    if (!confirm(t('CONFIRM_DELETE'))) return;
    loading.value = true;
    try {
        await TelegramAPI.deletePackage(pkg.id);
        await load();
    } finally {
        loading.value = false;
    }
};

onMounted(load);
</script>

<template>
    <v-row>
        <v-col cols="12">
            <UiParentCard :title="t('TELEGRAM_PACKAGES')">
                <template #action>
                    <v-btn class="me-lg-5" color="grey" size="small" variant="outlined" @click="openCreate">
                        {{ t('CREATE') }}
                    </v-btn>
                </template>

                <v-progress-linear :active="loading" indeterminate />

                <div v-if="!loading">
                    <v-table v-if="items.length > 0" density="comfortable" class="px-md-15">
                        <thead>
                            <tr class="text-capitalize bg-lightprimary">
                                <th class="text-left">{{ t('TITLE') }}</th>
                                <th class="text-left">{{ t('DAYS') }}</th>
                                <th class="text-left">{{ t('TRAFFIC_SIZE_GB') }}</th>
                                <th class="text-left">{{ t('TRAFFIC_TYPE') }}</th>
                                <th class="text-left">{{ t('PRICE') }}</th>
                                <th class="text-left">{{ t('STATUS') }}</th>
                                <th class="text-left">{{ t('ACTION') }}</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="pkg in items" :key="pkg.id">
                                <td>{{ pkg.title }}</td>
                                <td>{{ pkg.days }}</td>
                                <td>{{ pkg.traffic_size_gb }}</td>
                                <td>{{ formatTrafficType(pkg.traffic_type) }}</td>
                                <td>{{ pkg.price_text || '—' }}</td>
                                <td>
                                    <v-chip :color="pkg.is_active ? 'success' : 'grey'" size="small" variant="flat">
                                        {{ pkg.is_active ? t('ACTIVE') : t('INACTIVE') }}
                                    </v-chip>
                                </td>
                                <td>
                                    <v-btn icon="mdi-pencil" size="small" variant="text" @click="openEdit(pkg)" />
                                    <v-btn
                                        icon="mdi-delete"
                                        size="small"
                                        variant="text"
                                        color="error"
                                        @click="remove(pkg)"
                                    />
                                </td>
                            </tr>
                        </tbody>
                    </v-table>
                </div>

                <div v-if="loading || items.length == 0" class="ms-md-5 mb-md-5 text-capitalize">
                    {{ t('NO_DATA') }}
                </div>
            </UiParentCard>
        </v-col>

        <v-dialog v-model="dialog" max-width="600">
            <v-card>
                <v-card-title>
                    {{ isNew ? t('TELEGRAM_PACKAGE_CREATE') : t('TELEGRAM_PACKAGE_UPDATE') }}
                </v-card-title>
                <v-card-text>
                    <v-text-field v-model="editing.title" :label="t('TITLE')" variant="outlined" density="comfortable" />
                    <v-row>
                        <v-col cols="12" md="6">
                            <v-text-field
                                v-model.number="editing.days"
                                type="number"
                                :label="t('DAYS')"
                                variant="outlined"
                                density="comfortable"
                            />
                        </v-col>
                        <v-col cols="12" md="6">
                            <v-text-field
                                v-model.number="editing.traffic_size_gb"
                                type="number"
                                :label="`${t('TRAFFIC_SIZE')} (GB)`"
                                variant="outlined"
                                density="comfortable"
                            />
                        </v-col>
                    </v-row>
                    <v-select
                        v-model="editing.traffic_type"
                        :items="trafficTypes"
                        :label="t('TRAFFIC_TYPE')"
                        variant="outlined"
                        density="comfortable"
                    />
                    <v-text-field
                        v-model="editing.price_text"
                        :label="t('PRICE')"
                        variant="outlined"
                        density="comfortable"
                    />
                    <v-switch v-model="editing.is_active" :label="t('ACTIVE')" color="primary" hide-details />
                </v-card-text>
                <v-card-actions>
                    <v-spacer />
                    <v-btn variant="text" @click="dialog = false">{{ t('CANCEL') }}</v-btn>
                    <v-btn color="primary" :loading="loading" @click="save">
                        {{ t('SAVE') }}
                    </v-btn>
                </v-card-actions>
            </v-card>
        </v-dialog>
    </v-row>
</template>

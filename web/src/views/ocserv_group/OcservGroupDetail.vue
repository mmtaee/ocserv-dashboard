<script lang="ts" setup>
import UiParentCard from '@/components/shared/UiParentCard.vue';
import UiChildCard from '@/components/shared/UiChildCard.vue';
import { useI18n } from 'vue-i18n';
import { onMounted, ref } from 'vue';
import { type ModelsOcservGroup, type ModelsOcservGroupConfig, OcservGroupsApi } from '@/api';
import { getAuthorization } from '@/utils/request';
import { router } from '@/router';
import OcservGroupConfigDetail from '@/components/ocserv_group/OcservGroupConfigDetail.vue';

const props = defineProps<{ id?: number }>();

const { t } = useI18n();
const result = ref<ModelsOcservGroup>({ config: undefined, id: 0, name: '', owner: '' });

const configArrayKeys = ['route', 'no-route', 'dns', 'split-dns'];
const resultArrayObj = ref<ModelsOcservGroupConfig>({});
const resultOther = ref<ModelsOcservGroupConfig>({});

const getGroup = () => {
    if (props.id == undefined) {
        return;
    }
    const api = new OcservGroupsApi();
    api.ocservGroupsIdGet({
        ...getAuthorization(),
        id: props.id
    }).then((res) => {
        result.value = res.data;
        resultArrayObj.value = Object.entries(res.data?.config || {})
            .filter(([key]) => configArrayKeys.includes(key))
            .reduce<ModelsOcservGroupConfig>((acc, [key, val]) => {
                (acc as any)[key] = val;
                return acc;
            }, {} as ModelsOcservGroupConfig);

        // keep keys NOT in configArrayKeys
        resultOther.value = Object.entries(res.data?.config || {})
            .filter(([key]) => !configArrayKeys.includes(key))
            .reduce<ModelsOcservGroupConfig>((acc, [key, val]) => {
                (acc as any)[key] = val;
                return acc;
            }, {} as ModelsOcservGroupConfig);
    });
};

onMounted(() => {
    getGroup();
});
</script>

<template>
    <v-row>
        <v-col cols="12" md="12">
            <UiParentCard :title="t('OCSERV_GROUP_DETAIL_TITLE')">
                <template #header-prepend>
                    <v-tooltip :text="t('GO_BACK_TO_GROUPS')">
                        <template #activator="{ props }">
                            <v-icon start v-bind="props" @click.stop="router.push({ name: 'Ocserv Groups' })">
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
                        @click="router.push({ name: 'Ocserv Group Update', params: { id: props.id } })"
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
                                        <span class="font-medium text-gray-600">ID :</span>
                                        <span class="ml-2 me-15 text-primary">{{ result.id }}</span>
                                    </v-col>
                                    <v-col cols="12" md="4">
                                        <span class="font-medium text-gray-600 text-capitalize">{{ t('NAME') }}:</span>
                                        <span class="ml-2 text-primary">{{ result.name }}</span>
                                    </v-col>
                                </v-row>
                            </div>
                        </div>

                        <!-- Config section -->
                        <OcservGroupConfigDetail :resultArrayObj="resultArrayObj" :resultOther="resultOther" />
                    </div>
                </UiChildCard>
            </UiParentCard>
        </v-col>
    </v-row>
</template>

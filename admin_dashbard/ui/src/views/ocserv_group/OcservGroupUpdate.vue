<script lang="ts" setup>
import UiParentCard from '@/components/shared/UiParentCard.vue';
import { useI18n } from 'vue-i18n';
import OcservGroupForm from '@/components/ocserv_group/OcservGroupForm.vue';
import UiChildCard from '@/components/shared/UiChildCard.vue';
import { type ModelsOcservGroup, type ModelsOcservGroupConfig, OcservGroupsApi } from '@/api';
import { getAuthorization } from '@/utils/request';
import { onMounted, ref } from 'vue';
import { router } from '@/router';
import { useSnackbarStore } from '@/stores/snackbar';

const props = defineProps<{ id?: number }>();

const { t } = useI18n();
const loading = ref(false);
const result = ref<ModelsOcservGroup>({ config: undefined, id: 0, name: '', owner: '' });

const api = new OcservGroupsApi();

const updateGroup = (id: number, config: ModelsOcservGroupConfig) => {
    loading.value = true;
    api.ocservGroupsIdPatch({
        ...getAuthorization(),
        id: id,
        request: { config: config }
    })
        .then(() => {
            router.push({ name: 'Ocserv Group Detail', params: { id: id } });
            const snackbar = useSnackbarStore();
            snackbar.show({
                id: 1,
                message: t('SNACK_OCSERV_GROUP_UPDATE_SUCCESS'),
                color: 'success',
                timeout: 4000
            });
        })
        .finally(() => {
            loading.value = false;
        });
};

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
    });
};

onMounted(() => {
    getGroup();
});
</script>

<template>
    <v-row>
        <v-col cols="12" md="12">
            <UiParentCard :title="t('UPDATE_OCSERV_GROUP_TITLE')">
                <template #action>
                    <v-btn
                        class="me-lg-5"
                        color="grey"
                        size="small"
                        variant="flat"
                        @click="router.push({ name: 'Ocserv Groups' })"
                    >
                        {{ t('CANCEL') }}
                    </v-btn>
                </template>
                <UiChildCard class="px-3">
                    <OcservGroupForm
                        :btnText="t('UPDATE')"
                        :initData="result"
                        :loading="loading"
                        @updateGroup="updateGroup"
                    />
                </UiChildCard>
            </UiParentCard>
        </v-col>
    </v-row>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { type ModelsOcservGroup, type ModelsOcservGroupConfig, OcservGroupsApi } from '@/api';
import { getAuthorization } from '@/utils/request';
import OcservGroupForm from '@/components/ocserv_group/OcservGroupForm.vue';
import { useI18n } from 'vue-i18n';
import UiParentCard from '@/components/shared/UiParentCard.vue';
import UiChildCard from '@/components/shared/UiChildCard.vue';
import { useSnackbarStore } from '@/stores/snackbar';

const { t } = useI18n();
const loading = ref(false);
const defaults = ref<ModelsOcservGroup>({ config: undefined, name: '', id: 0 });
const api = new OcservGroupsApi();

const getGroupDefaults = () => {
    api.ocservGroupsDefaultsGet({ ...getAuthorization() }).then((res) => {
        defaults.value = { config: res.data, name: '', id: 0 };
    });
};

const updateGroupDefaults = (id: number, config: ModelsOcservGroupConfig) => {
    loading.value = true;
    api.ocservGroupsDefaultsPatch({
        ...getAuthorization(),
        request: { config: config }
    })
        .then(() => {
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

onMounted(() => {
    getGroupDefaults();
});
</script>

<template>
    <v-row>
        <v-col cols="12" md="12">
            <UiParentCard :title="t('OCSERV_GROUP_DEFAULTS')">
                <UiChildCard class="px-3">
                    <OcservGroupForm
                        :btnText="t('UPDATE')"
                        :initData="defaults"
                        :loading="loading"
                        @updateGroup="updateGroupDefaults"
                    />
                </UiChildCard>
            </UiParentCard>
        </v-col>
    </v-row>
</template>

<style scoped lang="scss"></style>

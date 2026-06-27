<script lang="ts" setup>
import UiParentCard from '@/components/shared/UiParentCard.vue';
import { useI18n } from 'vue-i18n';
import OcservGroupForm from '@/components/ocserv_group/OcservGroupForm.vue';
import UiChildCard from '@/components/shared/UiChildCard.vue';
import { router } from '@/router';
import { type OcservGroupCreateOcservGroupData, OcservGroupsApi } from '@/api';
import { getAuthorization } from '@/utils/request';
import { ref } from 'vue';
import { useSnackbarStore } from '@/stores/snackbar';

const { t } = useI18n();
const api = new OcservGroupsApi();
const loading = ref(false);

const createGroup = (data: OcservGroupCreateOcservGroupData) => {
    loading.value = true;
    api.ocservGroupsPost({
        ...getAuthorization(),
        request: data
    })
        .then((res) => {
            const snackbar = useSnackbarStore();
            snackbar.show({
                id: 1,
                message: t('SNACK_OCSERV_GROUP_CREATE_SUCCESS'),
                color: 'success',
                timeout: 4000
            });

            router.push({ name: 'Ocserv Group Detail', params: { id: res.data.id } });
        })
        .finally(() => {
            loading.value = false;
        });
};
</script>

<template>
    <v-row>
        <v-col cols="12" md="12">
            <UiParentCard :title="t('CREATE_OCSERV_GROUP_TITLE')">
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
                    <OcservGroupForm :loading="loading" @createGroup="createGroup" :btnText="t('CREATE')" />
                </UiChildCard>
            </UiParentCard>
        </v-col>
    </v-row>
</template>

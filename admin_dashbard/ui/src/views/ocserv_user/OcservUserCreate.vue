<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import UiParentCard from '@/components/shared/UiParentCard.vue';
import { router } from '@/router';
import UiChildCard from '@/components/shared/UiChildCard.vue';
import OcservUserForm from '@/components/ocserv_user/OcservUserForm.vue';
import { OcservGroupsApi, type OcservUserCreateOcservUserData, OcservUsersApi } from '@/api';
import { getAuthorization } from '@/utils/request';
import { onBeforeMount, ref } from 'vue';
import { useSnackbarStore } from '@/stores/snackbar';

const { t } = useI18n();
const loading = ref(false);
const api = new OcservUsersApi();
const groups = ref<string[]>([]);

const createUser = (data: OcservUserCreateOcservUserData) => {
    loading.value = true;
    api.ocservUsersPost({
        ...getAuthorization(),
        request: data
    })
        .then((res) => {
            router.push({ name: 'Ocserv User Detail', params: { uid: res.data.uid } });
            const snackbar = useSnackbarStore();
            snackbar.show({
                id: 1,
                message: t('SNACK_OCSERV_USER_CREATE_SUCCESS'),
                color: 'success',
                timeout: 4000
            });
        })
        .finally(() => {
            loading.value = false;
        });
};

onBeforeMount(() => {
    const api = new OcservGroupsApi();
    api.ocservGroupsLookupGet({
        ...getAuthorization()
    }).then((res) => {
        groups.value = res.data;
    });
});
</script>

<template>
    <v-row>
        <v-col cols="12" md="12">
            <UiParentCard :title="t('CREATE_OCSERV_USER_TITLE')">
                <template #header-prepend>
                    <v-tooltip :text="t('GO_BACK_TO_USERS')">
                        <template #activator="{ props }">
                            <v-icon start v-bind="props" @click.stop="router.push({ name: 'Ocserv Users' })">
                                mdi-arrow-left-top
                            </v-icon>
                        </template>
                    </v-tooltip>
                </template>
                <UiChildCard class="px-3">
                    <OcservUserForm :groups="groups" :loading="loading" @createUser="createUser" :btnText="t('CREATE')" />
                </UiChildCard>
            </UiParentCard>
        </v-col>
    </v-row>
</template>

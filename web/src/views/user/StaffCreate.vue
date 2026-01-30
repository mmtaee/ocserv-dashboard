<script lang="ts" setup>
import { router } from '@/router';
import UiParentCard from '@/components/shared/UiParentCard.vue';
import UiChildCard from '@/components/shared/UiChildCard.vue';
import { useI18n } from 'vue-i18n';
import StaffForm from '@/components/user/StaffForm.vue';
import { ref } from 'vue';
import { type SystemCreateUserData, UsersApi } from '@/api';
import { getAuthorization } from '@/utils/request';
import { useSnackbarStore } from '@/stores/snackbar';

const { t } = useI18n();
const loading = ref(false);

const createStaff = (data: SystemCreateUserData) => {
    loading.value = true;
    const api = new UsersApi();
    api.usersPost({
        ...getAuthorization(),
        request: data
    })
        .then(() => {
            const snackbar = useSnackbarStore();
            snackbar.show({
                id: 1,
                message: t('SNACK_STAFF_CREATE_SUCCESS'),
                color: 'success',
                timeout: 4000
            });

            router.push({ name: 'Staffs' });
        })
        .finally(() => {
            loading.value = false;
        });
};
</script>

<template>
    <v-row>
        <v-col cols="12" md="12">
            <UiParentCard :title="t('CREATE_STAFF_USER_TITLE')">
                <template #header-prepend>
                    <v-tooltip :text="t('GO_BACK_TO_STAFFS')">
                        <template #activator="{ props }">
                            <v-icon start v-bind="props" @click.stop="router.push({ name: 'Staffs' })">
                                mdi-arrow-left-top
                            </v-icon>
                        </template>
                    </v-tooltip>
                </template>
                <UiChildCard class="px-3">
                    <StaffForm :loading="loading" @createStaff="createStaff" />
                </UiChildCard>
            </UiParentCard>
        </v-col>
    </v-row>
</template>

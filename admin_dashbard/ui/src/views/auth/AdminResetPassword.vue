<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { ref } from 'vue';
import AdminResetPasswordForm from '@/components/auth/AdminResetPasswordForm.vue';
import Logo from '@/layouts/full/logo/Logo.vue';
import { type SystemResetAdminPassword, SystemUserApi } from '@/api';
import { useProfileStore } from '@/stores/profile';
import { router } from '@/router';

const { t } = useI18n();
const loading = ref(false);
const api = new SystemUserApi();

const reset = (data: SystemResetAdminPassword) => {
    api.systemUserResetPasswordPost({
        request: data
    })
        .then((res) => {
            const profileStore = useProfileStore();
            profileStore.setProfile(res.data.user);
            localStorage.setItem('token', res.data.token);
            router.push({ name: 'Dashboard' });
        })
        .finally(() => {
            loading.value = false;
        });
};
</script>

<template>
    <div class="authentication">
        <v-container class="pa-3" fluid>
            <v-row class="h-100vh d-flex justify-center align-center">
                <v-col class="d-flex align-center" cols="12" lg="4" xl="3">
                    <v-card class="px-sm-1 px-0 mx-auto" elevation="10" max-width="500" rounded="md">
                        <v-card-item class="pa-sm-8">
                            <div class="d-flex justify-center py-4">
                                <Logo />
                            </div>
                            <div class="text-body-1 text-muted text-center mb-5 text-capitalize">
                                {{ t('ADMIN_RESET_PASSWORD') }}
                            </div>
                            <AdminResetPasswordForm :loading="loading" @reset="reset" />
                        </v-card-item>
                    </v-card>
                </v-col>
            </v-row>
        </v-container>
    </div>
</template>

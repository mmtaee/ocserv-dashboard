<script lang="ts" setup>
import Logo from '@/layouts/full/logo/Logo.vue';
import AdminLoginForm from '@/components/auth/AdminLoginForm.vue';
import { useI18n } from 'vue-i18n';
import { type SystemLoginData, UsersApi } from '@/api';
import { useProfileStore } from '@/stores/profile';
import { router } from '@/router';
import { ref } from 'vue';

const { t } = useI18n();
const loading = ref(false);

const signIn = (data: SystemLoginData) => {
    loading.value = true;
    const api = new UsersApi();
    api.usersLoginPost({ request: data })
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
                                {{ t('ADMIN_LOGIN_TEXT') }}
                            </div>
                            <AdminLoginForm :loading="loading" @signIn="signIn" />
                        </v-card-item>
                    </v-card>
                </v-col>
            </v-row>
        </v-container>
    </div>
</template>

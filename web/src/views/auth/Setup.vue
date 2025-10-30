<script lang="ts" setup>
import SetupLogo from '@/layouts/full/logo/SetupLogo.vue';
import SetupForm from '@/components/auth/SetupForm.vue';
import { onMounted, ref } from 'vue';
import { SystemApi, type SystemSetupSystem } from '@/api';
import { useConfigStore } from '@/stores/config';
import { useProfileStore } from '@/stores/profile';
import { router } from '@/router';

const loading = ref(false);
const showOverlay = ref(true);
const currentMessage = ref('');
const messages = [
    'Welcome to the <span class="text-primary">Ocserv Dashboard</span>',
    'Before you can use this panel, create an <span class="text-primary">Admin</span> user',
    'Optionally, you can set up <span class="text-primary">reCAPTCHA v2</span> for the login page',
    "Let's Go"
];

const setup = (data: SystemSetupSystem) => {
    loading.value = true;
    const api = new SystemApi();
    api.systemSetupPost({
        request: data
    })
        .then((res) => {
            const configStore = useConfigStore();
            const profileStore = useProfileStore();

            configStore.setConfig(res.data.system.google_captcha_site_key);
            profileStore.setProfile(res.data.user);
            localStorage.setItem('token', res.data.token);
            router.push({ name: 'Dashboard' });
        })
        .finally(() => {
            loading.value = false;
        });
};

onMounted(() => {
    let index = 0;

    const showNextMessage = () => {
        if (index < messages.length) {
            currentMessage.value = messages[index];

            const isLast = index === messages.length - 1;
            const displayTime = isLast ? 3000 : 6000;

            index++;

            setTimeout(() => {
                currentMessage.value = '';
                setTimeout(showNextMessage, 500);
            }, displayTime);
        } else {
            showOverlay.value = false;
        }
    };

    showNextMessage();
});
</script>
<template>
    <!-- Welcome Overlay -->
    <v-overlay v-model="showOverlay" class="align-center justify-center" opacity="0">
        <transition name="fade" mode="in-out" type="animation" appear>
            <h1 v-if="currentMessage" key="message" v-html="currentMessage"></h1>
        </transition>
    </v-overlay>

    <!-- Main Content -->
    <div v-if="!showOverlay" class="authentication">
        <v-container class="pa-3" fluid>
            <v-row class="h-100vh d-flex justify-center align-center">
                <v-col class="d-flex align-center" cols="12" lg="4" xl="3">
                    <v-card class="px-sm-1 px-0 mx-auto" elevation="10" max-width="500" rounded="md">
                        <v-card-item class="pa-sm-8">
                            <div class="d-flex justify-center">
                                <SetupLogo />
                            </div>
                            <SetupForm :loading="loading" @setup="setup" />
                        </v-card-item>
                    </v-card>
                </v-col>
            </v-row>
        </v-container>
    </div>
</template>

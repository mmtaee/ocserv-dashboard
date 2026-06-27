<script setup lang="ts">
import Logo from '@/layouts/full/logo/Logo.vue';
import UiChildCard from '@/components/shared/UiChildCard.vue';
import type { CustomerCiscoSetupResponse } from '@/api';
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

const { t } = useI18n();
const router = useRouter();

const appStoreUrl = 'itms-apps://apps.apple.com/app/id1135064690';
const playStoreUrl = 'market://details?id=com.cisco.anyconnect.vpn.android.avf';

const setup = ref<CustomerCiscoSetupResponse | null>(loadSetup());

const hasSetup = computed(() => setup.value !== null);

function loadSetup(): CustomerCiscoSetupResponse | null {
    const raw = sessionStorage.getItem('customerCiscoSetup');
    if (!raw) return null;

    try {
        return JSON.parse(raw) as CustomerCiscoSetupResponse;
    } catch {
        sessionStorage.removeItem('customerCiscoSetup');
        return null;
    }
}

const copyCertificatePassword = (password: string) => {
    navigator.clipboard?.writeText(password);
};

const goBack = () => {
    router.push({ name: 'Customer Summary' });
};
</script>

<template>
    <div class="authentication cisco-setup-page">
        <v-container class="pa-3 pa-sm-6" fluid>
            <v-row class="cisco-setup-row d-flex justify-center align-start align-sm-center">
                <v-col cols="12" sm="10" md="8" lg="7" xl="6">
                    <UiChildCard class="px-2 px-sm-3">
                        <template #title-header>
                            <div class="d-flex align-center flex-column flex-sm-row text-center text-sm-start">
                                <Logo />
                                <span class="mt-3 mt-sm-0 ms-sm-4 text-primary text-h4 text-sm-h3">
                                    {{ t('CISCO_SETUP_TITLE') }}
                                </span>
                            </div>
                        </template>

                        <template #action>
                            <v-spacer />
                            <v-btn block class="mt-3 mt-sm-0" color="primary" variant="outlined" @click="goBack">
                                {{ t('BACK') }}
                            </v-btn>
                        </template>

                        <div v-if="!hasSetup" class="mt-6">
                            <v-alert class="mb-4" type="warning" variant="tonal">
                                {{ t('CISCO_SETUP_MISSING_DATA') }}
                            </v-alert>

                            <v-btn block color="primary" to="/summary">
                                {{ t('BACK') }}
                            </v-btn>
                        </div>

                        <div v-if="setup" class="space-y-4 mt-8 px-0 px-sm-1">
                            <v-alert class="mb-4" type="info" variant="tonal">
                                {{ t('CISCO_SETUP_SUPPORTED_PLATFORMS') }}
                            </v-alert>

                            <v-list lines="three">
                                <v-list-item>
                                    <template #prepend>
                                        <v-avatar color="primary">1</v-avatar>
                                    </template>

                                    <v-list-item-title>{{ t('CISCO_SETUP_STEP_INSTALL_APP') }}</v-list-item-title>
                                    <v-list-item-subtitle>
                                        {{ t('CISCO_SETUP_STEP_INSTALL_APP_DESC') }}
                                    </v-list-item-subtitle>

                                    <v-row class="mt-3">
                                        <v-col cols="12" sm="6">
                                            <v-btn
                                                block
                                                class="setup-action-btn text-none"
                                                color="primary"
                                                :href="appStoreUrl"
                                            >
                                                {{ t('DOWNLOAD_FROM_APP_STORE') }}
                                            </v-btn>
                                        </v-col>

                                        <v-col cols="12" sm="6">
                                            <v-btn
                                                block
                                                class="setup-action-btn text-none"
                                                color="success"
                                                :href="playStoreUrl"
                                            >
                                                {{ t('DOWNLOAD_FROM_PLAY_STORE') }}
                                            </v-btn>
                                        </v-col>
                                    </v-row>
                                </v-list-item>

                                <v-divider class="my-3" />

                                <v-list-item>
                                    <template #prepend>
                                        <v-avatar color="primary">2</v-avatar>
                                    </template>

                                    <v-list-item-title>{{ t('CISCO_SETUP_STEP_EXTERNAL_CONTROL') }}</v-list-item-title>
                                    <v-list-item-subtitle>
                                        {{ t('CISCO_SETUP_STEP_EXTERNAL_CONTROL_DESC') }}
                                    </v-list-item-subtitle>
                                </v-list-item>

                                <v-divider class="my-3" />

                                <v-list-item>
                                    <template #prepend>
                                        <v-avatar color="primary">3</v-avatar>
                                    </template>

                                    <v-list-item-title>{{ t('CISCO_SETUP_STEP_IMPORT_CERTIFICATE') }}</v-list-item-title>
                                    <v-list-item-subtitle>
                                        {{ t('CISCO_SETUP_STEP_IMPORT_CERTIFICATE_DESC') }}
                                    </v-list-item-subtitle>

                                    <div class="mt-3">
                                        <v-btn
                                            block
                                            class="setup-action-btn text-none"
                                            color="primary"
                                            :href="setup.certificate_import_uri"
                                        >
                                            {{ t('CISCO_IMPORT_CERTIFICATE') }}
                                        </v-btn>
                                    </div>
                                </v-list-item>

                                <v-divider class="my-3" />

                                <v-list-item>
                                    <template #prepend>
                                        <v-avatar color="primary">4</v-avatar>
                                    </template>

                                    <v-list-item-title>{{ t('CISCO_SETUP_STEP_ENTER_PASSWORD') }}</v-list-item-title>
                                    <v-list-item-subtitle>
                                        {{ t('CISCO_SETUP_STEP_ENTER_PASSWORD_DESC') }}
                                    </v-list-item-subtitle>

                                    <v-alert class="mt-3" type="warning" variant="tonal">
                                        {{ t('CISCO_SETUP_PASSWORD_HINT') }}
                                        <strong>{{ setup.certificate_password }}</strong>
                                        <v-btn
                                            class="ms-2"
                                            size="x-small"
                                            variant="text"
                                            @click="copyCertificatePassword(setup.certificate_password)"
                                        >
                                            {{ t('COPY') }}
                                        </v-btn>
                                    </v-alert>
                                </v-list-item>

                                <v-divider class="my-3" />

                                <v-list-item>
                                    <template #prepend>
                                        <v-avatar color="primary">5</v-avatar>
                                    </template>

                                    <v-list-item-title>{{ t('CISCO_SETUP_STEP_ADD_CONNECTION') }}</v-list-item-title>
                                    <v-list-item-subtitle>
                                        {{ t('CISCO_SETUP_STEP_ADD_CONNECTION_DESC') }}
                                    </v-list-item-subtitle>

                                    <div class="mt-3">
                                        <v-btn
                                            block
                                            class="setup-action-btn text-none"
                                            color="success"
                                            :href="setup.connection_create_uri"
                                        >
                                            {{ t('CISCO_ADD_CONNECTION') }}
                                        </v-btn>
                                    </div>
                                </v-list-item>

                                <v-divider class="my-3" />

                                <v-list-item>
                                    <template #prepend>
                                        <v-avatar color="primary">6</v-avatar>
                                    </template>

                                    <v-list-item-title>{{ t('CISCO_SETUP_STEP_CONNECT') }}</v-list-item-title>
                                    <v-list-item-subtitle>
                                        {{ t('CISCO_SETUP_STEP_CONNECT_DESC') }}
                                    </v-list-item-subtitle>
                                </v-list-item>
                            </v-list>

                            <v-alert class="mt-4" type="info" variant="tonal">
                                {{ t('CISCO_SETUP_EXPIRES_HINT') }}
                            </v-alert>
                        </div>
                    </UiChildCard>
                </v-col>
            </v-row>
        </v-container>
    </div>
</template>

<style scoped>
.cisco-setup-page {
    min-height: 100svh;
}

.cisco-setup-row {
    min-height: calc(100svh - 24px);
    padding-top: 16px;
    padding-bottom: 16px;
}

.setup-action-btn {
    min-height: 44px;
    height: auto;
    white-space: normal;
    text-align: center;
    line-height: 1.25;
    padding-top: 8px;
    padding-bottom: 8px;
    text-transform: none;
}

.setup-action-btn :deep(.v-btn__content) {
    white-space: normal;
    overflow: visible;
    text-overflow: unset;
    line-height: 1.25;
}

@media (min-width: 600px) {
    .cisco-setup-row {
        min-height: calc(100svh - 48px);
    }
}
</style>

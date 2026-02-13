<script setup lang="ts">
import UiChildCard from '@/components/shared/UiChildCard.vue';
import UiParentCard from '@/components/shared/UiParentCard.vue';
import { SystemApi, type SystemGetSystemResponse } from '@/api';
import { useI18n } from 'vue-i18n';
import { onMounted, ref } from 'vue';
import { getAuthorization } from '@/utils/request';
import SystemForm from '@/components/system/SystemForm.vue';

const { t } = useI18n();
const updateMode = ref(false);

const api = new SystemApi();

const systemData = ref<SystemGetSystemResponse>({});

const updateSystem = () => {
    // api.systemPatch({
    //     ...getAuthorization(),
    //
    // })
};

onMounted(() => {
    api.systemGet({
        ...getAuthorization()
    }).then((res) => {
        Object.assign(systemData.value, res.data);
        console.log(res.data);
    });
});
</script>

<template>
    <v-row>
        <v-col cols="12" md="12">
            <UiParentCard :title="t('SYSTEM')">
                <template #action>
                    <v-btn
                        v-if="!updateMode"
                        color="grey"
                        density="compact"
                        variant="outlined"
                        @click="updateMode = true"
                    >
                        {{ t('UPDATE') }}
                    </v-btn>
                </template>

                <UiChildCard v-if="!updateMode" :title="t('CONFIGS')" class="px-3">
                    <v-row align="start" justify="center">
                        <v-col>
                            <v-list>
                                <v-list-item class="mb-3">
                                    <template #prepend>
                                        <v-icon size="large">mdi-shield-check</v-icon>
                                    </template>
                                    <v-list-item-title class="text-subtitle-2 text-capitalize mb-2">
                                        {{ t('GOOGLE_CAPTCHA_SITE_KEY') }}
                                    </v-list-item-title>
                                    <v-list-item-subtitle class="text-subtitle-1">
                                        {{ systemData.google_captcha_site_key || t('NOT_SET') }}
                                    </v-list-item-subtitle>
                                </v-list-item>
                                <v-list-item class="mb-3">
                                    <template #prepend>
                                        <v-icon size="large">mdi-key</v-icon>
                                    </template>
                                    <v-list-item-title class="text-subtitle-2 text-capitalize mb-2">
                                        {{ t('GOOGLE_CAPTCHA_SECRET_KEY') }}
                                    </v-list-item-title>
                                    <v-list-item-subtitle class="text-subtitle-1">
                                        {{ systemData.google_captcha_secret_key || t('NOT_SET') }}
                                    </v-list-item-subtitle>
                                </v-list-item>

                                <v-list-item class="mb-3">
                                    <template #prepend>
                                        <v-icon size="large">mdi-delete-sweep</v-icon>
                                    </template>
                                    <v-list-item-title class="text-subtitle-2 text-capitalize mb-2">
                                        {{ t('AUTO_DELETE_INACTIVE_USERS') }}
                                    </v-list-item-title>
                                    <v-list-item-subtitle class="text-subtitle-1">
                                        {{ systemData.auto_delete_inactive_users }}
                                    </v-list-item-subtitle>
                                </v-list-item>

                                <v-list-item class="mb-3">
                                    <template #prepend>
                                        <v-icon size="large">mdi-timer-sand</v-icon>
                                    </template>
                                    <v-list-item-title class="text-subtitle-2 text-capitalize mb-2">
                                        {{ t('KEEP_INACTIVE_USER_DAYS') }}
                                    </v-list-item-title>
                                    <v-list-item-subtitle class="text-subtitle-1">
                                        {{ systemData.keep_inactive_user_days }}
                                    </v-list-item-subtitle>
                                </v-list-item>
                            </v-list>
                        </v-col>
                    </v-row>
                </UiChildCard>

                <UiChildCard v-else :title="t('CONFIGS')" class="px-3">
                    <v-col cols="6" md="6" sm="12">
                        <SystemForm :data="systemData" @updateSystem="updateSystem" @cancel="updateMode = false" />
                    </v-col>
                </UiChildCard>
            </UiParentCard>
        </v-col>
    </v-row>
</template>

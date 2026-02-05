<script lang="ts" setup>
import UiParentCard from '@/components/shared/UiParentCard.vue';
import { useI18n } from 'vue-i18n';
import UiChildCard from '@/components/shared/UiChildCard.vue';
import { useProfileStore } from '@/stores/profile';
import { formatDateTimeWithRelative } from '@/utils/convertors';
import { reactive, ref } from 'vue';
import { type SystemChangeUserPasswordBySelf, UsersApi } from '@/api';
import { requiredRule } from '@/utils/rules';
import { getAuthorization } from '@/utils/request';
import { useSnackbarStore } from '@/stores/snackbar';

type DateKeys = 'created_at' | 'updated_at' | 'last_login';

const profileStore = useProfileStore();
const { t } = useI18n();

const passwordForm = ref();
const loading = ref(false);
const showChangePassword = ref(false);
const validPassword = ref(false);
const showOldPassword = ref(false);
const showNewPassword = ref(false);
const data = reactive<SystemChangeUserPasswordBySelf>({
    old_password: '',
    new_password: ''
});

const rules = {
    required: (v: string) => requiredRule(v, t)
};

const resetForm = () => {
    passwordForm.value?.reset();
};

const profileDateProxy = (key: DateKeys): string => {
    return formatDateTimeWithRelative(profileStore.profile?.[key], t('NOT_AVAILABLE'));
};

const changePassword = () => {
    loading.value = true;
    const api = new UsersApi();
    api.usersPasswordPost({
        ...getAuthorization(),
        request: data
    })
        .then(() => {
            const snackbar = useSnackbarStore();
            snackbar.show({
                id: 1,
                message: t('SNACK_STAFF_PASSWORD_UPDATE_SUCCESS'),
                color: 'success',
                timeout: 4000
            });
            showChangePassword.value = false;
        })
        .finally(() => {
            loading.value = true;
            showChangePassword.value = false;
        });
};
</script>

<template>
    <v-row>
        <v-col cols="12" md="12">
            <UiParentCard :title="t('PROFILE')">
                <UiChildCard :title="t('INFO')" class="px-3">
                    <v-row align="start" justify="center">
                        <v-col>
                            <v-list>
                                <v-list-item class="mb-3">
                                    <template #prepend>
                                        <v-icon size="large">mdi-account-circle</v-icon>
                                    </template>
                                    <v-list-item-title class="text-subtitle-2 text-capitalize mb-2">
                                        {{ t('USERNAME') }}
                                    </v-list-item-title>
                                    <v-list-item-subtitle class="text-subtitle-1">
                                        {{ profileStore.profile?.username || 'guest' }}
                                    </v-list-item-subtitle>
                                </v-list-item>
                                <v-list-item class="mb-3">
                                    <template #prepend>
                                        <v-icon size="large">mdi-fingerprint</v-icon>
                                    </template>
                                    <v-list-item-title class="text-subtitle-2 text-capitalize mb-2">
                                        UID
                                    </v-list-item-title>
                                    <v-list-item-subtitle class="text-subtitle-1">
                                        {{ profileStore.profile?.uid || '—' }}
                                    </v-list-item-subtitle>
                                </v-list-item>
                                <v-list-item class="mb-3">
                                    <template #prepend>
                                        <v-icon v-if="profileStore.profile?.role" size="large">
                                            mdi-shield-account
                                        </v-icon>
                                        <v-icon v-else size="large"> mdi-account-outline</v-icon>
                                    </template>
                                    <v-list-item-title class="text-subtitle-2 text-capitalize mb-2">
                                        {{ t('ROLE') }}
                                    </v-list-item-title>
                                    <v-list-item-subtitle class="text-subtitle-1">
                                        {{ profileStore.profile?.role == 'admin' ? t('ADMIN') : t('STAFF') }}
                                    </v-list-item-subtitle>
                                </v-list-item>
                            </v-list>
                        </v-col>

                        <v-divider opacity="1" vertical />

                        <v-col>
                            <v-list>
                                <v-list-item class="mb-3">
                                    <template #prepend>
                                        <v-icon size="large">mdi-calendar</v-icon>
                                    </template>
                                    <v-list-item-title class="text-subtitle-2 text-capitalize mb-2">
                                        {{ t('ACCOUNT_CREATED') }}
                                    </v-list-item-title>
                                    <v-list-item-subtitle class="text-subtitle-1">
                                        {{ profileDateProxy('created_at') }}
                                    </v-list-item-subtitle>
                                </v-list-item>
                                <v-list-item class="mb-3">
                                    <template #prepend>
                                        <v-icon size="large">mdi-clock-outline</v-icon>
                                    </template>
                                    <v-list-item-title class="text-subtitle-2 text-capitalize mb-2">
                                        {{ t('LAST_LOGIN') }}
                                    </v-list-item-title>
                                    <v-list-item-subtitle class="text-subtitle-1 text-capitalize">
                                        {{ profileDateProxy('last_login') }}
                                    </v-list-item-subtitle>
                                </v-list-item>
                                <v-list-item class="mb-3">
                                    <template #prepend>
                                        <v-icon size="large">mdi-update</v-icon>
                                    </template>
                                    <v-list-item-title class="text-subtitle-2 text-capitalize mb-2">
                                        {{ t('LAST_UPDATED') }}
                                    </v-list-item-title>
                                    <v-list-item-subtitle class="text-subtitle-1 text-capitalize">
                                        {{ profileDateProxy('updated_at') }}
                                    </v-list-item-subtitle>
                                </v-list-item>
                            </v-list>
                        </v-col>
                    </v-row>
                </UiChildCard>

                <UiChildCard class="px-3">
                    <v-icon class="me-2" start>mdi-shield-key-outline</v-icon>

                    <span class="text-capitalize">{{ t('PASSWORD') }}</span>

                    <v-btn
                        v-if="!showChangePassword"
                        color="primary"
                        density="compact"
                        variant="plain"
                        @click="showChangePassword = true"
                    >
                        {{ t('CHANGE') }}
                    </v-btn>
                    <v-btn
                        v-if="showChangePassword"
                        color="primary"
                        density="compact"
                        variant="plain"
                        @click="(resetForm, (showChangePassword = false))"
                    >
                        {{ t('CANCEL') }}
                    </v-btn>

                    <div v-if="showChangePassword" class="mt-5">
                        <v-form ref="passwordForm" v-model="validPassword">
                            <v-row align="center" class="mx-2" justify="start">
                                <v-col md="4">
                                    <v-label class="font-weight-bold mb-1 text-capitalize">
                                        {{ t('OLD_PASSWORD') }}
                                    </v-label>
                                    <v-text-field
                                        v-model="data.old_password"
                                        :append-inner-icon="showOldPassword ? 'mdi-eye-off' : 'mdi-eye'"
                                        :rules="[rules.required]"
                                        :type="showOldPassword ? 'text' : 'password'"
                                        autocomplete="new-password"
                                        color="primary"
                                        prepend-inner-icon="mdi-key"
                                        variant="outlined"
                                        @click:append-inner="showOldPassword = !showOldPassword"
                                    />
                                </v-col>

                                <v-col md="4">
                                    <v-label class="font-weight-bold mb-1 text-capitalize">
                                        {{ t('NEW_PASSWORD') }}
                                    </v-label>
                                    <v-text-field
                                        v-model="data.new_password"
                                        :append-inner-icon="showNewPassword ? 'mdi-eye-off' : 'mdi-eye'"
                                        :rules="[rules.required]"
                                        :type="showNewPassword ? 'text' : 'password'"
                                        autocomplete="new-password"
                                        color="primary"
                                        prepend-inner-icon="mdi-key"
                                        variant="outlined"
                                        @click:append-inner="showNewPassword = !showNewPassword"
                                    />
                                </v-col>

                                <v-col md="auto">
                                    <v-btn
                                        :disabled="!validPassword || data.new_password == data.old_password"
                                        :loading="loading"
                                        color="primary"
                                        @click="changePassword"
                                    >
                                        {{ t('CHANGE_PASSWORD') }}
                                    </v-btn>
                                </v-col>
                            </v-row>
                        </v-form>
                    </div>
                </UiChildCard>
            </UiParentCard>
        </v-col>
    </v-row>
</template>

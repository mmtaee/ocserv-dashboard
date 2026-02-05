<script lang="ts" setup>
import { computed, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { type SystemLoginData } from '@/api';
import { requiredRule } from '@/utils/rules';
import Captcha from '@/components/auth/Captcha.vue';
import { useConfigStore } from '@/stores/config';

defineProps({
    loading: Boolean
});

const emit = defineEmits(['signIn']);

const configStore = useConfigStore();
const siteKey = configStore.config.googleCaptchaSiteKey;
const { t } = useI18n();
const valid = ref(true);
const showPassword = ref(false);
const data = reactive<SystemLoginData>({
    password: '',
    remember_me: false,
    token: '',
    username: ''
});
const rules = {
    required: (v: string) => requiredRule(v, t)
};

const isValid = computed(() => {
    if (siteKey === '') return !valid.value;
    return !(data.token?.trim() !== '' && valid.value);
});
</script>

<template>
    <v-form v-model="valid">
        <v-row class="d-flex mb-3">
            <v-col cols="12">
                <v-label class="font-weight-bold mb-1 text-capitalize">{{ t('ADMIN_USERNAME') }}</v-label>
                <v-text-field
                    v-model="data.username"
                    :rules="[rules.required]"
                    color="primary"
                    hide-details
                    variant="outlined"
                />
            </v-col>
            <v-col cols="12">
                <v-label class="font-weight-bold mb-1">{{ t('ADMIN_PASSWORD') }}</v-label>
                <v-text-field
                    v-model="data.password"
                    :append-inner-icon="showPassword ? 'mdi-eye-off' : 'mdi-eye'"
                    :rules="[rules.required]"
                    :type="showPassword ? 'text' : 'password'"
                    autocomplete="new-password"
                    color="primary"
                    hide-details
                    variant="outlined"
                    @click:append-inner="showPassword = !showPassword"
                    @keydown.enter="emit('signIn', data)"
                />
            </v-col>

            <v-col v-if="siteKey !== ''" cols="12">
                <v-row justify="center">
                    <Captcha v-model="data.token" :siteKey="siteKey" />
                </v-row>
            </v-col>

            <v-col class="pt-0" cols="12">
                <div class="d-flex flex-wrap align-center ml-n2">
                    <v-checkbox v-model="data.remember_me" color="primary" hide-details class="text-capitalize">
                        <template v-slot:label class="text-body-1">{{ t('REMEMBER_ME') }}</template>
                    </v-checkbox>
                    <div class="ml-sm-auto">
                        <v-tooltip location="top">
                            <template #activator="{ props }">
                                <span class="cursor-pointer text-primary text-subtitle-2 text-capitalize" v-bind="props">
                                    {{ t('FORGOT_PASSWORD') }} ?
                                </span>
                            </template>
                            <span v-html="t('FORGOT_PASSWORD_TOOLTIP_HELP')" />
                        </v-tooltip>
                    </div>
                </div>
            </v-col>
            <v-col class="pt-0" cols="12">
                <v-btn
                    :loading="loading"
                    :disabled="isValid"
                    block
                    color="primary"
                    flat
                    size="large"
                    @click="emit('signIn', data)"
                >
                    {{ t('SIGN_IN') }}
                </v-btn>
            </v-col>
        </v-row>
    </v-form>
</template>

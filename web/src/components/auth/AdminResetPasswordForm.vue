<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { type SystemResetAdminPassword } from '@/api';
import { requiredRule } from '@/utils/rules';

defineProps({
    loading: Boolean
});

const emit = defineEmits(['reset']);

const { t } = useI18n();
const valid = ref(true);
const showPassword = ref(false);
const data = reactive<SystemResetAdminPassword>({
    new_password: '',
    secret_key: '',
    username: ''
});

const rules = {
    required: (v: string) => requiredRule(v, t)
};
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
                <v-label class="font-weight-bold mb-1 text-capitalize">{{ t('ADMIN_NEW_PASSWORD') }}</v-label>
                <v-text-field
                    v-model="data.new_password"
                    :append-inner-icon="showPassword ? 'mdi-eye-off' : 'mdi-eye'"
                    :rules="[rules.required]"
                    :type="showPassword ? 'text' : 'password'"
                    autocomplete="new-password"
                    color="primary"
                    hide-details
                    variant="outlined"
                    @click:append-inner="showPassword = !showPassword"
                />
            </v-col>
            <v-col cols="12">
                <v-label class="font-weight-bold mb-1 text-capitalize">
                    {{ t('SECRET_KEY') }}

                    <v-tooltip location="top">
                        <template #activator="{ props }">
                            <v-icon v-bind="props" size="x-small" class="ms-2 mt-1"> mdi-help-circle-outline </v-icon>
                        </template>
                        <span>
                            {{ t('USE_SECRET_KEY_FROM_ENV_FILE_IN_SERVER') }}
                        </span>
                    </v-tooltip>
                </v-label>

                <v-text-field
                    v-model="data.secret_key"
                    :rules="[rules.required]"
                    color="primary"
                    variant="outlined"
                    hide-details
                    @keydown.enter="emit('reset', data)"
                />
            </v-col>
            <v-col cols="12">
                <v-btn
                    :loading="loading"
                    :disabled="!valid"
                    block
                    color="primary"
                    flat
                    size="large"
                    @click="emit('reset', data)"
                >
                    {{ t('RESET_PASSWORD') }}
                </v-btn>
            </v-col>
        </v-row>
    </v-form>
</template>

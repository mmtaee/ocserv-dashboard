<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { requiredRule } from '@/utils/rules';
import { reactive, ref } from 'vue';
import type { SystemCreateUserData } from '@/api';

defineProps({
    loading: {
        type: Boolean,
        default: false
    }
});

const emit = defineEmits(['createStaff']);

const { t } = useI18n();
const valid = ref(false);
const showPassword = ref(false);
const rules = {
    required: (v: string) => requiredRule(v, t)
};

const createData = reactive<SystemCreateUserData>({ password: '', username: '' });

const createStaff = () => {
    emit('createStaff', createData);
};
</script>

<template>
    <v-form v-model="valid">
        <v-row align="center" justify="start">
            <v-col cols="12">
                <h3 class="text-capitalize">{{ t('MAIN') }}</h3>
            </v-col>
            <v-col cols="12" lg="4" md="6">
                <v-label class="font-weight-bold mb-1 text-capitalize">{{ t('USERNAME') }}</v-label>
                <v-text-field
                    v-model="createData.username"
                    :rules="[rules.required]"
                    color="primary"
                    hide-details
                    variant="outlined"
                />
            </v-col>
            <v-col cols="12" lg="4" md="6">
                <v-label class="font-weight-bold mb-1 text-capitalize">{{ t('PASSWORD') }}</v-label>
                <v-text-field
                    v-model="createData.password"
                    :append-inner-icon="showPassword ? 'mdi-eye-off-outline' : 'mdi-eye-outline'"
                    :rules="[rules.required]"
                    :type="showPassword ? 'text' : 'password'"
                    autocomplete="new-password"
                    color="primary"
                    hide-details
                    variant="outlined"
                    @click:append-inner="showPassword = !showPassword"
                />
            </v-col>
        </v-row>
        <v-row align="center" class="me-0 mt-5" justify="end">
            <v-col cols="auto">
                <v-btn :disabled="!valid" :loading="loading" color="primary" @click="createStaff()">
                    {{ t('CREATE') }}
                </v-btn>
            </v-col>
        </v-row>
    </v-form>
</template>

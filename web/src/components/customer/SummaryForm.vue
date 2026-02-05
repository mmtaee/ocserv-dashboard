<script lang="ts" setup>
import { computed, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { type CustomerSummaryData } from '@/api';
import { requiredRule } from '@/utils/rules';
// import Captcha from '@/components/auth/Captcha.vue';
import { useConfigStore } from '@/stores/config';
import { useProfileStore } from '@/stores/profile';
import { router } from '@/router';

defineProps({
    loading: {
        type: Boolean,
        default: false
    }
});

const emit = defineEmits(['getSummary']);

const { t } = useI18n();
const valid = ref(true);
const showPassword = ref(false);
const data = reactive<CustomerSummaryData>({
    username: '',
    password: ''
});
const rules = {
    required: (v: string) => requiredRule(v, t)
};

const getSummary = () => {
    emit('getSummary', data);
};
</script>

<template>
    <v-form v-model="valid">
        <v-row class="d-flex mb-3">
            <v-col cols="12">
                <v-label class="font-weight-bold mb-1 text-capitalize">{{ t('OCSERV_USERNAME') }}</v-label>
                <v-text-field
                    v-model="data.username"
                    :rules="[rules.required]"
                    color="primary"
                    hide-details
                    variant="outlined"
                />
            </v-col>
            <v-col cols="12">
                <v-label class="font-weight-bold mb-1">{{ t('OCSERV_PASSWORD') }}</v-label>
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
                />
            </v-col>
            <v-col class="pt-0 mt-3" cols="12">
                <v-btn :loading="loading" :disabled="!valid" block color="primary" flat size="large" @click="getSummary">
                    {{ t('GET_MY_SUMMARY') }}
                </v-btn>
            </v-col>
        </v-row>
    </v-form>
</template>

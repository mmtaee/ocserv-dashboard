<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { ref } from 'vue';
import { requiredRule } from '@/utils/rules';

const props = defineProps({
    username: {
        type: String,
        required: true
    },
    show: {
        type: Boolean,
        default: false
    }
});

const emits = defineEmits(['changePassword', 'close']);

const { t } = useI18n();
const password = ref<String>('');
const showPassword = ref(false);

const rules = {
    required: (v: string) => requiredRule(v, t)
};
</script>
<template>
    <v-dialog v-model="props.show" max-width="500">
        <v-card>
            <v-card-title class="bg-warning text-capitalize">
                {{ t('CHANGE_STAFF_PASSWORD_DIALOG_TITLE') }}
            </v-card-title>

            <v-card-text>
                <v-row align="center" justify="start">
                    <v-col class="ma-0 pa-1 pt-3 text-capitalize" cols="12" md="12">
                        {{ t('CHANGE_STAFF_PASSWORD_DIALOG_TEXT') }}?
                    </v-col>
                    <v-col class="ma-0 pa-1 mt-1" cols="12" md="12">
                        <div>
                            <span class="text-capitalize"> {{ t('STAFF') }} </span>:
                            <span class="text-error text-capitalize font-weight-bold">{{ username }}</span>
                        </div>
                    </v-col>

                    <v-col class="ma-0 pa-1" cols="12" md="12">
                        <v-label class="font-weight-bold mb-1 text-capitalize">{{ t('PASSWORD') }}</v-label>
                        <v-text-field
                            v-model="password"
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
            </v-card-text>

            <v-card-actions class="mx-2 my-1">
                <v-spacer></v-spacer>
                <v-btn color="grey" variant="tonal" @click="emits('close')">
                    {{ t('CANCEL') }}
                </v-btn>
                <v-btn color="warning" variant="flat" @click="emits('changePassword', password)">
                    {{ t('CHANGE') }}
                </v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>
</template>

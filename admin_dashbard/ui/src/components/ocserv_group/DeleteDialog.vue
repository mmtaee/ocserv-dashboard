<script lang="ts" setup>
import { useI18n } from 'vue-i18n';

const props = defineProps({
    name: {
        type: String,
        required: true
    },
    show: {
        type: Boolean,
        default: false
    }
});

const emits = defineEmits(['deleteGroup', 'close']);

const { t } = useI18n();
</script>

<template>
    <v-dialog v-model="props.show" max-width="500">
        <v-card>
            <v-card-title class="bg-error text-capitalize">
                {{ t('DELETE_GROUP_DIALOG_TITLE') }}
            </v-card-title>

            <v-card-text>
                <span class="text-capitalize"> {{ t('DELETE_GROUP_DIALOG_TEXT') }}? </span>
                <div>
                    <span class="text-capitalize">{{ t('OCSERV_GROUP') }}: </span>
                    <span class="text-error text-capitalize font-weight-bold">{{ name }}</span>
                </div>
            </v-card-text>

            <v-card-actions class="mx-2 my-1">
                <v-spacer></v-spacer>

                <v-btn color="grey" variant="tonal" @click="emits('close')">
                    {{ t('CANCEL') }}
                </v-btn>
                <v-btn color="error" variant="flat" @click="emits('deleteGroup')">
                    {{ t('DELETE') }}
                </v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>
</template>

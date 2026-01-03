<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { ref } from 'vue';
import { formatDate } from '@/utils/convertors';

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

const emits = defineEmits(['activateUser', 'close']);

const { t } = useI18n();

const expireAt = ref(null);
const showDateMenu = ref(false);
</script>

<template>
    <v-dialog v-model="props.show" max-width="500">
        <v-card>
            <v-card-title class="bg-info text-capitalize">
                {{ t('ACTIVATE_USER_TITLE') }}
            </v-card-title>

            <v-card-text>
                <v-row>
                    <v-col cols="12" lg="12" md="12">
                        <div>
                            <span class="text-capitalize">{{ t('OCSERV_USER') }}: </span>
                            <span class="text-info text-capitalize font-weight-bold">{{ username }}</span>
                        </div>
                    </v-col>
                    <v-col cols="12">
                        <v-menu v-model="showDateMenu" :close-on-content-click="false" transition="scale-transition">
                            <template #activator="{ props }">
                                <v-label class="font-weight-bold mb-1 text-capitalize">
                                    {{ t('NEW_EXPIRE_AT') }}
                                </v-label>
                                <v-text-field
                                    :model-value="expireAt ? formatDate(expireAt) : ''"
                                    color="primary"
                                    hide-details
                                    readonly
                                    v-bind="props"
                                    variant="outlined"
                                    clearable
                                    @click:clear="expireAt = null"
                                />
                            </template>
                            <v-date-picker
                                v-model="expireAt"
                                :header="t('EXPIRE_AT')"
                                elevation="24"
                                title=""
                                @update:model-value="() => (showDateMenu = false)"
                            />
                        </v-menu>
                    </v-col>
                </v-row>
            </v-card-text>

            <v-card-actions class="mx-2 my-1">
                <v-spacer></v-spacer>

                <v-btn color="grey" variant="tonal" @click="emits('close')">
                    {{ t('CANCEL') }}
                </v-btn>
                <v-btn color="info" variant="flat" @click="emits('activateUser', expireAt)">
                    {{ t('ACTIVATE') }}
                </v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>
</template>

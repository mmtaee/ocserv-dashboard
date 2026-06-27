<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import type { ModelsOnlineUserSession } from '@/api';
import type { PropType } from 'vue';

const props = defineProps({
    username: {
        type: String,
        required: true
    },
    sessions: {
        type: Array as PropType<ModelsOnlineUserSession[]>,
        required: true
    },
    show: {
        type: Boolean,
        default: false
    }
});

const emits = defineEmits(['disconnect', 'terminate', 'close']);

const { t } = useI18n();
</script>

<template>
    <v-dialog v-model="props.show" max-width="650">
        <v-card>
            <v-card-title class="bg-primary text-capitalize">
                <v-row align="end" justify="space-between" class="no-gutters">
                    <v-col md="auto">{{ t('DISCONNECT_USER_DIALOG_TITLE') }} </v-col>
                    <v-col md="auto">
                        <v-icon @click="emits('close')">mdi-close</v-icon>
                    </v-col>
                </v-row>
            </v-card-title>

            <v-card-text>
                <span class="text-capitalize"> {{ t('DISCONNECT_USER_DIALOG_TEXT') }}? </span>
                <div>
                    <span class="text-capitalize">{{ t('OCSERV_USER') }}: </span>
                    <span class="text-primary text-capitalize font-weight-bold">{{ username }}</span>
                </div>

                <div>
                    <div v-if="sessions?.length">
                        <!-- Global action -->
                        <div class="d-flex justify-end mb-2">
                            <v-btn
                                class="me-3"
                                size="x-small"
                                color="error"
                                variant="flat"
                                @click="emits('terminate', 'all')"
                            >
                                <v-icon start>mdi-power-off</v-icon>
                                {{ t('TERMINATE_ALL') }}
                            </v-btn>
                            <v-btn size="x-small" color="warning" variant="flat" @click="emits('disconnect', 'all')">
                                <v-icon start>mdi-lan-disconnect</v-icon>
                                {{ t('DISCONNECT_ALL') }}
                            </v-btn>
                        </div>

                        <!-- Sessions -->
                        <div v-for="session in sessions" :key="session.ID" class="pa-3 mb-2 rounded border">
                            <!-- Header -->
                            <div class="d-flex justify-space-between align-center mb-2">
                                <div class="text-body-2 font-weight-medium">
                                    {{ session.Device }}
                                </div>

                                <v-chip size="x-small" variant="tonal" color="primary">
                                    {{ session.IPv4 }}
                                </v-chip>
                            </div>

                            <!-- Stats -->
                            <div class="text-caption text-capitalize text-medium-emphasis">
                                {{ t('AVERAGE') }} RX: {{ session['Average RX'] }} | {{ t('AVERAGE') }} TX:
                                {{ session['Average TX'] }}
                            </div>

                            <!-- Timing -->
                            <div class="text-caption text-capitalize text-medium-emphasis">
                                {{ t('LAST_CONNECTED_AT') }}: {{ session['_Last connected at'] }}
                            </div>

                            <div class="text-caption text-capitalize text-medium-emphasis">
                                {{ t('STARTED_AT') }}: {{ session['Session started at'] }}
                            </div>

                            <!-- Actions -->
                            <div class="d-flex justify-end">
                                <v-btn
                                    size="x-small"
                                    color="error"
                                    variant="flat"
                                    @click="emits('terminate', 'session', session.ID)"
                                >
                                    <v-icon start>mdi-power-off</v-icon>
                                    {{ t('TERMINATE') }}
                                </v-btn>

                                <v-btn
                                    size="x-small"
                                    color="warning"
                                    variant="flat"
                                    class="ms-3"
                                    @click="emits('disconnect', 'session', session.ID)"
                                >
                                    <v-icon start>mdi-lan-disconnect</v-icon>
                                    {{ t('DISCONNECT') }}
                                </v-btn>
                            </div>
                        </div>
                    </div>
                </div>
            </v-card-text>
        </v-card>
    </v-dialog>
</template>

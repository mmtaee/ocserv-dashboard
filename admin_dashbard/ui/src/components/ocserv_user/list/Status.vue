<script setup lang="ts">
import type { ModelsOcservUser } from '@/api';
import { useI18n } from 'vue-i18n';

defineProps<{ item: ModelsOcservUser }>();

const { t } = useI18n();
</script>

<template>
    <div class="text-capitalize">
        <!-- Locked -->
        <span v-if="item.is_locked && !Boolean(item.deactivated_at)">
            <v-icon color="warning" start>mdi-lock</v-icon>
            <span class="text-warning text-capitalize">{{ t('LOCKED') }}</span>
        </span>

        <!-- Deactivated -->
        <span v-else-if="Boolean(item.deactivated_at)">
            <v-icon color="error" start>mdi-close-network-outline</v-icon>
            <span class="text-error text-capitalize">{{ t('DEACTIVATED') }}</span>
        </span>

        <!-- Online -->
        <span v-else-if="item.is_online">
            <v-icon color="success" start>mdi-lan-connect</v-icon>

            <span class="text-success text-capitalize">
                {{ t('ONLINE') }}
            </span>

            <v-tooltip location="top" max-width="420">
                <template #activator="{ props }">
                    <v-chip v-bind="props" size="x-small" color="success" variant="outlined" class="ms-2">
                        {{ item.online_sessions.length }} sessions
                    </v-chip>
                </template>

                <div v-if="item.online_sessions?.length">
                    <div class="text-subtitle-2 mb-2 font-weight-medium">Active Sessions</div>

                    <v-divider class="mb-2" />

                    <div v-for="session in item.online_sessions" :key="session.ID" class="mb-2">
                        <div class="text-caption">
                            <strong>{{ session.Device }}</strong> · {{ session.IPv4 }}
                        </div>

                        <div class="text-caption text-medium-emphasis">
                            RX: {{ session['Average RX'] }} | TX: {{ session['Average TX'] }}
                        </div>

                        <div class="text-caption text-medium-emphasis">
                            {{ t('LAST_CONNECTED_AT') }}: {{ session['_Last connected at'] }}
                        </div>

                        <div class="text-caption text-medium-emphasis">
                            {{ t('STARTED_AT') }}: {{ session['Session started at'] }}
                        </div>
                    </div>
                </div>

                <div v-else class="text-caption text-medium-emphasis">No active sessions</div>
            </v-tooltip>
        </span>

        <!-- Disconnected -->
        <span v-else-if="!item.is_online">
            <v-icon color="grey" start>mdi-lan-disconnect</v-icon>
            <span class="text-grey text-capitalize">{{ t('DISCONNECTED') }}</span>
        </span>
    </div>
</template>

<style scoped lang="scss"></style>

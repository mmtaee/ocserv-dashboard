<script setup lang="ts">
import UiParentCard from '@/components/shared/UiParentCard.vue';
import Restore from '@/components/system/Restore.vue';
import Backup from '@/components/system/Backup.vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
</script>

<template>
    <v-row>
        <v-col cols="12" md="12">
            <UiParentCard :title="`Ocserv ${t('BACKUP_AND_RESTORE')}`">
                <div class="text-muted ms-8 text-capitalize">
                    {{ t('BACKUP_AND_RESTORE_TITLE') }}
                </div>

                <v-row align="center" justify="center" no-gutters class="pa-5">
                    <v-col cols="12" md="6">
                        <Backup />
                    </v-col>

                    <v-col cols="12" md="6">
                        <Restore />
                    </v-col>
                </v-row>
            </UiParentCard>
        </v-col>
    </v-row>
</template>

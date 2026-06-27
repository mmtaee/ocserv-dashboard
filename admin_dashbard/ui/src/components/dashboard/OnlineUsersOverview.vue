<script lang="ts" setup>
import type { ModelsOnlineUserSession } from '@/api';
import { useI18n } from 'vue-i18n';

defineProps<{
    sessions: ModelsOnlineUserSession[];
}>();

const { t } = useI18n();

const groupReformat = (group: string | undefined) => {
    if (group === undefined) {
        return '';
    }

    if (group == '(none)') return 'defaults';

    return group;
};
</script>
<template>
    <v-card elevation="10" height="450px">
        <v-card-item class="pa-6">
            <v-card-title class="text-h5 pt-sm-2 pb-7 text-capitalize">
                {{ t('LAST_ONLINE_USERS_OVERVIEW') }}
            </v-card-title>
            <v-table v-if="sessions.length > 0" class="month-table">
                <thead>
                    <tr>
                        <th class="text-subtitle-1 font-weight-bold">
                            {{ t('USERNAME') }}
                        </th>
                        <th class="text-subtitle-1 font-weight-bold">
                            {{ t('GROUP') }}
                        </th>
                        <th class="text-subtitle-1 font-weight-bold">RX ({{ t('AVERAGE') }})</th>
                        <th class="text-subtitle-1 font-weight-bold">TX ({{ t('AVERAGE') }})</th>
                        <th class="text-subtitle-1 font-weight-bold">
                            {{ t('CONNECTED_AT') }}
                        </th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="item in sessions.slice(0, 5)" :key="item.Username" class="month-item">
                        <td>
                            <p class="text-15 font-weight-medium text-muted">
                                {{ item.Username }}
                            </p>
                        </td>
                        <td>
                            <p class="text-15 font-weight-medium text-muted">
                                {{ groupReformat(item.Groupname) }}
                            </p>
                        </td>
                        <td>
                            <div class="">
                                <div class="text-13 text-muted">
                                    {{ item['Average RX'] }}
                                </div>
                            </div>
                        </td>
                        <td>
                            <div class="">
                                <div class="text-13 text-muted">
                                    {{ item['Average TX'] }}
                                </div>
                            </div>
                        </td>
                        <td>
                            <h6 class="text-body-1 text-muted">
                                {{ item['_Last connected at'] }}
                            </h6>
                        </td>
                    </tr>
                </tbody>
            </v-table>
            <div v-else class="text-capitalize">{{ t('NO_ONLINE_USER_OVERVIEW_FOUND') }}!</div>
        </v-card-item>
    </v-card>
</template>

<script setup lang="ts">
import UiChildCard from '@/components/shared/UiChildCard.vue';
import { useI18n } from 'vue-i18n';
import { computed, ref } from 'vue';
import { SystemBackupApi } from '@/api';
import { getAuthorization } from '@/utils/request';

const { t } = useI18n();

const api = new SystemBackupApi();
api.backupOcservGroupsGet({
    ...getAuthorization()
});

api.backupOcservUsersGet({
    ...getAuthorization()
});

const downloadDialog = ref(false);
const downloadProgress = ref(0);
const currentDownload = ref('');

const startDownload = async (name: 'ocservGroups' | 'ocservUsers') => {
    currentDownload.value = name;
    downloadProgress.value = 0;
    downloadDialog.value = true;

    try {
        let res;

        if (name === 'ocservGroups') {
            res = await api.backupOcservGroupsGet(
                {
                    ...getAuthorization()
                },
                {
                    responseType: 'blob'
                }
            );
        } else {
            res = await api.backupOcservUsersGet(
                {
                    ...getAuthorization()
                },
                {
                    responseType: 'blob'
                }
            );
        }

        // Create blob from response data
        const blob = new Blob([res.data], { type: 'application/gzip' });

        // Get filename from header if available
        const contentDisposition = res.headers['content-disposition'];
        let filename = `${name}-backup.json.gz`.toLowerCase();
        if (contentDisposition) {
            const match = contentDisposition.match(/filename="?(.+)"?/);
            if (match && match[1]) filename = match[1];
        }

        // Trigger download
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);

        // Optional: simulate progress
        const interval = setInterval(() => {
            if (downloadProgress.value < 100) {
                downloadProgress.value += 10;
            } else {
                clearInterval(interval);
            }
        }, 100);
    } catch (err) {
        console.error(err);
        alert(`Failed to download ${name} backup`);
    } finally {
        downloadDialog.value = false;
        downloadProgress.value = 0;
        currentDownload.value = '';
    }
};

const progressColor = computed(() => {
    if (downloadProgress.value < 30) return 'red';
    if (downloadProgress.value < 70) return 'orange';
    return 'green';
});
</script>

<template>
    <UiChildCard variant="elevated" :height="570">
        <template #title-header>
            <div class="text-17 text-capitalize mb-3 px-1">
                {{ t('BACKUP_DATA_TITLE') }}
            </div>
            <hr style="color: #eeeeee" class="mx-1" />
        </template>

        <v-list lines="two">
            <div class="ms-md-1 mb-md-5 text-capitalize">
                {{ t('DOWNLOAD_BACKUP') }}
            </div>

            <v-list-item variant="elevated" class="mx-1 my-3">
                <template #title>
                    <span class="text-capitalize">
                        {{ t('DOWNLOAD_GROUPS_BACKUP') }}
                    </span>
                </template>

                <template #subtitle>
                    <div class="text-subtitle-2">
                        <span class="text-capitalize">
                            {{ t('DOWNLOAD_A_BACKUP_OF_ALL_GROUPS') }}
                        </span>
                        (.json.gz)
                    </div>
                </template>

                <template v-slot:prepend>
                    <v-icon size="40">mdi-router-network</v-icon>
                </template>

                <template v-slot:append>
                    <v-btn color="primary" variant="flat" size="small" @click="startDownload('ocservGroups')">
                        {{ t('DOWNLOAD') }}
                    </v-btn>
                </template>
            </v-list-item>

            <v-list-item
                subtitle="Download Users Backup"
                title="Download Users Backup title"
                variant="elevated"
                class="ma-1 my-2"
            >
                <template #title>
                    <span class="text-capitalize">
                        {{ t('DOWNLOAD_USERS_BACKUP') }}
                    </span>
                </template>

                <template #subtitle>
                    <div class="text-subtitle-2">
                        <span class="text-capitalize text-subtitle-2">
                            {{ t('DOWNLOAD_A_BACKUP_OF_ALL_USERS') }}
                        </span>
                        (.json.gz)
                    </div>
                </template>

                <template v-slot:prepend>
                    <v-icon size="40">mdi-account-network</v-icon>
                </template>

                <template v-slot:append>
                    <v-btn color="primary" variant="flat" size="small" @click="startDownload('ocservUsers')">
                        {{ t('DOWNLOAD') }}
                    </v-btn>
                </template>
            </v-list-item>
        </v-list>
    </UiChildCard>

    <!-- Download Dialog -->
    <v-dialog v-model="downloadDialog" persistent max-width="400px">
        <v-card>
            <v-card-title class="text-h6">
                {{ t('DOWNLOADING') }}
            </v-card-title>
            <v-card-text>
                <v-progress-linear v-model="downloadProgress" height="12" :color="progressColor" striped rounded />
                <div class="text-subtitle-2 mt-2">{{ downloadProgress }}%</div>
            </v-card-text>
        </v-card>
    </v-dialog>
</template>

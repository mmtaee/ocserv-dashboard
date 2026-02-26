<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import UiParentCard from '@/components/shared/UiParentCard.vue';
import UiChildCard from '@/components/shared/UiChildCard.vue';
import { ref } from 'vue';

const { t } = useI18n();
const restoreType = ref('groups');
const file = ref<File | null>(null);
const fileInput = ref<HTMLInputElement | null>(null);

function triggerFileSelect() {
    fileInput.value?.click();
}

function onFileSelected(event: Event) {
    const target = event.target as HTMLInputElement;
    const selectedFile = target.files?.[0];

    if (selectedFile && selectedFile.type !== 'application/json' && !selectedFile.name.endsWith('.json')) {
        if (fileInput.value) {
            fileInput.value.value = '';
        }
        return;
    }

    file.value = selectedFile || null;
}

function onDrop(e: any) {
    file.value = e.dataTransfer.files[0];
}

function clearFile() {
    file.value = null;

    // optional: also reset hidden input value
    if (fileInput.value) {
        fileInput.value.value = '';
    }
}
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
                                        <v-btn color="primary" variant="flat" size="small">
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
                                        <v-btn color="primary" variant="flat" size="small">
                                            {{ t('DOWNLOAD') }}
                                        </v-btn>
                                    </template>
                                </v-list-item>
                            </v-list>
                        </UiChildCard>
                    </v-col>

                    <v-col cols="12" md="6">
                        <UiChildCard variant="flat" :height="570">
                            <template #title-header>
                                <div class="text-17 text-capitalize mb-3 px-1">
                                    {{ t('RESTORE_DATA_TITLE') }}
                                </div>
                                <hr style="color: #eeeeee" class="mx-1" />
                            </template>
                            <div class="ms-md-1 mb-md-5 mt-md-2 text-capitalize">
                                {{ t('UPLOAD_BACKUP_FILE') }}
                            </div>

                            <!-- Drag & Drop Area -->
                            <v-sheet
                                class="pa-8 text-center border-dashed rounded-lg"
                                elevation="0"
                                style="border: 2px dashed #dcdcdc"
                                @dragover.prevent
                                @drop.prevent="onDrop"
                            >
                                <v-icon size="60" class="mb-4" color="grey"> mdi-cloud-upload-outline </v-icon>

                                <div class="mb-5 text-medium-emphasis" v-if="!file">
                                    {{ t('DRAG_DROP_BACKUP') }}
                                </div>
                                <div class="text-medium-emphasis" v-else style="color: #888888 !important">
                                    {{ file.name }}
                                    <v-btn icon variant="text" size="small" @click="clearFile">
                                        <v-icon size="18">mdi-close-circle-outline</v-icon>
                                    </v-btn>
                                </div>

                                <input
                                    ref="fileInput"
                                    type="file"
                                    class="d-none"
                                    @change="onFileSelected"
                                    accept=".json"
                                />

                                <!-- Browse Button -->
                                <v-btn color="primary" class="mt-3" @click="triggerFileSelect">
                                    {{ t('BROWSE_FILE') }}
                                </v-btn>
                            </v-sheet>

                            <!-- Restore Type -->
                            <div class="mt-6">
                                <div class="mb-2 text-medium-emphasis">{{ t('RESTORE_TYPE') }}:</div>

                                <v-radio-group v-model="restoreType" inline density="comfortable">
                                    <v-radio :label="t('RESTORE_GROUPS')" value="groups" />
                                    <v-radio :label="t('RESTORE_USERS')" value="users" />
                                </v-radio-group>
                            </div>

                            <!-- Upload & Restore Button -->
                            <div class="mt-3 text-center">
                                <v-btn color="primary" size="large" :disabled="!file">
                                    {{ t('UPLOAD_AND_RESTORE') }}
                                </v-btn>
                            </div>
                        </UiChildCard>
                    </v-col>
                </v-row>
            </UiParentCard>
        </v-col>
    </v-row>
</template>

<style scoped></style>

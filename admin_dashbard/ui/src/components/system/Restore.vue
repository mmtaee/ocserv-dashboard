<script setup lang="ts">
import UiChildCard from '@/components/shared/UiChildCard.vue';
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { type BackupRestoreResponse, SystemRestoreApi } from '@/api';
import { getAuthorization } from '@/utils/request';
import RestoreResultDialog from '@/components/system/RestoreResultDialog.vue';

const { t } = useI18n();

const restoreType = ref<'users' | 'groups'>('groups');
const file = ref<File | null>(null);
const fileInput = ref<HTMLInputElement | null>(null);
const showResults = ref(false);
const result = ref<BackupRestoreResponse>({});

function triggerFileSelect() {
    fileInput.value?.click();
}

const detectRestoreType = (file: File): 'users' | 'groups' => {
    const name = file.name.toLowerCase();

    if (/users/.test(name)) return 'users';
    if (/groups/.test(name)) return 'groups';

    return restoreType.value;
};

function onFileSelected(event: Event) {
    const target = event.target as HTMLInputElement;
    const selectedFile = target.files?.[0];

    if (!selectedFile) return;

    const isJson = selectedFile.type === 'application/json' || selectedFile.name.endsWith('.json');

    const isGzJson = selectedFile.name.endsWith('.json.gz');

    if (!isJson && !isGzJson) {
        if (fileInput.value) fileInput.value.value = '';
        return;
    }

    file.value = selectedFile || null;

    if (file.value.name) {
        restoreType.value = detectRestoreType(selectedFile);
    }
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

const restore = () => {
    if (!file.value) {
        return;
    }

    const api = new SystemRestoreApi();

    if (restoreType.value == 'users') {
        api.backupOcservUsersPost({
            ...getAuthorization(),
            file: file.value
        }).then((res) => {
            result.value = res.data;
            showResults.value = true;
        });
    } else {
        api.backupOcservGroupsPost({
            ...getAuthorization(),
            file: file.value
        }).then((res) => {
            result.value = res.data;
            showResults.value = true;
        });
    }
};
</script>

<template>
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

            <input ref="fileInput" type="file" class="d-none" @change="onFileSelected" accept=".json,.gz" />

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
            <v-btn color="primary" size="large" :disabled="!file" @click="restore">
                {{ t('UPLOAD_AND_RESTORE') }}
            </v-btn>
        </div>
    </UiChildCard>

    <RestoreResultDialog
        :show="showResults"
        :title="restoreType"
        @close="showResults = false"
        :inserted="result?.inserted || []"
        :existing="result?.existing || []"
    />
</template>

<script setup lang="ts">
import UiParentCard from '@/components/shared/UiParentCard.vue';
import { useI18n } from 'vue-i18n';
import UiChildCard from '@/components/shared/UiChildCard.vue';
import { OcservOcpasswdApi, type OcservUserSyncOcpasswdRequest, type UserOcpasswd } from '@/api';
import { getAuthorization } from '@/utils/request';
import { computed, onMounted, reactive, ref } from 'vue';
import Pagination from '@/components/shared/Pagination.vue';
import type { Meta } from '@/types/metaTypes/MetaType';
import SaveDBDialog from '@/components/ocserv_user/SaveDBDialog.vue';
import SyncDBResultDialog from '@/components/ocserv_user/SyncDBResultDialog.vue';

const { t } = useI18n();
const meta = reactive<Meta>({
    page: 1,
    size: 10,
    sort: 'ASC',
    total_records: 0
});
const loading = ref(false);
const users = reactive<UserOcpasswd[]>([]);
const selectedUsers = ref<string[]>([]);
const selectAllUser = ref(false);
const showDBDialog = ref(false);
const syncedUsernames = ref<string[]>(['test', 'test2']);
const showSyncResultDialog = ref(false);

const api = new OcservOcpasswdApi();

const sync = () => {
    api.ocservUsersOcpasswdGet({
        ...getAuthorization(),
        ...meta
    })
        .then((res) => {
            users.splice(0, users.length, ...(res.data.result ?? []));
            Object.assign(meta, res.data.meta);
        })
        .finally(() => {
            selectedUsers.value = [];
            selectAllUser.value = false;
        });
};

function updateMeta(newMeta: Meta) {
    Object.assign(meta, newMeta);
    sync();
}

const toggleSelectAll = (value: boolean) => {
    if (value) {
        selectedUsers.value = users.map((u) => u.username || '');
    } else {
        selectedUsers.value = [];
    }
};

const saveToDB = (config: OcservUserSyncOcpasswdRequest) => {
    loading.value = true;

    config.users = users.filter((u) => u.username && selectedUsers.value.includes(u.username));

    api.ocservUsersOcpasswdSyncPost({
        ...getAuthorization(),
        request: config
    })
        .then((res) => {
            syncedUsernames.value = res.data;
            showDBDialog.value = false;
            selectedUsers.value = [];
            showSyncResultDialog.value = true;
            sync();
        })
        .finally(() => {
            loading.value = false;
        });
};

onMounted(() => {
    sync();
});
</script>

<template>
    <v-row>
        <v-col cols="12" md="12">
            <UiParentCard :title="t('SYNC_PAGE_TITLE')">
                <template #action>
                    <v-btn class="me-lg-5" color="primary" size="small" variant="flat" @click="sync">
                        {{ t('RELOAD') }}
                    </v-btn>
                </template>

                <div class="mx-10 text-justify text-muted text-subtitle-1">
                    {{ t('OCSERV_USER_SYNC_HELP_1') }}
                    {{ t('OCSERV_USER_SYNC_HELP_2') }}.
                </div>
                <div class="mx-10 mb-5 text-justify text-muted text-subtitle-1 mt-2">
                    <v-icon color="info" size="small" class="me-1 mb-1">mdi-information-outline</v-icon>
                    <span class="text-capitalize text-info">{{ t('NOTE') }}</span
                    >: {{ t('OCSERV_USER_SYNC_HELP_3') }}.
                </div>

                <UiChildCard>
                    <v-progress-linear :active="loading" indeterminate></v-progress-linear>

                    <div v-if="!loading && users.length > 0">
                        <v-row align="center" justify="space-between" class="my-3 mx-lg-15">
                            <v-col cols="auto" class="ma-0 pa-0 text-capitalize">
                                {{ t('SELECTED_USERS') }}: {{ selectedUsers.length }}
                            </v-col>
                            <v-col cols="auto" class="ma-0 pa-0">
                                <v-btn
                                    :style="{ visibility: !selectedUsers.length ? 'hidden' : 'visible' }"
                                    class="me-lg-5"
                                    color="lightprimary"
                                    size="small"
                                    variant="flat"
                                    @click="showDBDialog = true"
                                >
                                    {{ t('ASSIGN_CONFIGURATION') }}
                                </v-btn>
                            </v-col>
                        </v-row>

                        <v-table class="px-md-15">
                            <thead>
                                <tr class="text-capitalize bg-lightprimary">
                                    <th class="text-left">
                                        <v-row align="center" justify="start">
                                            <v-col cols="auto" class="ma-0 pa-0">
                                                <v-checkbox
                                                    v-model="selectAllUser"
                                                    class="text-capitalize text-subtitle-2"
                                                    color="primary"
                                                    hide-details
                                                    @update:model-value="
                                                        (val: unknown) => toggleSelectAll(val as boolean)
                                                    "
                                                />
                                            </v-col>
                                            <v-col cols="auto" class="ma-0 pa-0">
                                                {{ t('USERNAME') }}
                                            </v-col>
                                        </v-row>
                                    </th>
                                    <th class="text-left">{{ t('GROUP') }}</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr v-for="item in users" :key="item.username">
                                    <td>
                                        <v-row align="center" justify="start">
                                            <v-col cols="auto" class="ma-0 pa-0">
                                                <v-checkbox
                                                    class="text-capitalize text-subtitle-2"
                                                    :value="item.username"
                                                    v-model="selectedUsers"
                                                    color="primary"
                                                    hide-details
                                                />
                                            </v-col>
                                            <v-col cols="auto" class="ma-0 pa-0">
                                                {{ item.username }}
                                            </v-col>
                                        </v-row>
                                    </td>
                                    <td>{{ item.group }}</td>
                                </tr>
                            </tbody>
                        </v-table>

                        <Pagination :meta="meta" @update="updateMeta" />
                    </div>

                    <div v-else class="ms-md-5 mb-md-5 text-capitalize">
                        {{ t('NO_USER_FOUND_TABLE') }}
                    </div>
                </UiChildCard>
            </UiParentCard>
        </v-col>
    </v-row>

    <SaveDBDialog :show="showDBDialog" @saveToDB="saveToDB" @close="showDBDialog = false" :loading="loading" />

    <SyncDBResultDialog :show="showSyncResultDialog" :usernames="syncedUsernames" @close="showSyncResultDialog = false" />
</template>

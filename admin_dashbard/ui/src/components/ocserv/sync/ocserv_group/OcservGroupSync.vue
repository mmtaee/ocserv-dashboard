<script setup lang="ts">
import UiParentCard from '@/components/shared/UiParentCard.vue';
import UiChildCard from '@/components/shared/UiChildCard.vue';
import { useI18n } from 'vue-i18n';
import { onMounted, reactive, ref } from 'vue';
import { type GroupUnsyncedGroup, OcservUnsyncedGroupApi } from '@/api';
import { getAuthorization } from '@/utils/request';
import OcservGroupConfigDialog from '@/components/ocserv_group/OcservGroupConfigDialog.vue';
import OcservGroupSyncDBResultDialog from '@/components/ocserv/sync/ocserv_group/OcservGroupSyncDBResultDialog.vue';

const { t } = useI18n();

const loading = ref(false);
const groups = reactive<GroupUnsyncedGroup[]>([]);
const selectedGroups = ref<string[]>([]);
const selectAllGroup = ref(false);

const selectedGroup = ref<GroupUnsyncedGroup>({ config: {}, name: '', path: '' });
const configDialog = ref(false);

const syncedGroups = ref<string[]>([]);
const showDBDialog = ref(false);
const showSyncResultDialog = ref(false);

const api = new OcservUnsyncedGroupApi();
const sync = () => {
    api.ocservGroupsUnsyncedGet({
        ...getAuthorization()
    }).then((res) => {
        groups.splice(0, groups.length, ...(res.data ?? []));
    });
};

const saveToDB = () => {
    loading.value = true;

    const selectedGroupsFilter = groups.filter((g) => g.name && selectedGroups.value.includes(g.name));

    api.ocservGroupsSyncPost({
        ...getAuthorization(),
        request: {
            groups: selectedGroupsFilter
        }
    })
        .then((res) => {
            syncedGroups.value = res.data;
            showDBDialog.value = false;
            selectedGroups.value = [];
            showSyncResultDialog.value = true;
            sync();
        })
        .finally(() => {
            loading.value = false;
        });
};

const toggleSelectAll = (value: boolean) => {
    if (value) {
        selectedGroups.value = groups.map((u) => u.name || '');
    } else {
        selectedGroups.value = [];
    }
};

const showConfig = (group: GroupUnsyncedGroup) => {
    Object.assign(selectedGroup.value, group);
    configDialog.value = true;
};

onMounted(() => {
    sync();
});
</script>

<template>
    <v-row>
        <v-col cols="12" md="12">
            <UiParentCard :title="t('GROUP_SYNC_PAGE_TITLE')" variant="text" :minHeight="650">
                <template #action>
                    <v-btn class="me-lg-5" color="primary" size="small" variant="flat" @click="sync">
                        {{ t('RELOAD') }}
                    </v-btn>
                </template>

                <UiChildCard>
                    <div class="mx-10 text-justify text-muted text-subtitle-1">
                        {{ t('OCSERV_GROUP_SYNC_HELP_1') }}
                    </div>
                    <div class="mx-10 mb-5 text-justify text-muted text-subtitle-1 mt-2">
                        <v-icon color="info" size="small" class="me-1 mb-1">mdi-information-outline</v-icon>
                        <span class="text-capitalize text-info">{{ t('NOTE') }}</span
                        >: {{ t('OCSERV_GROUP_SYNC_HELP_2') }}.
                    </div>

                    <v-progress-linear :active="loading" indeterminate></v-progress-linear>

                    <div v-if="!loading && groups.length > 0">
                        <v-row align="center" justify="space-between" class="my-3 mx-lg-15">
                            <v-col cols="auto" class="ma-0 pa-0 text-capitalize">
                                {{ t('SELECTED_GROUPS') }}: {{ selectedGroups.length }}
                            </v-col>
                            <v-col cols="auto" class="ma-0 pa-0">
                                <v-btn
                                    :style="{ visibility: !selectedGroups.length ? 'hidden' : 'visible' }"
                                    class="me-lg-5"
                                    color="lightprimary"
                                    size="small"
                                    variant="flat"
                                    @click="saveToDB"
                                    :loading="loading"
                                >
                                    {{ t('SYNC') }}
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
                                                    v-model="selectAllGroup"
                                                    class="text-capitalize text-subtitle-2"
                                                    color="primary"
                                                    hide-details
                                                    @update:model-value="
                                                        (val: unknown) => toggleSelectAll(val as boolean)
                                                    "
                                                />
                                            </v-col>
                                            <v-col cols="auto" class="ma-0 pa-0">
                                                {{ t('NAME') }}
                                            </v-col>
                                        </v-row>
                                    </th>
                                    <th class="text-left">{{ t('PATH') }}</th>
                                    <th class="text-left">{{ t('CONFIG') }}</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr v-for="item in groups" :key="item.name">
                                    <td width="30%">
                                        <v-row align="center" justify="start">
                                            <v-col cols="auto" class="ma-0 pa-0">
                                                <v-checkbox
                                                    class="text-capitalize text-subtitle-2"
                                                    :value="item.name"
                                                    v-model="selectedGroups"
                                                    color="primary"
                                                    hide-details
                                                />
                                            </v-col>
                                            <v-col cols="auto" class="ma-0 pa-0">
                                                {{ item.name }}
                                            </v-col>
                                        </v-row>
                                    </td>
                                    <td width="50%">{{ item.path }}</td>
                                    <td class="text-start" width="10%">
                                        <v-icon color="info" size="small" end class="mx-4" @click="showConfig(item)">
                                            mdi-eye
                                        </v-icon>
                                    </td>
                                </tr>
                            </tbody>
                        </v-table>
                    </div>

                    <div v-if="loading || groups.length == 0" class="ms-md-5 mb-md-5 text-capitalize">
                        {{ t('NO_GROUP_FOUND_TABLE') }}
                    </div>
                </UiChildCard>
            </UiParentCard>
        </v-col>
    </v-row>

    <OcservGroupConfigDialog :show="configDialog" @close="configDialog = false" :group="selectedGroup" />

    <OcservGroupSyncDBResultDialog
        :show="showSyncResultDialog"
        :groupNames="syncedGroups"
        @close="showSyncResultDialog = false"
    />
</template>

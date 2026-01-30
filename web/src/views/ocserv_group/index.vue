<script lang="ts" setup>
import UiParentCard from '@/components/shared/UiParentCard.vue';
import { useI18n } from 'vue-i18n';
import { router } from '@/router';
import { type ModelsOcservGroup, OcservGroupsApi } from '@/api';
import { onMounted, reactive, ref } from 'vue';
import { getAuthorization } from '@/utils/request';
import DeleteDialog from '@/components/ocserv_group/DeleteDialog.vue';
import type { Meta } from '@/types/metaTypes/MetaType';
import { useProfileStore } from '@/stores/profile';
import Pagination from '@/components/shared/Pagination.vue';

const { t } = useI18n();
const api = new OcservGroupsApi();
const meta = reactive<Meta>({
    page: 1,
    size: 25,
    sort: 'ASC',
    total_records: 0
});
const groups = reactive<ModelsOcservGroup[]>([]);
const loading = ref(false);
const deleteDialog = ref(false);
const deleteGroupName = ref('');
const deleteGroupID = ref(0);

const profileStore = useProfileStore();

// TODO: should change to role base
const isAdmin = ref(profileStore.isAdmin);

const getGroups = () => {
    loading.value = true;
    api.ocservGroupsGet({
        ...getAuthorization(),
        ...meta
    })
        .then((res) => {
            groups.splice(0, groups.length, ...(res.data.result ?? []));
            Object.assign(meta, res.data.meta);
        })
        .finally(() => {
            loading.value = false;
        });
};

const detailGroup = async (id: number) => {
    await router.push({ name: 'Ocserv Group Detail', params: { id: id } });
};

const editGroup = async (id: number) => {
    await router.push({ name: 'Ocserv Group Update', params: { id: id } });
};

const deleteGroupHandler = (id: number, name: string) => {
    deleteGroupID.value = id;
    deleteGroupName.value = name;
    deleteDialog.value = true;
};

const cancelDeleteGroup = () => {
    deleteGroupID.value = 0;
    deleteGroupName.value = '';
    deleteDialog.value = false;
};

const deleteGroup = () => {
    api.ocservGroupsIdDelete({
        ...getAuthorization(),
        id: deleteGroupID.value
    })
        .then((_) => {
            getGroups();
        })
        .finally(() => {
            cancelDeleteGroup();
        });
};

const updateMeta = (newMeta: Meta) => {
    Object.assign(meta, newMeta);
    getGroups();
};
</script>

<template>
    <v-row>
        <v-col cols="12" md="12">
            <UiParentCard :title="t('OCSERV_GROUPS')">
                <template #action>
                    <v-btn
                        class="me-lg-5"
                        color="grey"
                        size="small"
                        variant="flat"
                        @click="router.push({ name: 'Ocserv Group Create' })"
                    >
                        {{ t('CREATE') }}
                    </v-btn>
                </template>

                <v-progress-linear :active="loading" indeterminate></v-progress-linear>

                <v-table v-if="!loading && groups.length > 0" class="px-md-15 my-md-10">
                    <thead>
                        <tr class="text-capitalize bg-lightprimary">
                            <th class="text-left">ID</th>
                            <th class="text-left">{{ t('NAME') }}</th>
                            <!--                            // TODO: should change to role base-->
                            <th class="text-left" v-if="isAdmin">{{ t('OWNER') }}</th>
                            <th class="text-left">{{ t('ACTION') }}</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="item in groups" :key="item.name">
                            <td>{{ item.id }}</td>
                            <td>{{ item.name }}</td>
                            <!--                            // TODO: should change to role base-->
                            <td v-if="isAdmin">{{ item.owner }}</td>
                            <td>
                                <v-menu>
                                    <template v-slot:activator="{ props }">
                                        <v-icon start v-bind="props"> mdi-dots-vertical</v-icon>
                                    </template>

                                    <v-list>
                                        <v-list-item @click="detailGroup(item?.id || 0)">
                                            <v-list-item-title class="text-primary text-capitalize me-5">
                                                {{ t('DETAIL') }}
                                            </v-list-item-title>
                                            <template v-slot:prepend>
                                                <v-icon class="ms-2" color="primary">mdi-information-outline</v-icon>
                                            </template>
                                        </v-list-item>

                                        <v-list-item @click="editGroup(item?.id || 0)">
                                            <v-list-item-title class="text-info text-capitalize me-5">
                                                {{ t('UPDATE') }}
                                            </v-list-item-title>
                                            <template v-slot:prepend>
                                                <v-icon class="ms-2" color="info">mdi-pencil</v-icon>
                                            </template>
                                        </v-list-item>

                                        <v-list-item @click="deleteGroupHandler(item?.id || 0, item.name)">
                                            <v-list-item-title class="text-error text-capitalize me-5">
                                                {{ t('DELETE') }}
                                            </v-list-item-title>
                                            <template v-slot:prepend>
                                                <v-icon class="ms-2" color="error">mdi-delete</v-icon>
                                            </template>
                                        </v-list-item>
                                    </v-list>
                                </v-menu>
                            </td>
                        </tr>
                    </tbody>
                </v-table>

                <div v-if="loading || groups.length == 0" class="ms-md-5 mb-md-5 text-capitalize">
                    {{ t('NO_GROUP_FOUND_TABLE') }}
                </div>

                <Pagination @update="updateMeta" :totalRecords="meta.total_records" />
            </UiParentCard>
        </v-col>
    </v-row>

    <DeleteDialog :name="deleteGroupName" :show="deleteDialog" @close="cancelDeleteGroup" @deleteGroup="deleteGroup" />
</template>

<style scoped>
tbody tr:nth-child(even) td {
    background-color: #f5f5f5;
}

@media (min-width: 992px) {
    tbody tr:nth-child(even) td {
        background-color: #f5f5f5;
    }
    tbody tr:nth-child(even) td:first-child {
        border-radius: 8px 0 0 8px;
    }
    tbody tr:nth-child(even) td:last-child {
        border-radius: 0 8px 8px 0;
    }
}
</style>

<script lang="ts" setup>
import { router } from '@/router';
import { type ModelsUser, SystemUsersApi } from '@/api';
import UiParentCard from '@/components/shared/UiParentCard.vue';
import { useI18n } from 'vue-i18n';
import { onMounted, reactive, ref } from 'vue';
import { getAuthorization } from '@/utils/request';
import { formatDateTimeWithRelative } from '@/utils/convertors';
import DeleteDialog from '@/components/user/DeleteDialog.vue';
import ChangePasswordDialog from '@/components/user/ChangePasswordDialog.vue';
import Pagination from '@/components/shared/Pagination.vue';
import type { Meta } from '@/types/metaTypes/MetaType';
import { useSnackbarStore } from '@/stores/snackbar';

const { t } = useI18n();
const loading = ref(false);

const changePasswordDialog = ref(false);

const deleteDialog = ref(false);
const staffName = ref('');
const staffUID = ref('');

const staffs = reactive<ModelsUser[]>([]);
const meta = reactive<Meta>({
    page: 1,
    size: 25,
    sort: 'ASC',
    total_records: 0
});

const snackbar = useSnackbarStore();

const api = new SystemUsersApi();

const getStaffs = () => {
    loading.value = true;

    api.systemUsersGet({
        ...getAuthorization(),
        ...meta
    })
        .then((res) => {
            staffs.splice(0, staffs.length, ...(res.data.result ?? []));
            Object.assign(meta, res.data.meta);
        })
        .finally(() => {
            loading.value = false;
        });
};

const activities = async (uid: string, username: string) => {
    await router.push({ name: 'Staff Activities', query: { uid: uid, username: username } });
};

const deleteStaffHandler = (uid: string, username: string) => {
    staffUID.value = uid;
    staffName.value = username;
    deleteDialog.value = true;
};

const cancelDeleteStaff = () => {
    staffUID.value = '';
    staffName.value = '';
    deleteDialog.value = false;
};

const deleteStaff = () => {
    api.systemUsersUidDelete({
        ...getAuthorization(),
        uid: staffUID.value
    })
        .then(() => {
            snackbar.show({
                id: 1,
                message: t('USER_DELETED_SUCCESS_SNACK'),
                color: 'success',
                timeout: 4000
            });
            getStaffs();
        })
        .finally(() => {
            cancelDeleteStaff();
        });
};

const changePasswordHandler = (uid: string, username: string) => {
    staffUID.value = uid;
    staffName.value = username;
    changePasswordDialog.value = true;
};

const cancelChangePassword = () => {
    staffUID.value = '';
    staffName.value = '';
    changePasswordDialog.value = false;
};

const changePassword = (password: string) => {
    api.systemUsersUidPasswordPost({
        ...getAuthorization(),
        uid: staffUID.value,
        request: {
            password: password
        }
    })
        .then(() => {
            snackbar.show({
                id: 1,
                message: t('PASSWORD_CHANGE_SUCCESS_SNACK'),
                color: 'success',
                timeout: 4000
            });
        })
        .finally(() => {
            cancelChangePassword();
        });
};

const updateMeta = (newMeta: Meta) => {
    Object.assign(meta, newMeta);
    getStaffs();
};
</script>

<template>
    <v-row>
        <v-col cols="12" md="12">
            <UiParentCard :title="t('STAFFS')">
                <template #action>
                    <v-btn
                        class="me-lg-5"
                        color="grey"
                        size="small"
                        variant="flat"
                        @click="router.push({ name: 'Staff Create' })"
                    >
                        {{ t('CREATE') }}
                    </v-btn>
                </template>

                <v-progress-linear :active="loading" indeterminate></v-progress-linear>

                <div v-if="!loading && staffs.length > 0">
                    <v-table class="px-md-15">
                        <thead>
                            <tr class="text-capitalize bg-lightprimary">
                                <th class="text-left">UID</th>
                                <th class="text-left">{{ t('USERNAME') }}</th>
                                <th class="text-left">{{ t('CREATED_AT') }}</th>
                                <th class="text-left">{{ t('LAST_LOGIN') }}</th>
                                <th class="text-left">{{ t('ACTION') }}</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="item in staffs" :key="item.username">
                                <td>{{ item.uid }}</td>
                                <td>{{ item.username }}</td>
                                <td>{{ formatDateTimeWithRelative(item.created_at, t('NOT_AVAILABLE')) }}</td>
                                <td>{{ formatDateTimeWithRelative(item.last_login, t('NOT_AVAILABLE')) }}</td>
                                <td>
                                    <v-menu>
                                        <template v-slot:activator="{ props }">
                                            <v-icon start v-bind="props"> mdi-dots-vertical</v-icon>
                                        </template>

                                        <v-list>
                                            <v-list-item @click="activities(item?.uid, item.username)">
                                                <v-list-item-title class="text-grey text-capitalize me-5">
                                                    {{ t('ACTIVITIES') }}
                                                </v-list-item-title>
                                                <template v-slot:prepend>
                                                    <v-icon class="ms-2" color="grey">mdi-history</v-icon>
                                                </template>
                                            </v-list-item>

                                            <v-list-item @click="changePasswordHandler(item?.uid, item.username)">
                                                <v-list-item-title class="text-warning text-capitalize me-5">
                                                    {{ t('CHANGE_PASSWORD') }}
                                                </v-list-item-title>
                                                <template v-slot:prepend>
                                                    <v-icon class="ms-2" color="warning">mdi-key</v-icon>
                                                </template>
                                            </v-list-item>

                                            <v-list-item @click="deleteStaffHandler(item?.uid, item.username)">
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
                </div>

                <div v-if="loading || staffs.length == 0" class="ms-md-5 mb-md-5 text-capitalize">
                    {{ t('NO_STAFF_FOUND_TABLE') }}
                </div>

                <Pagination @update="updateMeta" :totalRecords="meta.total_records" />
            </UiParentCard>
        </v-col>
    </v-row>

    <ChangePasswordDialog
        :show="changePasswordDialog"
        :username="staffName"
        @changePassword="changePassword"
        @close="cancelChangePassword"
    />

    <DeleteDialog :show="deleteDialog" :username="staffName" @close="cancelDeleteStaff" @deleteStaff="deleteStaff" />
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

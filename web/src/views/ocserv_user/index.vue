<script lang="ts" setup>
import { router } from '@/router';
import UiParentCard from '@/components/shared/UiParentCard.vue';
import { useI18n } from 'vue-i18n';
import { onMounted, reactive, ref } from 'vue';
import { type ModelsOcservUser, ModelsOcservUserTrafficTypeEnum, OcservUsersApi } from '@/api';
import { getAuthorization } from '@/utils/request';
import { bytesToGB, formatDate, trafficTypesTransformer } from '@/utils/convertors';
import DeleteDialog from '@/components/ocserv_user/DeleteDialog.vue';
import Pagination from '@/components/shared/Pagination.vue';
import type { Meta } from '@/types/metaTypes/MetaType';
import { useSnackbarStore } from '@/stores/snackbar';
import { useProfileStore } from '@/stores/profile';

const { t } = useI18n();
const loading = ref(false);
const api = new OcservUsersApi();
const meta = reactive<Meta>({
    page: 1,
    size: 5,
    sort: 'ASC',
    total_records: 0
});
const deleteDialog = ref(false);
const deleteUserName = ref('');
const deleteUserUID = ref('');

const users = reactive<ModelsOcservUser[]>([]);
const snackbar = useSnackbarStore();

const profileStore = useProfileStore();
const isAdmin = ref(profileStore.isAdmin);

const getUsers = () => {
    loading.value = true;
    api.ocservUsersGet({
        ...getAuthorization(),
        ...meta
    })
        .then((res) => {
            users.splice(0, users.length, ...(res.data.result ?? []));
            Object.assign(meta, res.data.meta);
        })
        .finally(() => {
            loading.value = false;
        });
};

const detailUser = async (uid: string) => {
    await router.push({ name: 'Ocserv User Detail', params: { uid: uid } });
};

const editUser = async (uid: string) => {
    await router.push({ name: 'Ocserv User Update', params: { uid: uid } });
};

const disconnect = (username: string) => {
    api.ocservUsersUsernameDisconnectPost({
        ...getAuthorization(),
        username: username
    })
        .then(() => {
            let index = users.findIndex((i) => (i.username = username));
            if (index > -1) {
                users[index].is_online = false;
            }
        })
        .finally(() => {
            snackbar.show({
                id: 1,
                message: t('USER_DISCONNECTED_SUCCESS_SNACK'),
                color: 'success',
                timeout: 4000
            });
        });
};

const lock = (uid: string) => {
    api.ocservUsersUidLockPost({
        ...getAuthorization(),
        uid: uid
    })
        .then(() => {
            let index = users.findIndex((i) => (i.uid = uid));
            if (index > -1) {
                users[index].is_locked = true;
            }
        })
        .finally(() => {
            snackbar.show({
                id: 1,
                message: t('USER_LOCKED_SUCCESSFULLY_SNACK'),
                color: 'success',
                timeout: 4000
            });
        });
};

const unlock = (uid: string) => {
    api.ocservUsersUidUnlockPost({
        ...getAuthorization(),
        uid: uid
    })
        .then(() => {
            let index = users.findIndex((i) => (i.uid = uid));
            if (index > -1) {
                users[index].is_locked = false;
            }
        })
        .finally(() => {
            snackbar.show({
                id: 1,
                message: t('USER_UNLOCKED_SUCCESSFULLY_SNACK'),
                color: 'success',
                timeout: 4000
            });
        });
};

const statistics = async (uid: string, username: string) => {
    await router.push({ name: 'Ocserv User Statistics', params: { uid: uid }, query: { username: username } });
};

const deleteUserHandler = (uid: string, name: string) => {
    deleteUserUID.value = uid;
    deleteUserName.value = name;
    deleteDialog.value = true;
};

const cancelDeleteUser = () => {
    deleteUserUID.value = '';
    deleteUserName.value = '';
    deleteDialog.value = false;
};

const deleteUser = () => {
    api.ocservUsersUidDelete({
        ...getAuthorization(),
        uid: deleteUserUID.value
    })
        .then((_) => {
            getUsers();
        })
        .finally(() => {
            cancelDeleteUser();
        });
};

function updateMeta(newMeta: Meta) {
    Object.assign(meta, newMeta);
    getUsers();
}

onMounted(() => {
    getUsers();
});
</script>

<template>
    <v-row>
        <v-col cols="12" md="12">
            <UiParentCard :title="t('OCSERV_USERS')">
                <template #action>
                    <v-btn
                        class="me-lg-5"
                        color="grey"
                        size="small"
                        variant="flat"
                        @click="router.push({ name: 'Ocserv User Create' })"
                    >
                        {{ t('CREATE') }}
                    </v-btn>
                </template>

                <v-progress-linear :active="loading" indeterminate></v-progress-linear>

                <div v-if="!loading && users.length > 0">
                    <v-table class="px-md-15">
                        <thead>
                            <tr class="text-capitalize bg-lightprimary">
                                <th class="text-left">{{ t('USERNAME') }}</th>
                                <th class="text-left" v-if="isAdmin">{{ t('OWNER') }}</th>
                                <th class="text-left">{{ t('GROUP') }}</th>
                                <th class="text-left">{{ t('TRAFFIC') }}</th>
                                <th class="text-left">{{ t('BANDWIDTHS') }}</th>
                                <th class="text-left">{{ t('DATES') }}</th>
                                <th class="text-left">{{ t('STATUS') }}</th>
                                <th class="text-left">{{ t('ACTION') }}</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="item in users" :key="item.username">
                                <td>{{ item.username }}</td>
                                <td v-if="isAdmin">{{ item.owner || '' }}</td>
                                <td>{{ item.group }}</td>
                                <td class="text-capitalize">
                                    <div>
                                        {{ t('TRAFFIC_TYPE') }}:<br />
                                        <span class="text-info text-capitalize">
                                            {{ trafficTypesTransformer(item.traffic_type) }}
                                        </span>
                                    </div>
                                    <div>
                                        {{ t('TRAFFIC_SIZE') }}:<br />
                                        <span
                                            v-if="item.traffic_type != ModelsOcservUserTrafficTypeEnum.FREE"
                                            class="text-info text-capitalize"
                                        >
                                            {{ item.traffic_size }} GB
                                        </span>

                                        <span v-else class="text-info text-capitalize">
                                            {{ t('FREE') }}
                                        </span>
                                    </div>
                                </td>
                                <td style="cursor: pointer">
                                    <div>
                                        RX:
                                        <span
                                            v-if="item.traffic_type != ModelsOcservUserTrafficTypeEnum.FREE"
                                            class="text-muted text-subtitle-2"
                                        >
                                            ({{ t('CURRENT') }})
                                        </span>
                                        <br />
                                        <v-tooltip :text="`${item.rx.toLocaleString()} bytes`">
                                            <template #activator="{ props }">
                                                <span v-bind="props" class="text-info">
                                                    {{ bytesToGB(item.rx, 6) }} GB
                                                </span>
                                            </template>
                                        </v-tooltip>
                                    </div>
                                    <div>
                                        TX:
                                        <span
                                            v-if="item.traffic_type != ModelsOcservUserTrafficTypeEnum.FREE"
                                            class="text-muted text-subtitle-2"
                                        >
                                            ({{ t('CURRENT') }})
                                        </span>
                                        <br />
                                        <v-tooltip :text="`${item.tx.toLocaleString()} bytes`">
                                            <template #activator="{ props }">
                                                <span v-bind="props" class="text-info">
                                                    {{ bytesToGB(item.tx, 4) }} GB
                                                </span>
                                            </template>
                                        </v-tooltip>
                                    </div>
                                </td>
                                <td class="text-capitalize">
                                    <div>
                                        {{ t('EXPIRE_AT') }}:<br />
                                        <span class="text-info text-capitalize">
                                            {{ formatDate(item.expire_at) }}
                                        </span>
                                    </div>
                                    <div>
                                        {{ t('DEACTIVATED_AT') }}:<br />
                                        <span class="text-info text-capitalize">
                                            {{ formatDate(item.deactivated_at) || t('USER_IS_ACTIVE_NOW') }}
                                        </span>
                                    </div>
                                </td>
                                <td>
                                    <div class="text-capitalize">
                                        {{ t('STATUS') }}:<br />
                                        <!-- Locked -->
                                        <span v-if="item.is_locked && item.deactivated_at == null">
                                            <v-icon color="warning" start>mdi-lock</v-icon>
                                            <span class="text-warning text-capitalize">{{ t('LOCKED') }}</span>
                                        </span>

                                        <!-- Deactivated -->
                                        <span v-if="item.is_locked && item.deactivated_at">
                                            <v-icon color="accent" start>mdi-close-network-outline</v-icon>
                                            <span class="text-accent text-capitalize">{{ t('DEACTIVATED') }}</span>
                                        </span>

                                        <!-- Disconnected -->
                                        <span v-else-if="!item.is_online">
                                            <v-icon color="error" start>mdi-lan-disconnect</v-icon>
                                            <span class="text-error text-capitalize">{{ t('DISCONNECTED') }}</span>
                                        </span>

                                        <!-- Online -->
                                        <span v-else>
                                            <v-icon color="success" start>mdi-lan-connect</v-icon>
                                            <span class="text-success text-capitalize">{{ t('ONLINE') }}</span>
                                        </span>
                                    </div>
                                </td>
                                <td>
                                    <v-menu>
                                        <template v-slot:activator="{ props }">
                                            <v-icon start v-bind="props"> mdi-dots-vertical</v-icon>
                                        </template>

                                        <v-list>
                                            <v-list-item @click="detailUser(item?.uid)">
                                                <v-list-item-title class="text-primary text-capitalize me-5">
                                                    {{ t('DETAIL') }}
                                                </v-list-item-title>
                                                <template v-slot:prepend>
                                                    <v-icon class="ms-2" color="primary">mdi-information-outline</v-icon>
                                                </template>
                                            </v-list-item>

                                            <v-list-item
                                                @click="editUser(item?.uid)"
                                                v-if="!(item.is_locked && item.deactivated_at)"
                                            >
                                                <v-list-item-title class="text-info text-capitalize me-5">
                                                    {{ t('UPDATE') }}
                                                </v-list-item-title>
                                                <template v-slot:prepend>
                                                    <v-icon class="ms-2" color="info">mdi-pencil</v-icon>
                                                </template>
                                            </v-list-item>

                                            <v-list-item
                                                v-if="item.is_online && !item.is_locked && !item.deactivated_at"
                                                @click="disconnect(item?.username)"
                                            >
                                                <v-list-item-title class="text-grey text-capitalize me-5">
                                                    {{ t('DISCONNECT') }}
                                                </v-list-item-title>
                                                <template v-slot:prepend>
                                                    <v-icon class="ms-2" color="grey">mdi-lan-disconnect</v-icon>
                                                </template>
                                            </v-list-item>

                                            <v-list-item
                                                v-if="!item.is_locked && !item.deactivated_at"
                                                @click="lock(item?.uid)"
                                            >
                                                <v-list-item-title class="text-warning text-capitalize me-5">
                                                    {{ t('LOCK') }}
                                                </v-list-item-title>
                                                <template v-slot:prepend>
                                                    <v-icon class="ms-2" color="warning">mdi-lock</v-icon>
                                                </template>
                                            </v-list-item>

                                            <v-list-item
                                                v-if="item.is_locked && !item.deactivated_at"
                                                @click="unlock(item?.uid)"
                                            >
                                                <v-list-item-title class="text-grey text-capitalize me-5">
                                                    {{ t('UNLOCK') }}
                                                </v-list-item-title>
                                                <template v-slot:prepend>
                                                    <v-icon class="ms-2" color="grey">mdi-lock</v-icon>
                                                </template>
                                            </v-list-item>

                                            <v-list-item @click="statistics(item?.uid, item.username)">
                                                <v-list-item-title class="text-grey text-capitalize me-5">
                                                    {{ t('STATISTICS') }}
                                                </v-list-item-title>
                                                <template v-slot:prepend>
                                                    <v-icon class="ms-2" color="grey">mdi-chart-bar-stacked</v-icon>
                                                </template>
                                            </v-list-item>

                                            <v-list-item @click="deleteUserHandler(item?.uid, item.username)">
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

                    <Pagination :meta="meta" @update="updateMeta" />
                </div>

                <div v-else class="ms-md-5 mb-md-5 text-capitalize">
                    {{ t('NO_USER_FOUND_TABLE') }}
                </div>
            </UiParentCard>
        </v-col>
    </v-row>

    <DeleteDialog :show="deleteDialog" :username="deleteUserName" @close="cancelDeleteUser" @deleteUser="deleteUser" />
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

<script lang="ts" setup>
import { router } from '@/router';
import UiParentCard from '@/components/shared/UiParentCard.vue';
import { useI18n } from 'vue-i18n';
import { onMounted, reactive, ref } from 'vue';
import {
    type ModelsOcservUser,
    ModelsOcservUserTrafficTypeEnum,
    OcservUsersApi,
    type OcservUsersGetFilterEnum,
    ReportApi,
    type ReportOcservUserReportResponse
} from '@/api';
import { getAuthorization } from '@/utils/request';
import { bytesToGB, formatDate, trafficTypesTransformer } from '@/utils/convertors';
import DeleteDialog from '@/components/ocserv_user/DeleteDialog.vue';
import Pagination from '@/components/shared/Pagination.vue';
import type { Meta } from '@/types/metaTypes/MetaType';
import { useSnackbarStore } from '@/stores/snackbar';
import { useProfileStore } from '@/stores/profile';
import SessionLogsDialog from '@/components/ocserv_user/SessionLogsDialog.vue';
import StatisticsDialog from '@/components/ocserv_user/StatisticsDialog.vue';

const { t } = useI18n();
const loading = ref(false);
const q = ref('');
const api = new OcservUsersApi();
const meta = reactive<Meta>({
    page: 1,
    size: 10,
    sort: 'ASC',
    total_records: 0
});

const deleteDialog = ref(false);
const deleteUserName = ref('');
const deleteUserUID = ref('');

const activateDialog = ref(false);
const activateUserName = ref('');
const activateUserUID = ref('');

const statisticsDialog = ref(false);
const statisticsUsername = ref('');
const statisticsUID = ref('');

const sessionLogsDialog = ref(false);
const sessionLogsUsername = ref('');
const sessionLogsUID = ref('');

const users = ref<ModelsOcservUser[]>([]);
const snackbar = useSnackbarStore();

const profileStore = useProfileStore();
const isAdmin = ref(profileStore.isAdmin);

const userStats = ref<ReportOcservUserReportResponse>({
    active: 0,
    deactivated: 0,
    online: 0,
    locked: 0
});

const filter = ref<OcservUsersGetFilterEnum>();

const getUsers = () => {
    loading.value = true;
    api.ocservUsersGet({
        ...getAuthorization(),
        ...meta,
        q: q.value,
        filter: filter.value
    })
        .then((res) => {
            users.value = res.data.result ?? [];
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
            let index = users.value.findIndex((i) => i.username === username);
            if (index > -1) {
                users.value[index].is_online = false;
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
            let index = users.value.findIndex((i) => i.uid === uid);
            if (index > -1) {
                users.value[index].is_locked = true;
            }
            getUserStats();
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
            let index = users.value.findIndex((i) => i.uid === uid);
            if (index > -1) {
                users.value[index].is_locked = false;
            }
            getUserStats();
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

const activateUser = (expireAt: string) => {
    expireAt = formatDate(expireAt);
    api.ocservUsersUidActivatePost({
        ...getAuthorization(),
        uid: activateUserUID.value,
        request: {
            expire_at: expireAt
        }
    })
        .then(() => {
            let index = users.value.findIndex((i) => i.uid === activateUserUID.value);
            if (index > -1) {
                users.value[index].is_locked = false;
                users.value[index].deactivated_at = undefined;
                users.value[index].expire_at = expireAt;
                users.value[index].is_online = false;
            }
            getUserStats();
        })
        .finally(() => {
            cancelActivateUser();
            snackbar.show({
                id: 1,
                message: t('USER_ACTIVATE_SUCCESSFULLY_SNACK'),
                color: 'success',
                timeout: 4000
            });
        });
};

const statistics = async (uid: string, username: string) => {
    statisticsDialog.value = true;
    statisticsUsername.value = username;
    statisticsUID.value = uid;
};

const sessionLogs = async (uid: string, username: string) => {
    sessionLogsDialog.value = true;
    sessionLogsUsername.value = username;
    sessionLogsUID.value = uid;
};

const deleteUserHandler = (uid: string, username: string) => {
    deleteUserUID.value = uid;
    deleteUserName.value = username;
    deleteDialog.value = true;
};

const activateUserHandler = (uid: string, username: string) => {
    activateUserUID.value = uid;
    activateUserName.value = username;
    activateDialog.value = true;
};

const cancelDeleteUser = () => {
    deleteUserUID.value = '';
    deleteUserName.value = '';
    deleteDialog.value = false;
};

const cancelActivateUser = () => {
    activateUserUID.value = '';
    activateUserName.value = '';
    activateDialog.value = false;
};

const closeStatisticsDialog = () => {
    statisticsUID.value = '';
    statisticsUsername.value = '';
    statisticsDialog.value = false;
};

const closeSessionLogsDialog = () => {
    sessionLogsUID.value = '';
    sessionLogsUsername.value = '';
    sessionLogsDialog.value = false;
};

const deleteUser = () => {
    api.ocservUsersUidDelete({
        ...getAuthorization(),
        uid: deleteUserUID.value
    })
        .then((_) => {
            getUsers();
            getUserStats();
        })
        .finally(() => {
            cancelDeleteUser();
        });
};

const updateMeta = (newMeta: Meta) => {
    Object.assign(meta, newMeta);
    getUsers();
};

const search = (clear: boolean = false) => {
    if (clear) {
        q.value = '';
    }

    if (q.value.length > 1 || clear || filter.value) {
        if (q.value.length < 2) {
            q.value = '';
        }
        getUsers();
    }
};

const reload = () => {
    q.value = '';
    filter.value = undefined;
    getUsers();
};

const getUserStats = () => {
    const apiStats = new ReportApi();
    apiStats
        .reportsUsersGet({
            ...getAuthorization()
        })
        .then((res) => {
            Object.assign(userStats.value, res.data);
        });
};

onMounted(() => {
    getUserStats();
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
                        variant="outlined"
                        @click="router.push({ name: 'Ocserv User Create' })"
                    >
                        {{ t('CREATE') }}
                    </v-btn>
                </template>

                <div class="mx-15 mb-5">
                    <v-row align="center" justify="center">
                        <v-col cols="12" lg="2" sm="6">
                            <v-card class="text-center" elevation="10">
                                <v-card-title class="text-subtitle-1 mt-2 text-capitalize">
                                    {{ t('ONLINE') }} {{ t('USERS') }}
                                </v-card-title>

                                <v-card-text class="text-muted text-h5">
                                    {{ userStats.online || 0 }}
                                </v-card-text>
                            </v-card>
                        </v-col>

                        <v-col cols="12" lg="2" sm="6">
                            <v-card class="text-center" elevation="10">
                                <v-card-title class="text-subtitle-1 mt-2 text-capitalize">
                                    {{ t('ACTIVE') }} {{ t('USERS') }}
                                </v-card-title>

                                <v-card-text class="text-muted text-h5">
                                    {{ userStats.active || 0 }}
                                </v-card-text>
                            </v-card>
                        </v-col>

                        <v-col cols="12" lg="2" sm="6">
                            <v-card class="text-center" elevation="10">
                                <v-card-title class="text-subtitle-1 mt-2 text-capitalize">
                                    {{ t('DEACTIVATED') }} {{ t('USERS') }}
                                </v-card-title>

                                <v-card-text class="text-muted text-h5">
                                    {{ userStats.deactivated }}
                                </v-card-text>
                            </v-card>
                        </v-col>

                        <v-col cols="12" lg="2" sm="6">
                            <v-card class="text-center" elevation="10">
                                <v-card-title class="text-subtitle-1 mt-2 text-capitalize">
                                    {{ t('LOCKED') }} {{ t('USERS') }}
                                </v-card-title>

                                <v-card-text class="text-muted text-h5">
                                    {{ userStats.locked }}
                                </v-card-text>
                            </v-card>
                        </v-col>
                    </v-row>
                </div>

                <v-progress-linear :active="loading" indeterminate></v-progress-linear>

                <div v-if="!loading">
                    <div class="mb-3">
                        <v-row align="center" class="px-md-15 mb-3 text-capitalize" justify="start">
                            <v-col cols="12" md="3" sm="5">
                                <v-text-field
                                    v-model="q"
                                    :label="t('USERNAME')"
                                    clearable
                                    color="primary"
                                    density="compact"
                                    hide-details
                                    variant="outlined"
                                    @click:clear="search(true)"
                                    @keyup.enter.native="search(false)"
                                />
                            </v-col>

                            <v-col cols="12" md="auto" sm="5" class="ma-0 pa-0 mt-5 me-5">
                                <v-radio-group inline v-model="filter">
                                    <v-radio value="active" :label="t('ACTIVE')" hide-details />
                                    <v-radio value="online" :label="t('ONLINE')" hide-details />
                                    <v-radio value="deactivated" :label="t('DEACTIVATED')" hide-details />
                                    <v-radio value="locked" :label="t('LOCKED')" hide-details />
                                </v-radio-group>
                            </v-col>

                            <v-col class="ma-0 pa-0" cols="12" md="auto">
                                <v-btn color="info" size="small" @click="search(false)">
                                    <v-icon start>mdi-magnify</v-icon>
                                    {{ t('SEARCH') }}
                                </v-btn>
                            </v-col>

                            <v-col cols="12" md="auto">
                                <v-btn color="secondary" size="small" variant="outlined" @click="reload">
                                    {{ t('RELOAD') }}
                                </v-btn>
                            </v-col>
                        </v-row>
                    </div>
                    <v-table v-if="users.length > 0" class="px-md-15">
                        <thead>
                            <tr class="text-capitalize bg-lightprimary">
                                <th class="text-left">{{ t('USERNAME') }}</th>
                                <th v-if="isAdmin" class="text-left">{{ t('OWNER') }}</th>
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
                                                <span class="text-info" v-bind="props">
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
                                                <span class="text-info" v-bind="props">
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
                                            {{ formatDate(item.expire_at) || t('UNLIMITED') }}
                                        </span>
                                    </div>
                                    <div v-if="item.deactivated_at">
                                        {{ t('DEACTIVATED_AT') }}:<br />
                                        <span class="text-info text-capitalize">
                                            {{ formatDate(item.deactivated_at) }}
                                        </span>
                                    </div>
                                </td>
                                <td>
                                    <div class="text-capitalize">
                                        {{ t('STATUS') }}:<br />
                                        <!-- Locked -->
                                        <span v-if="item.is_locked && !Boolean(item.deactivated_at)">
                                            <v-icon color="warning" start>mdi-lock</v-icon>
                                            <span class="text-warning text-capitalize">{{ t('LOCKED') }}</span>
                                        </span>

                                        <!-- Deactivated -->
                                        <span v-else-if="Boolean(item.deactivated_at)">
                                            <v-icon color="error" start>mdi-close-network-outline</v-icon>
                                            <span class="text-error text-capitalize">{{ t('DEACTIVATED') }}</span>
                                        </span>

                                        <!-- Online -->
                                        <span v-else-if="item.is_online">
                                            <v-icon color="success" start>mdi-lan-connect</v-icon>
                                            <span class="text-success text-capitalize">{{ t('ONLINE') }}</span>
                                        </span>

                                        <!-- Disconnected -->
                                        <span v-else-if="!item.is_online">
                                            <v-icon color="grey" start>mdi-lan-disconnect</v-icon>
                                            <span class="text-grey text-capitalize">{{ t('DISCONNECTED') }}</span>
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
                                                v-if="!(item.is_locked && item.deactivated_at)"
                                                @click="editUser(item?.uid)"
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
                                                <v-list-item-title class="text-error text-capitalize me-5">
                                                    {{ t('DISCONNECT') }}
                                                </v-list-item-title>
                                                <template v-slot:prepend>
                                                    <v-icon class="ms-2" color="error">mdi-lan-disconnect</v-icon>
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

                                            <v-list-item
                                                v-if="item.deactivated_at"
                                                @click="activateUserHandler(item.uid, item.username)"
                                            >
                                                <v-list-item-title class="text-success text-capitalize me-5">
                                                    {{ t('ACTIVATE') }}
                                                </v-list-item-title>
                                                <template v-slot:prepend>
                                                    <v-icon class="ms-2" color="success">mdi-network-outline</v-icon>
                                                </template>
                                            </v-list-item>

                                            <v-list-item @click="statistics(item.uid, item.username)">
                                                <v-list-item-title class="text-grey text-capitalize me-5">
                                                    {{ t('STATISTICS') }}
                                                </v-list-item-title>
                                                <template v-slot:prepend>
                                                    <v-icon class="ms-2" color="grey">mdi-chart-bar-stacked</v-icon>
                                                </template>
                                            </v-list-item>

                                            <v-list-item @click="sessionLogs(item.uid, item.username)">
                                                <v-list-item-title class="text-grey text-capitalize me-5">
                                                    {{ t('SESSION_LOGS') }}
                                                </v-list-item-title>
                                                <template v-slot:prepend>
                                                    <v-icon class="ms-2" color="grey">mdi-timeline-text-outline</v-icon>
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
                </div>

                <div v-if="loading || users.length == 0" class="ms-md-5 mb-md-5 text-capitalize">
                    {{ t('NO_USER_FOUND_TABLE') }}
                </div>

                <Pagination :totalRecords="meta.total_records" @update="updateMeta" />
            </UiParentCard>
        </v-col>
    </v-row>

    <ActivateDialog
        :show="activateDialog"
        :username="activateUserName"
        @close="cancelActivateUser"
        @activateUser="activateUser"
    />

    <DeleteDialog :show="deleteDialog" :username="deleteUserName" @close="cancelDeleteUser" @deleteUser="deleteUser" />

    <StatisticsDialog
        :show="statisticsDialog"
        :username="statisticsUsername"
        :uid="statisticsUID"
        @close="closeStatisticsDialog"
    />

    <SessionLogsDialog
        :show="sessionLogsDialog"
        :username="sessionLogsUsername"
        :uid="sessionLogsUID"
        @close="closeSessionLogsDialog"
    />
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

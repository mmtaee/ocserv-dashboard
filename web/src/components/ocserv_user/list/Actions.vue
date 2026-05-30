<script setup lang="ts">
import { type ModelsOcservUser, type ModelsOnlineUserSession, OcservUsersApi } from '@/api';
import { getAuthorization } from '@/utils/request';
import { useI18n } from 'vue-i18n';
import { router } from '@/router';
import { useSnackbarStore } from '@/stores/snackbar';
import DeleteDialog from '@/components/ocserv_user/dialogs/DeleteDialog.vue';
import StatisticsDialog from '@/components/ocserv_user/dialogs/StatisticsDialog.vue';
import ActivateDialog from '@/components/ocserv_user/dialogs/ActivateDialog.vue';
import SessionLogsDialog from '@/components/ocserv_user/dialogs/SessionLogsDialog.vue';
import DisconnectDialog from '@/components/ocserv_user/dialogs/DisconnectDialog.vue';
import { ref } from 'vue';
import { formatDate } from 'date-fns';

defineProps<{ item: ModelsOcservUser }>();

const emit = defineEmits(['actions']);

const { t } = useI18n();
const snackbar = useSnackbarStore();
const api = new OcservUsersApi();

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

const disconnectDialog = ref(false);
const disconnectUsername = ref('');
const disconnectSessions = ref<ModelsOnlineUserSession[]>([]);

const detailUser = async (uid: string) => {
    await router.push({ name: 'Ocserv User Detail', params: { uid: uid } });
};

const editUser = async (uid: string) => {
    await router.push({ name: 'Ocserv User Update', params: { uid: uid } });
};

const notifier = (msg: string) => {
    snackbar.show({
        id: 1,
        message: msg,
        color: 'success',
        timeout: 4000
    });
};

type ActionTypes = 'session' | 'all';

const disconnect = (type: ActionTypes, sessionID: string | null) => {
    switch (type) {
        case 'session': {
            if (sessionID) {
                api.ocservUsersIdDisconnectByIdPost({
                    ...getAuthorization(),
                    id: sessionID
                })
                    .then(() => {
                        notifier(t('USER_DISCONNECTED_SUCCESS_SNACK'));
                        emit('actions', 'disconnect_session', disconnectUsername.value, { id: sessionID });
                    })
                    .finally(() => {
                        if (disconnectSessions.value.length == 0) {
                            disconnectDialog.value = false;
                        }
                    });
            }
            break;
        }
        case 'all': {
            api.ocservUsersUsernameDisconnectPost({
                ...getAuthorization(),
                username: disconnectUsername.value
            }).then(() => {
                emit('actions', 'disconnect', disconnectUsername.value);
                disconnectDialog.value = false;
                notifier(t('USER_DISCONNECTED_SUCCESS_SNACK'));
            });
            break;
        }
    }
};

const terminate = (type: ActionTypes, sessionID: string | null) => {
    switch (type) {
        case 'session': {
            if (sessionID) {
                api.ocservUsersIdTerminateByIdPost({
                    ...getAuthorization(),
                    id: sessionID
                })
                    .then(() => {
                        notifier(t('USER_TERMINATED_SUCCESS_SNACK'));
                        emit('actions', 'terminate_session', disconnectUsername.value, { id: sessionID });
                    })
                    .finally(() => {
                        if (disconnectSessions.value.length == 0) {
                            disconnectDialog.value = false;
                        }
                    });
            }
            break;
        }
        case 'all': {
            api.ocservUsersUsernameTerminatePost({
                ...getAuthorization(),
                username: disconnectUsername.value
            }).then(() => {
                emit('actions', 'terminate', disconnectUsername.value);
                disconnectDialog.value = false;
                notifier(t('USER_TERMINATED_SUCCESS_SNACK'));
            });
            break;
        }
    }
};

const lock = (uid: string) => {
    api.ocservUsersUidLockPost({
        ...getAuthorization(),
        uid: uid
    }).then(() => {
        emit('actions', 'lock', uid);
        notifier(t('USER_LOCKED_SUCCESSFULLY_SNACK'));
    });
};

const unlock = (uid: string) => {
    api.ocservUsersUidUnlockPost({
        ...getAuthorization(),
        uid: uid
    }).then(() => {
        emit('actions', 'unlock', uid);
        notifier(t('USER_UNLOCKED_SUCCESSFULLY_SNACK'));
    });
};

const activateUser = (expireAt: string | null) => {
    if (expireAt == null) {
        expireAt = '';
    }

    const formattedExpireAt = formatDate(expireAt, '');

    api.ocservUsersUidActivatePost({
        ...getAuthorization(),
        uid: activateUserUID.value,
        request: {
            expire_at: formattedExpireAt || undefined
        }
    }).then(() => {
        emit('actions', 'activateUser', activateUserUID.value, {
            formattedExpireAt: formattedExpireAt
        });
        cancelActivateUser();
        notifier(t('USER_ACTIVATE_SUCCESSFULLY_SNACK'));
    });
};

const downloadCertificate = (uid: string, username: string) => {
    api.ocservUsersUidCertificateGet(
        {
            ...getAuthorization(),
            uid: uid
        },
        {
            responseType: 'blob'
        }
    ).then((res) => {
        const blob = res.data instanceof Blob
            ? res.data
            : new Blob([res.data], { type: 'application/x-pkcs12' });

        const url = window.URL.createObjectURL(blob);
        const link = document.createElement('a');

        link.href = url;
        link.setAttribute('download', `${username}.p12`);
        document.body.appendChild(link);
        link.click();
        link.remove();

        window.URL.revokeObjectURL(url);
    });
};

const disconnectHandler = async (username: string, sessions: ModelsOnlineUserSession[]) => {
    disconnectDialog.value = true;
    disconnectUsername.value = username;
    disconnectSessions.value = sessions;
};

const statisticsHandler = async (uid: string, username: string) => {
    statisticsDialog.value = true;
    statisticsUsername.value = username;
    statisticsUID.value = uid;
};

const sessionLogsHandler = async (uid: string, username: string) => {
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

const cancelDisconnect = () => {
    disconnectDialog.value = false;
    disconnectUsername.value = '';
    disconnectSessions.value = [];
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
            emit('actions', 'deleteUser', _);
        })
        .finally(() => {
            cancelDeleteUser();
        });
};
</script>

<template>
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

            <v-list-item v-if="!(item.is_locked && item.deactivated_at)" @click="editUser(item?.uid)">
                <v-list-item-title class="text-info text-capitalize me-5">
                    {{ t('UPDATE') }}
                </v-list-item-title>
                <template v-slot:prepend>
                    <v-icon class="ms-2" color="info">mdi-pencil</v-icon>
                </template>
            </v-list-item>

            <v-list-item
                v-if="item.is_online && !item.is_locked && !item.deactivated_at"
                @click="disconnectHandler(item?.username, item.online_sessions)"
            >
                <v-list-item-title class="text-error text-capitalize me-5">
                    {{ t('DISCONNECT') }}
                </v-list-item-title>
                <template v-slot:prepend>
                    <v-icon class="ms-2" color="error">mdi-lan-disconnect</v-icon>
                </template>
            </v-list-item>

            <v-list-item v-if="!item.is_locked && !item.deactivated_at" @click="lock(item?.uid)">
                <v-list-item-title class="text-warning text-capitalize me-5">
                    {{ t('LOCK') }}
                </v-list-item-title>
                <template v-slot:prepend>
                    <v-icon class="ms-2" color="warning">mdi-lock</v-icon>
                </template>
            </v-list-item>

            <v-list-item v-if="item.is_locked && !item.deactivated_at" @click="unlock(item?.uid)">
                <v-list-item-title class="text-grey text-capitalize me-5">
                    {{ t('UNLOCK') }}
                </v-list-item-title>
                <template v-slot:prepend>
                    <v-icon class="ms-2" color="grey">mdi-lock</v-icon>
                </template>
            </v-list-item>

            <v-list-item v-if="item.deactivated_at" @click="activateUserHandler(item.uid, item.username)">
                <v-list-item-title class="text-success text-capitalize me-5">
                    {{ t('ACTIVATE') }}
                </v-list-item-title>
                <template v-slot:prepend>
                    <v-icon class="ms-2" color="success">mdi-network-outline</v-icon>
                </template>
            </v-list-item>

            <v-list-item @click="statisticsHandler(item.uid, item.username)">
                <v-list-item-title class="text-grey text-capitalize me-5">
                    {{ t('STATISTICS') }}
                </v-list-item-title>
                <template v-slot:prepend>
                    <v-icon class="ms-2" color="grey">mdi-chart-bar-stacked</v-icon>
                </template>
            </v-list-item>

            <v-list-item @click="sessionLogsHandler(item.uid, item.username)">
                <v-list-item-title class="text-grey text-capitalize me-5">
                    {{ t('SESSION_LOGS') }}
                </v-list-item-title>
                <template v-slot:prepend>
                    <v-icon class="ms-2" color="grey">mdi-timeline-text-outline</v-icon>
                </template>
            </v-list-item>

            <v-list-item @click="downloadCertificate(item.uid, item.username)" v-if="item.certificate_enabled">
                <v-list-item-title class="text-success text-capitalize me-5">
                    {{ t('DOWNLOAD_CERTIFICATE') }}
                </v-list-item-title>
                <template v-slot:prepend>
                    <v-icon class="ms-2" color="success">mdi-certificate-outline</v-icon>
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

    <DisconnectDialog
        :show="disconnectDialog"
        :username="disconnectUsername"
        :sessions="disconnectSessions"
        @close="cancelDisconnect"
        @disconnect="disconnect"
        @terminate="terminate"
    />

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

<style scoped lang="scss"></style>

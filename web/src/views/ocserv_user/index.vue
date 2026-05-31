<script lang="ts" setup>
import { router } from '@/router';
import UiParentCard from '@/components/shared/UiParentCard.vue';
import { useI18n } from 'vue-i18n';
import { reactive, ref } from 'vue';
import {
    type ModelsOcservUser,
    ModelsOcservUserTrafficTypeEnum,
    OcservUsersApi,
    type OcservUsersGetFilterEnum
} from '@/api';
import { getAuthorization } from '@/utils/request';
import { bytesToGB, bytesToTrafficSize, formatDate, trafficTypesTransformer } from '@/utils/convertors';
import Pagination from '@/components/shared/Pagination.vue';
import type { Meta } from '@/types/metaTypes/MetaType';
import { useProfileStore } from '@/stores/profile';
import Status from '@/components/ocserv_user/list/Status.vue';
import Actions from '@/components/ocserv_user/list/Actions.vue';
import Stats from '@/components/ocserv_user/list/Stats.vue';
import SearchAndFilter from '@/components/ocserv_user/list/SearchAndFilter.vue';

const statsRef = ref<InstanceType<typeof Stats> | null>(null);

const { t } = useI18n();
const loading = ref(false);

const api = new OcservUsersApi();
const meta = reactive<Meta>({
    page: 1,
    size: 10,
    sort: 'ASC',
    total_records: 0
});

const users = ref<ModelsOcservUser[]>([]);
const profileStore = useProfileStore();
const isAdmin = ref(profileStore.isAdmin);

const getUsers = (
    q: string | null = null,
    filter: OcservUsersGetFilterEnum | undefined = undefined,
    group: string | null = null
) => {
    loading.value = true;
    api.ocservUsersGet({
        ...getAuthorization(),
        ...meta,
        q: q || '',
        filter: filter,
        group: group || undefined
    })
        .then((res) => {
            users.value = res.data.result ?? [];
            Object.assign(meta, res.data.meta);
        })
        .finally(() => {
            loading.value = false;
        });
};

const updateMeta = (newMeta: Meta) => {
    Object.assign(meta, newMeta);
    getUsers(null);
};

const reloadStats = () => {
    statsRef.value?.getUserStats();
};

type ActionTypes =
    | 'lock'
    | 'unlock'
    | 'disconnect'
    | 'disconnect_session'
    | 'terminate'
    | 'terminate_session'
    | 'delete'
    | 'activate';

type ActivateExtra = {
    formattedExpireAt: string;
    id: string | number;
};

const actions = (act: ActionTypes, identifier: string, extra: ActivateExtra | null = null) => {
    switch (act) {
        case 'lock': {
            const index = users.value.findIndex((i) => i.uid === identifier);

            if (index > -1) {
                users.value[index].is_locked = true;
            }

            reloadStats();
            break;
        }

        case 'unlock': {
            const index = users.value.findIndex((i) => i.uid === identifier);

            if (index > -1) {
                users.value[index].is_locked = false;
            }

            reloadStats();
            break;
        }

        case 'disconnect':
        case 'terminate': {
            const index = users.value.findIndex((i) => i.username === identifier);
            if (index > -1) {
                users.value[index].is_online = false;
                users.value[index].online_sessions.splice(0);
            }
            break;
        }

        case 'disconnect_session':
        case 'terminate_session': {
            if (extra?.id == null) return;
            const index = users.value.findIndex((i) => i.username === identifier);
            if (index > -1) {
                let sessionIndex = users.value[index].online_sessions.findIndex((i) => i.ID === extra.id);
                if (sessionIndex > -1) {
                    users.value[index].online_sessions.splice(sessionIndex, 1);
                }
            }

            if (users.value[index].online_sessions.length == 0) {
                users.value[index].is_online = false;
            }
            break;
        }

        case 'delete': {
            getUsers(null);
            reloadStats();
            break;
        }

        case 'activate': {
            const index = users.value.findIndex((i) => i.uid === identifier);

            if (index > -1 && extra) {
                users.value[index].is_locked = false;
                users.value[index].deactivated_at = undefined;
                users.value[index].expire_at = extra.formattedExpireAt || undefined;
                users.value[index].is_online = false;
                users.value[index].rx = 0;
                users.value[index].tx = 0;
            }

            reloadStats();
            break;
        }
    }
};
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

                <Stats ref="statsRef" />

                <SearchAndFilter @getUsers="getUsers" />

                <v-progress-linear :active="loading" indeterminate></v-progress-linear>

                <div>
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
                                <th class="text-left">{{ t('CERTIFICATE') }}</th>
                                <th class="text-left">{{ t('ACTION') }}</th>
                            </tr>
                        </thead>
                        <tbody v-if="!loading">
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
					    {{ bytesToTrafficSize(item.traffic_size) }}
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
                                            class="text-medium-emphasis text-subtitle-2"
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
                                            class="text-medium-emphasis text-subtitle-2"
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
                                    <div>
                                        {{ t('TOTAL') }}:
                                        <span
                                            v-if="item.traffic_type != ModelsOcservUserTrafficTypeEnum.FREE"
                                            class="text-muted text-subtitle-2"
                                        >
                                            ({{ t('CURRENT') }})
                                        </span>
                                        <br />
                                        <v-tooltip :text="`${(item.rx + item.tx).toLocaleString()} bytes`">
                                            <template #activator="{ props }">
                                                <span class="text-info" v-bind="props">
                                                    {{ bytesToGB(item.rx + item.tx, 4) }} GB
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
                                    <Status :item="item" />
                                </td>

                                <td>
                                    <span
                                        :class="item.certificate_enabled ? 'text-success' : 'text-warning'"
                                        class="text-capitalize"
                                    >
                                        {{ item.certificate_enabled ? t('ENABLED') : t('DISABLED') }}
                                    </span>
                                </td>
                                <td>
                                    <Actions :item="item" @actions="actions" />
                                </td>
                            </tr>
                        </tbody>
                        <tbody v-if="loading || users.length == 0">
                        {{ t('NO_USER_FOUND_TABLE') }}
                        </tbody>
                    </v-table>
                </div>

                <Pagination :totalRecords="meta.total_records" @update="updateMeta" />
            </UiParentCard>
        </v-col>
    </v-row>
</template>

<!--suppress CssUnresolvedCustomProperty -->
<style scoped>
tbody tr:nth-child(even) td {
    background-color: rgba(var(--v-theme-on-surface), 0.04);
}

@media (min-width: 992px) {
    tbody tr:nth-child(even) td {
        background-color: rgba(var(--v-theme-on-surface), 0.04);
    }

    tbody tr:nth-child(even) td:first-child {
        border-radius: 8px 0 0 8px;
    }

    tbody tr:nth-child(even) td:last-child {
        border-radius: 0 8px 8px 0;
    }
}
</style>

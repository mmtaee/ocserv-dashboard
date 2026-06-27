<script lang="ts" setup>
import UiParentCard from '@/components/shared/UiParentCard.vue';
import { useI18n } from 'vue-i18n';
import UiChildCard from '@/components/shared/UiChildCard.vue';
import {
    type ModelsOcservUser,
    ModelsOcservUserTrafficTypeEnum,
    OcservGroupsApi,
    OcservUsersApi,
    type OcservUserUpdateOcservUserData
} from '@/api';
import { getAuthorization } from '@/utils/request';
import { onBeforeMount, onMounted, ref } from 'vue';
import { router } from '@/router';
import OcservUserForm from '@/components/ocserv_user/OcservUserForm.vue';
import { useSnackbarStore } from '@/stores/snackbar';

const props = defineProps<{ uid: string }>();

const { t } = useI18n();
const loading = ref(false);
const result = ref<ModelsOcservUser>({
    created_at: '',
    group: '',
    is_locked: false,
    is_online: false,
    owner: '',
    password: '',
    rx: 0,
    traffic_size: 0,
    traffic_type: ModelsOcservUserTrafficTypeEnum.FREE,
    tx: 0,
    uid: '',
    username: '',
    online_sessions: []
});
const groups = ref<string[]>([]);

const api = new OcservUsersApi();

const updateUser = (uid: string, data: OcservUserUpdateOcservUserData) => {
    loading.value = true;
    api.ocservUsersUidPatch({
        ...getAuthorization(),
        uid: uid,
        request: data
    })
        .then(() => {
            const snackbar = useSnackbarStore();
            snackbar.show({
                id: 1,
                message: t('SNACK_OCSERV_USER_UPDATE_SUCCESS'),
                color: 'success',
                timeout: 4000
            });
            router.push({ name: 'Ocserv User Detail', params: { uid: uid } });
        })
        .finally(() => {
            loading.value = false;
        });
};

const getUser = () => {
    if (props.uid == undefined) {
        return;
    }
    const api = new OcservUsersApi();
    api.ocservUsersUidGet({
        ...getAuthorization(),
        uid: props.uid
    }).then((res) => {
        result.value = res.data;
    });
};

onMounted(() => {
    getUser();
});

onBeforeMount(() => {
    const api = new OcservGroupsApi();
    api.ocservGroupsLookupGet({
        ...getAuthorization()
    }).then((res) => {
        groups.value = res.data;
    });
});
</script>

<template>
    <v-row>
        <v-col cols="12" md="12">
            <UiParentCard :title="t('UPDATE_OCSERV_USER_TITLE')">
                <template #action>
                    <v-btn
                        class="me-lg-5"
                        color="grey"
                        size="small"
                        variant="flat"
                        @click="router.push({ name: 'Ocserv Users' })"
                    >
                        {{ t('CANCEL') }}
                    </v-btn>
                </template>
                <UiChildCard class="px-3">
                    <OcservUserForm
                        :btnText="t('UPDATE')"
                        :groups="groups"
                        :initData="result"
                        :loading="loading"
                        @updateUser="updateUser"
                    />
                </UiChildCard>
            </UiParentCard>
        </v-col>
    </v-row>
</template>

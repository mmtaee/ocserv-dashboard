<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { onMounted, ref } from 'vue';
import { OcservGroupsApi, type OcservUsersGetFilterEnum } from '@/api';
import { getAuthorization } from '@/utils/request';

const emit = defineEmits(['getUsers']);

const { t } = useI18n();
const q = ref('');
const filter = ref<OcservUsersGetFilterEnum>();
const group = ref<string | null>(null);
const groups = ref<string[]>([]);

const getGroups = () => {
    const groupsApi = new OcservGroupsApi();
    groupsApi
        .ocservGroupsLookupGet({ ...getAuthorization() })
        .then((res) => {
            groups.value = (res.data ?? []).filter(Boolean);
        })
        .catch(() => {
            groups.value = [];
        });
};

const search = (clear: boolean = false) => {
    if (clear) {
        q.value = '';
    }

    if (q.value.length > 1 || clear || filter.value || group.value) {
        if (q.value.length < 2) {
            q.value = '';
        }
        emit('getUsers', q.value, filter.value, group.value);
    }
};

const reload = () => {
    q.value = '';
    filter.value = undefined;
    group.value = null;
    emit('getUsers', q.value, filter.value, group.value);
};

onMounted(() => {
    getGroups();
});
</script>

<template>
    <div class="mb-3">
        <v-row align="center" class="px-md-15 mb-3 text-capitalize" justify="start">
            <v-col cols="12" md="7" sm="5">
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

            <v-col cols="12" md="2" sm="4">
                <v-select
                    v-model="group"
                    :items="groups"
                    :label="t('GROUP')"
                    :placeholder="t('GROUP_FILTER_ALL')"
                    clearable
                    color="primary"
                    density="compact"
                    hide-details
                    variant="outlined"
                />
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
</template>

<style scoped lang="scss"></style>

// Usage // function updateMeta(newMeta: Meta) { // Object.assign(meta, newMeta); // getUsers(); // } //
<!--<Pagination @update="updateMeta" :totalRecords="meta.total_records" />-->

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { OcservUsersGetSortEnum } from '@/api';

const emit = defineEmits(['update']);
const { t } = useI18n();

const props = defineProps({
    totalRecords: {
        type: Number,
        required: true
    }
});

const route = useRoute();
const router = useRouter();

const page = computed({
    get: () => Number(route.query.page) || 1,
    set: (v) => updateQuery({ page: v })
});
const size = computed({
    get: () => Number(route.query.size) || 10,
    set: (v) => updateQuery({ size: v })
});
const sort = computed({
    get: () =>
        route.query.sort && route.query.sort === 'DESC' ? OcservUsersGetSortEnum.DESC : OcservUsersGetSortEnum.ASC,
    set: (v: OcservUsersGetSortEnum) => updateQuery({ sort: v })
});

const pages = computed(() => Math.ceil(props.totalRecords / size.value));

const totalRecordsItems = computed(() => {
    const standardSizes = [5, 10, 25, 50, 100];
    const total = props.totalRecords ?? 0;
    const items: number[] = [];

    for (let i = 0; i < standardSizes.length; i++) {
        const size = standardSizes[i];
        if (size <= total) {
            items.push(size);
        } else {
            if (!items.includes(size)) items.push(size);
            break;
        }
    }

    if (!items.length && total > 0) items.push(total);
    return items;
});

onMounted(() => {
    const q = route.query;
    const page = Number(q.page || 1);
    const size = Math.max(Number(q.size) || 10, 5);
    const sort = q.sort && q.sort === 'DESC' ? OcservUsersGetSortEnum.DESC : OcservUsersGetSortEnum.ASC;
    router.replace({
        query: {
            ...q,
            page,
            size,
            sort
        }
    });

    updateQuery({ page, size, sort });
});

const updateQuery = (params: { page?: number; size?: number; sort?: string }) => {
    const newQuery = {
        ...route.query,
        ...params
    };

    if (newQuery.size && newQuery.size > props.totalRecords) {
        newQuery.page = 1;
    }

    router.replace({ query: newQuery });
    const meta = {
        page: newQuery.page,
        size: newQuery.size,
        sort: newQuery.sort
    };
    emit('update', meta);
};
</script>

<template>
    <v-row align="center" class="my-4" justify="center" v-if="totalRecords > 0">
        <v-col cols="12" lg="2" md="2">
            <v-select
                v-if="totalRecords > 5"
                v-model="size"
                :hint="`${t('TOTAL_RECORD')}: ${totalRecords}`"
                :items="totalRecordsItems"
                :label="t('ITEMS_PER_PAGE')"
                class="mt-md-6"
                color="primary"
                density="compact"
                persistent-hint
                variant="outlined"
            />
        </v-col>

        <v-col cols="12" lg="6" md="6">
            <v-pagination v-model="page" :length="pages" :total-visible="5" class="my-5" v-if="pages > 1" />
        </v-col>

        <v-col cols="12" lg="2" md="2">
            <v-select
                v-model="sort"
                :items="['ASC', 'DESC']"
                :label="t('SORT')"
                class="mt-md-6"
                color="primary"
                density="compact"
                persistent-hint
                variant="outlined"
            />
        </v-col>
    </v-row>
</template>

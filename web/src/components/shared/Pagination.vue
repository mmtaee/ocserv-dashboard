// Usage // function updateMeta(newMeta: Meta) { // Object.assign(meta, newMeta); // getUsers(); // } //
<!--<Pagination :meta="meta" @update="updateMeta" />-->

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { computed } from 'vue';
import type { Meta } from '@/types/metaTypes/MetaType';

const props = defineProps<{
    meta: Meta;
}>();

const emit = defineEmits(['update']);

const { t } = useI18n();

const pages = computed(() => {
    const total = props.meta.total_records ?? 0;
    const size = props.meta.size || 1;
    return Math.ceil(total / size);
});

const refresh = () => {
    emit('update', props.meta);
};

const totalRecordsItems = computed(() => {
    const options = [5, 10, 25, 50, 100];
    const total = props.meta.total_records ?? 0;

    return options.filter((size) => size <= total).length ? options.filter((size) => size <= total) : [total];
});
</script>

<template>
    <v-row align="center" class="my-4" justify="center">
        <v-col cols="12" lg="2" md="2">
            <v-select
                v-if="meta.total_records > 5"
                v-model="meta.size"
                :hint="`${t('TOTAL_RECORD')}: ${meta.total_records}`"
                :items="totalRecordsItems"
                :label="t('ITEMS_PER_PAGE')"
                class="mt-md-6"
                color="primary"
                density="compact"
                persistent-hint
                variant="outlined"
                @update:modelValue="refresh"
            />
        </v-col>

        <v-col cols="12" lg="6" md="6">
            <v-pagination
                v-model="meta.page"
                :length="pages"
                :total-visible="5"
                class="my-5"
                @update:modelValue="refresh"
            />
        </v-col>

        <v-col cols="12" lg="2" md="2">
            <v-select
                v-model="meta.sort"
                :items="['ASC', 'DESC']"
                :label="t('SORT')"
                class="mt-md-6"
                color="primary"
                density="compact"
                persistent-hint
                variant="outlined"
                @update:modelValue="refresh"
            />
        </v-col>
    </v-row>
</template>

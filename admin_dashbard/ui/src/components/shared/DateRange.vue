<script lang="ts" setup>
import { computed, ref, watch } from 'vue';
import { formatDate } from '@/utils/convertors';
import { useI18n } from 'vue-i18n';

interface InitDate {
    dateStart: Date;
    dateEnd: Date;
}

const props = defineProps({
    loading: {
        type: Boolean,
        default: false
    },
    initDate: {
        type: Object as () => InitDate,
        required: false
    },
    disableMore30Days: {
        type: Boolean,
        default: true
    },
    preHook: {
        type: Boolean,
        default: true
    }
});

const emit = defineEmits(['search']);

const { t } = useI18n();

const dateStart = ref<Date | null>(null);
const dateEnd = ref<Date | null>(null);

const showDateStartMenu = ref(false);
const showDateEndMenu = ref(false);

const searchDisable = computed(() => {
    // if both empty → disable
    if (!dateStart.value && !dateEnd.value) return true;

    // if only start is filled → valid
    if (dateStart.value && !dateEnd.value) return false;

    // if only end is filled → valid
    if (!dateStart.value && dateEnd.value) return false;

    if (dateStart.value && dateEnd.value && props.disableMore30Days) {
        const start = new Date(dateStart.value);
        const end = new Date(dateEnd.value);

        // disable if start >= end
        if (start.getTime() >= end.getTime()) return true;

        // calculate difference in days
        const diffMs = end.getTime() - start.getTime();
        const diffDays = diffMs / (1000 * 60 * 60 * 24);

        // disable if difference is more than 31 days (~1 month)
        return diffDays > 31;
    }
});

const search = () => {
    emit('search', formatDate(dateStart.value ?? ''), formatDate(dateEnd.value ?? ''));
};

watch(
    () => props.initDate,
    (val) => {
        if (val?.dateStart) {
            dateStart.value = val.dateStart;
        }
        if (val?.dateEnd) {
            dateEnd.value = val.dateEnd;
        }
        if (props.preHook) {
            search();
        }
    },
    { immediate: true, deep: true }
);
</script>

<template>
    <v-row align="center" justify="start">
        <v-col cols="12" lg="4" md="6">
            <v-menu v-model="showDateStartMenu" :close-on-content-click="false" transition="scale-transition">
                <template #activator="{ props }">
                    <v-label class="font-weight-bold mb-1 text-capitalize">
                        {{ t('DATE_START') }}
                    </v-label>
                    <v-text-field
                        :model-value="dateStart ? formatDate(dateStart) : ''"
                        color="primary"
                        hide-details
                        readonly
                        v-bind="props"
                        variant="outlined"
                    />
                </template>
                <v-date-picker
                    v-model="dateStart"
                    :header="t('DATE_START')"
                    :max="new Date()"
                    elevation="24"
                    title=""
                    @update:model-value="() => (showDateStartMenu = false)"
                />
            </v-menu>
        </v-col>

        <v-col cols="12" lg="4" md="6">
            <v-menu v-model="showDateEndMenu" :close-on-content-click="false" transition="scale-transition">
                <template #activator="{ props }">
                    <v-label class="font-weight-bold mb-1 text-capitalize">
                        {{ t('DATE_END') }}
                    </v-label>
                    <v-text-field
                        :model-value="dateEnd ? formatDate(dateEnd) : ''"
                        color="primary"
                        hide-details
                        readonly
                        v-bind="props"
                        variant="outlined"
                    />
                </template>
                <v-date-picker
                    v-model="dateEnd"
                    :header="t('DATE_END')"
                    :max="new Date()"
                    elevation="24"
                    title=""
                    @update:model-value="() => (showDateEndMenu = false)"
                />
            </v-menu>
        </v-col>

        <v-col class="mt-md-7" cols="auto">
            <v-btn :disabled="searchDisable" :loading="loading" color="primary" size="large" @click="search">
                {{ t('SEARCH') }}
            </v-btn>
        </v-col>
    </v-row>
</template>

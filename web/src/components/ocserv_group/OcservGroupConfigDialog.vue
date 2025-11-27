<script setup lang="ts">
import OcservGroupConfigDetail from '@/components/ocserv_group/OcservGroupConfigDetail.vue';
import { type PropType, ref, watch } from 'vue';
import type { GroupUnsyncedGroup, ModelsOcservGroupConfig } from '@/api';
import { useI18n } from 'vue-i18n';

const props = defineProps({
    show: {
        type: Boolean,
        default: false
    },
    group: {
        type: Object as PropType<GroupUnsyncedGroup>,
        required: true
    }
});

const emits = defineEmits(['close']);

const { t } = useI18n();

const resultArrayObj = ref<ModelsOcservGroupConfig>({});
const resultOther = ref<ModelsOcservGroupConfig>({});
const configArrayKeys = ['route', 'no-route', 'dns', 'split-dns'];

watch(
    () => props.group,
    (newGroup) => {
        console.log('newGroup: ', newGroup);
        resultArrayObj.value = Object.entries(newGroup?.config || {})
            .filter(([key]) => configArrayKeys.includes(key))
            .reduce<ModelsOcservGroupConfig>((acc, [key, val]) => {
                (acc as any)[key] = val;
                return acc;
            }, {} as ModelsOcservGroupConfig);

        resultOther.value = Object.entries(newGroup?.config || {})
            .filter(([key]) => !configArrayKeys.includes(key))
            .reduce<ModelsOcservGroupConfig>((acc, [key, val]) => {
                (acc as any)[key] = val;
                return acc;
            }, {} as ModelsOcservGroupConfig);
    },
    { immediate: false, deep: true }
);
</script>

<template>
    <v-dialog v-model="props.show" max-width="1200">
        <v-card>
            <v-card-title class="bg-primary text-capitalize">
                {{ group?.name }}
            </v-card-title>

            <v-card-text>
                <OcservGroupConfigDetail :resultArrayObj="resultArrayObj" :resultOther="resultOther" />
            </v-card-text>

            <v-card-actions class="mx-2 my-1">
                <v-spacer></v-spacer>

                <v-btn color="grey" variant="tonal" @click="emits('close')">
                    {{ t('CLOSE') }}
                </v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>
</template>

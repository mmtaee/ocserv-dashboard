<script lang="ts" setup>
import { onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import UiParentCard from '@/components/shared/UiParentCard.vue';
import UiChildCard from '@/components/shared/UiChildCard.vue';
import OcservGroupSync from '@/components/ocserv/sync/ocserv_group/OcservGroupSync.vue';
import OcservUserSync from '@/components/ocserv/sync/ocserv_user/OcservUserSync.vue';
import { useRoute, useRouter } from 'vue-router';

const { t } = useI18n();
const selectedTab = ref(0);
const route = useRoute();
const router = useRouter();
const tabs = [
    { title: t('GROUPS'), component: OcservGroupSync, color: 'purple' },
    { title: t('USERS'), component: OcservUserSync, color: 'orange' }
];

watch(selectedTab, (newIndex) => {
    router.replace({
        query: { ...route.query, tab: newIndex.toString() }
    });
});

onMounted(() => {
    const tabParam = route.query.tab;
    const index = parseInt(tabParam as string, 10);
    if (!isNaN(index) && index >= 0 && index < tabs.length) {
        selectedTab.value = index;
    }
});
</script>

<template>
    <UiParentCard variant="flat">
        <!-- Tabs header -->
        <v-tabs v-model="selectedTab" color="primary" align-tabs="center">
            <v-tab v-for="(tab, index) in tabs" :key="index">
                {{ tab.title }}
            </v-tab>
        </v-tabs>

        <v-divider class="my-2"></v-divider>

        <!-- Tabs content using v-window -->
        <v-window v-model="selectedTab">
            <v-window-item v-for="(tab, index) in tabs" :key="index" :value="index">
                <v-sheet class="pa-5" :color="tab.color">
                    <component :is="tab.component" />
                </v-sheet>
            </v-window-item>
        </v-window>
    </UiParentCard>
</template>

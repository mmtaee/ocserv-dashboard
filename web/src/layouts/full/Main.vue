<script lang="ts" setup>
import { computed, onMounted, ref, watch } from 'vue';
import { useDisplay } from 'vuetify';
import NavGroup from './vertical-sidebar/NavGroup/index.vue';
import NavItem from './vertical-sidebar/NavItem/index.vue';
import ExtraBox from './vertical-sidebar/extrabox/ExtraBox.vue';
import ProfileDD from './vertical-header/ProfileDD.vue';
import NavCollapse from './vertical-sidebar/NavCollapse/NavCollapse.vue';
import LanguageDD from '@/layouts/full/vertical-header/LanguageDD.vue';
import ThemeDD from '@/layouts/full/vertical-header/ThemeDD.vue';
import logoUrl from '@/assets/images/logo-circule.png';
import { getSidebarItems } from '@/layouts/full/vertical-sidebar/sidebarItem';
import { useServerStore } from '@/stores/config';
import { useThemeStore } from '@/stores/theme';

const sidebarMenu = getSidebarItems();
const { mdAndDown, smAndDown } = useDisplay();
const serverStore = useServerStore();
const themeStore = useThemeStore();

themeStore.sync();

const sDrawer = ref(true);

const serverInfo = computed(() => serverStore.getOcservVersion.split(', ').filter(Boolean));

const release = computed(() => serverStore.getDashboardRelease);

onMounted(() => {
    sDrawer.value = !mdAndDown.value;
});

watch(mdAndDown, (val) => {
    sDrawer.value = !val;
});
</script>

<template>
    <v-navigation-drawer
        v-model="sDrawer"
        :width="280"
        app
        class="leftSidebar"
        color="surface"
        elevation="0"
        left
    >
        <div class="sidebar-brand d-flex align-center px-5 py-4">
            <v-img :src="logoUrl" alt="logo" class="me-3" max-width="36" />
            <div class="d-flex flex-column">
                <span class="text-subtitle-1 font-weight-bold text-primary">Ocserv Dashboard</span>
                <span v-if="release.Current" class="text-caption text-medium-emphasis">
                    {{ release.Current }}
                </span>
            </div>
            <v-spacer />
            <v-btn
                v-if="mdAndDown"
                aria-label="close menu"
                icon
                size="small"
                variant="text"
                @click="sDrawer = false"
            >
                <v-icon size="22">mdi-close</v-icon>
            </v-btn>
        </div>
        <v-divider />
        <perfect-scrollbar class="scrollnavbar">
            <v-list class="pa-4" density="comfortable" nav>
                <template v-for="(item, i) in sidebarMenu" :key="`nav-${i}`">
                    <NavGroup v-if="item.header" :item="item" />
                    <NavCollapse v-else-if="item.children" :item="item" :level="0" class="leftPadding" />
                    <NavItem v-else :item="item" class="leftPadding" />
                </template>
            </v-list>
            <div class="px-4 pb-4">
                <ExtraBox />
            </div>
        </perfect-scrollbar>
    </v-navigation-drawer>

    <v-app-bar
        class="top-header"
        color="surface"
        elevation="1"
        flat
        height="64"
    >
        <v-btn
            aria-label="toggle menu"
            class="ms-2"
            icon
            size="small"
            variant="text"
            @click="sDrawer = !sDrawer"
        >
            <v-icon size="24">mdi-menu</v-icon>
        </v-btn>

        <div class="d-flex align-center ms-2">
            <v-img :src="logoUrl" alt="logo" max-width="32" />
            <span class="ms-2 text-subtitle-1 font-weight-bold text-primary d-none d-sm-inline">
                Ocserv Dashboard
                <span v-if="release.Current" class="text-caption text-medium-emphasis"> ({{ release.Current }})</span>
            </span>
        </div>

        <v-spacer />

        <div v-if="!smAndDown && serverInfo.length" class="server-info text-caption text-medium-emphasis me-3">
            <span v-for="(line, idx) in serverInfo" :key="idx" class="d-block">
                {{ line }}
            </span>
        </div>

        <ThemeDD />
        <LanguageDD />
        <ProfileDD />
    </v-app-bar>
</template>

<style lang="scss" scoped>
.top-header {
    border-bottom: 1px solid rgb(var(--v-theme-borderColor));
}

.sidebar-brand {
    min-height: 64px;
}

.server-info {
    line-height: 1.2;
    text-align: end;
    max-width: 260px;
    overflow: hidden;
    text-overflow: ellipsis;
}
</style>

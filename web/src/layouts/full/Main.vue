<script lang="ts" setup>
import { onMounted, ref, watch } from 'vue';
import { useDisplay } from 'vuetify';
import NavGroup from './vertical-sidebar/NavGroup/index.vue';
import NavItem from './vertical-sidebar/NavItem/index.vue';
import ExtraBox from './vertical-sidebar/extrabox/ExtraBox.vue';
import ProfileDD from './vertical-header/ProfileDD.vue';
import NavCollapse from './vertical-sidebar/NavCollapse/NavCollapse.vue';
import logoUrl from '@/assets/images/logo-circule.png';
import { getSidebarItems } from '@/layouts/full/vertical-sidebar/sidebarItem';
import LanguageDD from '@/layouts/full/vertical-header/LanguageDD.vue';

const sidebarMenu = getSidebarItems();

const { mdAndDown } = useDisplay();
const sDrawer = ref(true);

onMounted(() => {
    sDrawer.value = !mdAndDown.value;
});

watch(mdAndDown, (val) => {
    sDrawer.value = !val;
});
</script>

<template>
    <v-navigation-drawer v-model="sDrawer" :width="300" app class="leftSidebar" elevation="0" left>
        <div class="py-3 bg-primary text-h5">
            <span class="mx-5">Ocserv Dashboard</span>
            <v-btn class="mx-7" icon size="small" variant="text" @click="sDrawer = !sDrawer">
                <v-icon size="25" end>mdi-chevron-left</v-icon>
            </v-btn>
        </div>
        <perfect-scrollbar class="scrollnavbar">
            <v-list class="pa-6">
                <template v-for="(item, i) in sidebarMenu">
                    <NavGroup v-if="item.header" :key="item.title" :item="item" />
                    <NavCollapse v-else-if="item.children" :item="item" :level="0" class="leftPadding" />
                    <NavItem v-else :item="item" class="leftPadding" />
                </template>
            </v-list>
            <div class="pa-4">
                <ExtraBox />
            </div>
        </perfect-scrollbar>
    </v-navigation-drawer>
    <v-app-bar class="top-header bg-primary" elevation="0" height="64">
        <div v-if="!sDrawer">
            <v-img :src="logoUrl" class="hidden-md-and-up" width="50px" />
            <v-row align="center" class="hidden-md-and-down ms-3" justify="center">
                <v-col class="ma-0 pa-0" cols="12" md="6">
                    <v-img :src="logoUrl" width="50px" />
                </v-col>
                <v-divider class="me-1 ms-2" vertical />
                <v-col class="ma-0 pa-0" cols="12" md="3">
                    <v-btn class="mx-1" icon size="small" variant="text" @click="sDrawer = !sDrawer">
                        <v-icon size="25">mdi-menu</v-icon>
                    </v-btn>
                </v-col>
            </v-row>
        </div>
        <div v-else class="hidden-md-and-down" style="margin-left: 310px !important; margin-right: 310px !important">
            <v-img :src="logoUrl" width="50px" />
        </div>
        <div class="d-flex align-center justify-space-between w-100">
            <div>
                <v-btn class="hidden-lg-and-up" icon size="small" variant="text" @click="sDrawer = !sDrawer">
                    <v-icon size="20">mdi-menu</v-icon>
                </v-btn>
            </div>
            <div>
                <LanguageDD />
                <ProfileDD />
            </div>
        </div>
    </v-app-bar>
</template>

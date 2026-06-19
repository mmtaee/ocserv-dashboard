<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { ref } from 'vue'
const { t } = useI18n()
const drawer = ref(true)
</script>

<template>
  <v-app>
    <v-app-bar color="primary" dark>
      <v-app-bar-nav-icon @click="drawer = !drawer"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ t('dashboard') }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn color="error">{{ t('logout') }}</v-btn>
    </v-app-bar>

    <v-navigation-drawer v-model="drawer" app>
      <v-list>
        <v-list-item
          v-for="item in [
            { title: t('dashboard'), icon: 'mdi-view-dashboard', to: '/' },
            { title: t('profile'), icon: 'mdi-account', to: '/profile' },
            { title: t('usage'), icon: 'mdi-chart-bar', to: '/usage' },
          ]"
          :key="item.title"
          :to="item.to"
          :prepend-icon="item.icon"
          :title="item.title"
        ></v-list-item>
      </v-list>
    </v-navigation-drawer>

    <v-main>
      <router-view />
    </v-main>
  </v-app>
</template>

<style scoped></style>

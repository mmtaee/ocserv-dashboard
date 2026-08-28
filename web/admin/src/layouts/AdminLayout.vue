<script setup lang="ts">
import { ref, watch } from 'vue'
import { useDisplay } from 'vuetify'

import AdminNavbar from '@/components/AdminNavbar.vue'
import AdminSidebar from '@/components/AdminSidebar.vue'

const { mobile } = useDisplay()
const navigationOpen = ref(!mobile.value)

watch(mobile, (isMobile) => {
  navigationOpen.value = !isMobile
})
</script>

<template>
  <v-layout class="admin-layout">
    <AdminSidebar v-model="navigationOpen" :temporary="mobile" />
    <AdminNavbar @toggle-navigation="navigationOpen = !navigationOpen" />
    <v-main>
      <v-container>
        <router-view />
      </v-container>
    </v-main>
  </v-layout>
</template>

<style scoped>
.admin-layout {
  min-height: 100vh;
}
</style>

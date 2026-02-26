<script lang="ts" setup>
import { RouterView } from 'vue-router';
import MainView from './Main.vue';
</script>

<template>
    <v-locale-provider>
        <v-app>
            <MainView />
            <v-main>
                <v-container class="page-wrapper" fluid style="background-color: #eeee">
                    <div class="maxWidth">
                        <RouterView />
                    </div>
                </v-container>
            </v-main>
        </v-app>
    </v-locale-provider>
</template>

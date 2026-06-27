<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { router } from '@/router';

const { t } = useI18n();
</script>
<template>
    <div class="authentication">
        <v-container class="pa-3" fluid>
            <v-row class="h-100vh d-flex justify-center align-center">
                <v-col class="d-flex align-center" cols="12" lg="4" xl="3">
                    <v-card class="pa-10">
                        <v-card-title class="text-capitalize">
                            {{ t('LOGIN_WELCOME_NOTE') }}
                        </v-card-title>

                        <v-card-subtitle class="mb-8 text-capitalize text-subtitle-1">
                            {{ t('LOGIN_SELECT_ROUTE_HELP_TEXT') }}
                        </v-card-subtitle>

                        <v-col class="pt-0 mt-5" cols="12">
                            <v-btn block color="primary" flat size="large" @click="router.push({ name: 'Admin Login' })">
                                {{ t('ADMIN_LOGIN_BTN_ROUTE') }}
                            </v-btn>
                        </v-col>

                        <v-col class="pt-0 mt-2" cols="12">
                            <v-btn
                                block
                                color="muted"
                                flat
                                size="large"
                                @click="router.push({ name: 'Customer Summary' })"
                            >
                                {{ t('CUSTOMERS_SUMMARY_BTN_ROUTE') }}
                            </v-btn>
                        </v-col>
                    </v-card>
                </v-col>
            </v-row>
        </v-container>
    </div>
</template>

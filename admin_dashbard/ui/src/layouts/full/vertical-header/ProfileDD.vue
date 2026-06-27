<script lang="ts" setup>
import { useProfileStore } from '@/stores/profile';
import { useI18n } from 'vue-i18n';
import { ref } from 'vue';
import UiChildCard from '@/components/shared/UiChildCard.vue';
import { router } from '@/router';

const profileStore = useProfileStore();

const { t } = useI18n();

const logoutDialog = ref(false);

const logout = () => {
    profileStore.clearProfile();
    localStorage.removeItem('token');
    logoutDialog.value = false;
    router.push({ name: 'Admin Login' });
};
</script>

<template>
    <v-menu :close-on-content-click="true" dark>
        <template v-slot:activator="{ props }">
            <v-btn class="profileBtn custom-hover-primary" icon v-bind="props" variant="text">
                <v-avatar size="35">
                    <img alt="user" height="35" src="@/assets/images/avatar.jpg" />
                </v-avatar>
            </v-btn>
        </template>

        <v-sheet class="mt-2" elevation="10" rounded="md" width="200">
            <v-list class="py-0" density="compact" lines="one">
                <v-list-item-title class="text-uppercase px-5 pt-3 text-textPrimary" value="username">
                    {{ profileStore.profile?.username }}
                    <span class="text-subtitle-2 text-capitalize px-1">
                        ({{ profileStore.is_admin ? t('ADMIN') : t('STAFF') }})
                    </span>
                </v-list-item-title>

                <hr class="my-3" />

                <v-list-item class="mt-2" color="muted" to="/profile" value="change password">
                    <v-list-item-title class="text-body-1">
                        <v-icon color="primary" start> mdi-account-outline</v-icon>
                        <span class="ml-2 text-primary text-capitalize">
                            {{ t('PROFILE') }}
                        </span>
                    </v-list-item-title>
                </v-list-item>
            </v-list>
            <div class="pt-4 pb-4 px-5 text-center">
                <v-btn block color="error" variant="outlined" @click="logoutDialog = true">{{ t('LOGOUT') }}</v-btn>
            </div>
        </v-sheet>
    </v-menu>

    <!--  logout action dialog  -->
    <v-dialog v-model="logoutDialog" width="450">
        <UiChildCard divider titleColor="error">
            <template #title-header>
                <v-icon start>mdi-logout</v-icon>
                <span class="text-capitalize">
                    {{ t('LOGOUT_DIALOG_TITLE') }}
                </span>
            </template>

            <div class="text-capitalize">{{ t('LOGOUT_TEXT') }}?</div>

            <template #action>
                <div>
                    <v-btn color="muted" variant="text" @click="logoutDialog = false">
                        {{ t('CANCEL') }}
                    </v-btn>
                    <v-btn class="ms-2 me-1" color="error" variant="flat" @click="logout">
                        {{ t('LOGOUT') }}
                    </v-btn>
                </div>
            </template>
        </UiChildCard>
    </v-dialog>
</template>

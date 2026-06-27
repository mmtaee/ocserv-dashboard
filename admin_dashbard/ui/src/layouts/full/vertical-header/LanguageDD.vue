<script lang="ts" setup>
import { applyLocaleDirection, languages } from '@/plugins/i18n';
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';

const { locale } = useI18n();
const currentLang = ref(locale.value);

function changeLanguage(lang: string) {
    currentLang.value = lang;
    localStorage.setItem('language', currentLang.value);
    locale.value = lang;
    applyLocaleDirection(lang);
    window.location.reload();
}
</script>

<template>
    <v-menu :close-on-content-click="true" dark>
        <template v-slot:activator="{ props }">
            <v-btn v-bind="props" variant="plain">
                {{ currentLang }}
                <v-icon end> mdi-translate</v-icon>
            </v-btn>
        </template>

        <v-sheet class="mt-3 pb-3" elevation="10" rounded="md">
            <v-list class="py-0 text-center" density="compact" lines="one">
                <v-list-item
                    v-for="item in languages"
                    :key="item.code"
                    :value="item.code"
                    class="text-uppercase px-5 pt-3 text-textPrimary cursor-pointer"
                    @click="changeLanguage(item.code)"
                >
                    <span :class="item.code === currentLang ? 'text-primary' : ''">
                        {{ item.label }} - {{ item.code }}
                    </span>
                </v-list-item>
            </v-list>
        </v-sheet>
    </v-menu>
</template>

<style lang="scss" scoped></style>

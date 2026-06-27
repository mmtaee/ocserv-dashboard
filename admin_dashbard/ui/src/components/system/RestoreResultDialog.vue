<script setup lang="ts">
import type { PropType } from 'vue';
import { useI18n } from 'vue-i18n';
import UiChildCard from '@/components/shared/UiChildCard.vue';

const { t } = useI18n();

const props = defineProps({
    show: {
        type: Boolean,
        default: false
    },
    title: {
        type: String as PropType<'users' | 'groups'>,
        default: 'users'
    },
    inserted: {
        type: Array as PropType<string[]>,
        required: true
    },
    existing: {
        type: Array as PropType<string[]>,
        required: true
    }
});

const emits = defineEmits(['close']);

const titleTranslated = props.title == 'users' ? t('USERS') : t('GROUPS');
</script>

<template>
    <v-dialog v-model="props.show" max-width="1200">
        <v-card>
            <v-card-title class="bg-primary">
                <v-row align="end" justify="space-between" class="no-gutters">
                    <v-col md="auto"> {{ t('RESTORE_RESULT') }} ({{ titleTranslated }}) </v-col>
                    <v-col md="auto">
                        <v-icon @click="emits('close')">mdi-close</v-icon>
                    </v-col>
                </v-row>
            </v-card-title>

            <UiChildCard>
                <template v-slot:title-header> {{ t('ALREADY_EXISTS') }}: </template>

                <v-row align="center" justify="start">
                    <v-col md="auto" v-for="(username, index) in existing" :key="`existing-users-${index}`">
                        <v-icon color="secondary" start>mdi-account-network</v-icon> {{ username }} <br />
                    </v-col>
                </v-row>
            </UiChildCard>

            <UiChildCard>
                <template v-slot:title-header> {{ t('INSERTED') }}: </template>

                <v-row align="center" justify="start">
                    <v-col md="auto" v-for="(username, index) in inserted" :key="`inserted-users-${index}`">
                        <v-icon color="secondary" start>mdi-account-network</v-icon> {{ username }} <br />
                    </v-col>
                </v-row>
            </UiChildCard>

            <v-card-actions class="mx-2 my-1">
                <v-spacer />
                <v-btn color="grey" variant="tonal" @click="emits('close')">
                    {{ t('CLOSE') }}
                </v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';

const props = defineProps({
    show: {
        type: Boolean,
        default: false
    },
    usernames: {
        type: Array as () => string[]
    }
});

const emits = defineEmits(['close']);

const { t } = useI18n();
</script>

<template>
    <v-dialog v-model="props.show" max-width="450">
        <v-card>
            <v-card-title class="bg-primary text-capitalize">
                {{ t('OCSERV_USER_SYNC_DB_RESULT_TITLE_DIALOG') }}
            </v-card-title>

            <v-card-text>
                <span v-for="(username, index) in usernames" :key="`users-${index}`">
                    <v-icon color="secondary" start>mdi-database-check-outline</v-icon> {{ username }} <br />
                </span>
            </v-card-text>

            <v-card-actions class="mx-2 my-1">
                <v-spacer></v-spacer>
                <v-btn color="grey" variant="tonal" @click="emits('close')">
                    {{ t('CLOSE') }}
                </v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>
</template>

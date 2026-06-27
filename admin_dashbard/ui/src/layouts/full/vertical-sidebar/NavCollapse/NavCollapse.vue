<script setup>
import NavItem from '../NavItem/index.vue';

defineProps({ item: Object, level: Number });
</script>

<template>
    <v-list-group no-action>
        <template v-slot:activator="{ props }">
            <v-list-item :value="item.title" class="mb-1" rounded v-bind="props">
                <template v-slot:prepend> </template>
                <v-list-item-title class="mr-auto">{{ item.title }}</v-list-item-title>
                <v-list-item-subtitle v-if="item.subCaption" class="text-caption mt-n1 hide-menu">
                    {{ item.subCaption }}
                </v-list-item-subtitle>
            </v-list-item>
        </template>
        <template v-for="(subitem, i) in item.children" v-if="item.children" :key="i">
            <NavCollapse v-if="subitem.children" :item="subitem" :level="level + 1" />
            <NavItem v-else :item="subitem" :level="level + 1"></NavItem>
        </template>
    </v-list-group>
</template>

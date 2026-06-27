import { defineStore } from 'pinia';

export const useLoadingStore = defineStore('loading', {
    state: () => ({
        loading: false
    }),
    actions: {
        show() {
            this.loading = true;
        },
        hide() {
            this.loading = false;
        }
    },
    getters: {
        isLoading: (state) => state.loading
    }
});

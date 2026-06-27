import { defineStore } from 'pinia';
import type { SnackbarItem } from '@/types/storeTypes/StoreSnackbarType';

export const useSnackbarStore = defineStore('snackbar', {
    state: () => ({
        snackbars: [] as SnackbarItem[]
    }),

    actions: {
        show(input: SnackbarItem | SnackbarItem[]) {
            const items = Array.isArray(input) ? input : [input];
            items.forEach(({ id, message, color, timeout }) => {
                let idx = Math.floor(Math.random() * 1000000);
                if (id) {
                    idx = id + idx;
                }
                this.snackbars.push({
                    id: idx,
                    message,
                    color: color || 'info',
                    timeout: timeout || 3000
                });
            });
        },

        remove(id: number) {
            this.snackbars = this.snackbars.filter((snack) => snack.id !== id);
        }
    }
});

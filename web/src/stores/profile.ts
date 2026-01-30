import { defineStore } from 'pinia';
import { type ModelsUser, UsersApi } from '@/api';
import { getAuthorization } from '@/utils/request';

export const useProfileStore = defineStore('profile', {
    state: (): ModelsUser => ({
        role: 'staff',
        uid: '',
        username: '',
        created_at: undefined,
        updated_at: undefined,
        last_login: ''
    }),

    actions: {
        async getProfile() {
            const api = new UsersApi();
            try {
                const res = await api.usersProfileGet(getAuthorization());
                if (res.data) {
                    this.setProfile(res.data);
                }
            } catch (error) {
                console.error('Failed to fetch user profile', error);
                this.clearProfile();
            }
        },

        setProfile(user: ModelsUser) {
            Object.assign(this, user);
        },

        clearProfile() {
            Object.assign(this, {
                _: 0,
                role: null,
                uid: '',
                username: '',
                created_at: undefined,
                updated_at: undefined,
                last_login: ''
            });
        }
    },
    getters: {
        profile(state): ModelsUser | null {
            return state;
        },

        isAdmin(state: ModelsUser) {
            return state.role == 'admin' || state.role == 'super-admin';
        }
    }
});

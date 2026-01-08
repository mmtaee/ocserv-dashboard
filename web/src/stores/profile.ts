import { defineStore } from 'pinia';
import { type ModelsUser, SystemUsersApi } from '@/api';
import { getAuthorization } from '@/utils/request';

export const useProfileStore = defineStore('profile', {
    state: (): ModelsUser => ({
        is_admin: false,
        uid: '',
        username: '',
        created_at: undefined,
        updated_at: undefined,
        last_login: ''
    }),

    actions: {
        async getProfile() {
            const api = new SystemUsersApi();
            try {
                const res = await api.systemUsersProfileGet(getAuthorization());
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
                is_admin: false,
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
        // TODO: should change to role base
        isAdmin(state: ModelsUser) {
            return state.is_admin;
        }
    }
});

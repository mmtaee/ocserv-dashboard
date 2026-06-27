import { defineStore } from 'pinia';
import { OCCTLApi, SystemApi } from '@/api';
import type { ConfigState, Release, ServerState } from '@/types/storeTypes/StoreConfigType';

export const useServerStore = defineStore('server', {
    state: (): ServerState => ({
        OcservVersion: '',
        OcctlVersion: '',
        Status: '',
        Release: {
            Current: '',
            Latest: ''
        }
    }),
    actions: {
        async fetchServerInfo() {
            const api = new OCCTLApi();
            await api
                .occtlServerInfoGet()
                .then((res) => {
                    if (res.data) {
                        this.OcservVersion = res.data.version.ocserv_version || '';
                        this.OcctlVersion = (res.data.version.occtl_version || '').replace(/\n/g, '<br />');
                    }
                })
                .catch(() => {});
        },
        async fetchDashboardVersion() {
            const api = new SystemApi();
            api.systemReleaseGet()
                .then((res) => {
                    this.Release = {
                        Current: res.data.current,
                        Latest: res.data.latest
                    };
                })
                .catch(() => {});
        },
        async setStatus(status: string) {
            this.Status = status;
        }
    },
    getters: {
        getOcservVersion: (state) => state.OcservVersion,
        getOcctlVersion: (state) => state.OcctlVersion,
        getStatus: (state) => state.Status,
        getDashboardRelease: (state) => state.Release
    }
});

export const useConfigStore = defineStore('config', {
    state: (): ConfigState => ({
        setup: false,
        googleCaptchaSiteKey: '',
        telegramBotEnabled: false
    }),

    actions: {
        async fetchConfig() {
            const api = new SystemApi();
            await api.systemInitGet().then((res) => {
                if (res.data) {
                    this.googleCaptchaSiteKey = res.data.google_captcha_site_key || '';
                    this.telegramBotEnabled = res.data.telegram_bot_enabled || false;
                    this.setup = true;
                }
            });
            return this.setup;
        },
        setConfig(googleCaptchaSiteKey: string | undefined) {
            if (googleCaptchaSiteKey) {
                this.googleCaptchaSiteKey = googleCaptchaSiteKey;
            }
            this.setup = true;
        }
    },
    getters: {
        config(state): ConfigState {
            return {
                setup: state.setup,
                googleCaptchaSiteKey: state.googleCaptchaSiteKey,
                telegramBotEnabled: state.telegramBotEnabled
            };
        }
    }
});

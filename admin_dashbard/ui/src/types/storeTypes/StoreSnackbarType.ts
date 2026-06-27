export type SnackbarColor = 'success' | 'error' | 'info' | 'warning';

export type SnackbarItem = {
    id?: number;
    message: string;
    color: SnackbarColor;
    timeout: number;
};

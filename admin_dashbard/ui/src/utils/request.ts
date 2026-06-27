import type { Authorization } from '@/types/requestTypes/RequestType';

function getAuthorization(): Authorization {
    const token = localStorage.getItem('token');
    return {
        authorization: `Bearer ${token ?? ''}`
    };
}

export { getAuthorization };

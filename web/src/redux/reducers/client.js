export const SAVE = 'SAVE_CLIENT';
export const CLEAR = 'CLEAR_CLIENT';

export default function client (state={}, action) {
    switch (action.type) {
        case SAVE:
            return { ...state, ...action.data };
        case CLEAR:
            return {};
        default:
            return state;
    }
}
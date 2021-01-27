export const SAVE = 'SAVE_ROLE';
export const CLEAR = 'CLEAR_ROLE';

export default function role (state={}, action) {
    switch (action.type) {
        case SAVE:
            return { ...state, ...action.data };
        case CLEAR:
            return {};
        default:
            return state;
    }
}
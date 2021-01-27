export const SAVE = 'SAVE_COMMON';

export default function common (state={}, action) {
    switch (action.type) {
        case SAVE:
            return { ...state, ...action.data };
        default:
            return state;
    }
}
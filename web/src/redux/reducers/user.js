export const SAVE = 'SAVE_USER';
export const CLEAR = 'CLEAR_USER';

export default function user (state={}, action) {
    switch (action.type) {
        case SAVE:
            return { ...state, ...action.data };
        case CLEAR:
            return {};
        default:
            return state;
    }
}

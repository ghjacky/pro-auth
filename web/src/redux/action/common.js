import { SAVE } from '../reducers/common';
import request from '../../utils/request';

export const loadingStart = (loadingKey = 'loading') => {
    let data = {};
    data[loadingKey] = true;
    return {type: SAVE, data};
};


export const loadingEnd = (loadingKey= 'loading') => {
    let data = {};
    data[loadingKey] = false;
    return {type: SAVE, data};
};


export const queryCurrentUser = (cb) => {
    return (dispatch) => {
        request(`/api/user`, { method: 'GET'})
            .then(res => {
                if (res.res_code === 0) {
                    dispatch({type: SAVE, data: { user :res.data }})
                    if (cb) {
                        cb()
                    }
                }
            });
    };
};

export const generateUserSecret = (cb) => {
    return (dispatch) => {
        request(`/api/user/secret/generate`, { method: 'POST'})
            .then(res => {
                if (res.res_code === 0) {
                    dispatch(queryCurrentUser(cb))
                }
            });
    };
};

export const queryUserList = (client_id)=> {
    return (dispatch) => {
        request(`/api/users/members?client_id=${client_id}`, { method: 'GET'})
            .then(res => {
                if (res.res_code === 0) {
                    dispatch({type: SAVE, data: { userList :res.data }})
                }
            });
    };
};

export const querySystemSetting = () => {
    return (dispatch) => {
        request(`/system/setting`, { method: 'GET'})
            .then(res => {
                if (res.res_code === 0) {
                    dispatch({type: SAVE, data: { setting :res.data }})
                }
            });
    };
};

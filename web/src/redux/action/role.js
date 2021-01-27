import request from "../../utils/request";
import {SAVE} from "../reducers/role";
import {loadingStart, loadingEnd} from './common';


export const queryUserRoles = (saveKey, client_id, role_type = 'normal', relate_resource = false, relate_user = false, is_all = false, is_route = false) => {
    return (dispatch) => {
        dispatch(loadingStart());
        request(`/api/userRoles?client_id=${client_id}&role_type=${role_type}&relate_resource=${relate_resource}&relate_user=${relate_user}&is_all=${is_all}&is_route=${is_route}`, {method: 'GET'})
            .then(res => {
                dispatch(loadingEnd());
                if (res.res_code === 0) {
                    const data = {};
                    data[saveKey] = res.data;
                    dispatch({type: SAVE, data: data})
                }
            });
    };
};


export const queryClientRoles = (saveKey, client_id, relate_resource = false, relate_user = false, is_tree = false) => {
    return (dispatch) => {
        dispatch(loadingStart());
        request(`/api/roles?client_id=${client_id}&relate_resource=${relate_resource}&relate_user=${relate_user}&is_tree=${is_tree}`, {method: 'GET'})
            .then(res => {
                dispatch(loadingEnd());
                if (res.res_code === 0) {
                    const data = {};
                    data[saveKey] = res.data;
                    dispatch({type: SAVE, data: data})
                }
            });
    }
};


export const  queryUserAndClientRoles = (saveKeys, client_id, relate_resource = false, relate_user = false, is_all = false, role_type = 'admin') => {
    return (dispatch) => {
        dispatch(loadingStart());
        const urls = [
            `/api/userRoles?client_id=${client_id}&role_type=${role_type}&is_all=${is_all}`,
            `/api/roles?client_id=${client_id}&relate_resource=${relate_resource}&relate_user=${relate_user}`
        ];
        Promise.all(
            urls.map(url => request(url, {method: 'GET'}, false))
        ).then(resArray => {
            dispatch(loadingEnd());
            if (resArray[0].res_code === 0 && resArray[1].res_code === 0) {
                const data = {};
                data[saveKeys[0]] = resArray[0].data;
                data[saveKeys[1]] = resArray[1].data;
                dispatch({type: SAVE, data:data});

            }
        })
    }
};


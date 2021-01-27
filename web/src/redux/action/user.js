import request from "../../utils/request";
import { loadingStart, loadingEnd } from './common';
import {SAVE} from "../reducers/user";

export const queryUsers = (page, pageSize, id="") => {
    return (dispatch) => {
        dispatch(loadingStart());
        let url = `/api/users?page_size=${pageSize}&p=${page}`;
        if (id.length > 0) {
            url = `/api/users?page_size=${pageSize}&p=${page}&id=${id}`;
        }

        const urls = [ url ];
        Promise.all(
            urls.map(url => request(url, {method: 'GET'}, false))
        ).then(resArray => {
            dispatch(loadingEnd());
            if (resArray[0].res_code === 0) {
                dispatch({type: SAVE, data: {users: resArray[0].data.data, paginator: resArray[0].data.paginator}})
            }
        })
    }
};

export const queryUser = (user_id) => {
    return (dispatch) => {
        dispatch(loadingStart());
        request(`/api/user/${user_id}`, {
            method: 'GET'}, false)
            .then(res =>{
                dispatch(loadingEnd());
                if(res.res_code === 0) {
                    dispatch({type: SAVE, data: { user: res.data}})

                }
            })
    }
};


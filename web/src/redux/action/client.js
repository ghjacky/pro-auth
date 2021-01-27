import request from "../../utils/request";
import { loadingStart, loadingEnd } from './common';
import {SAVE} from "../reducers/client";


export const queryClients = () => {
    return (dispatch) => {
        dispatch(loadingStart());
        const urls = [
            '/api/userClients',
            '/api/allClients'
        ];
        Promise.all(
            urls.map(url => request(url, {method: 'GET'}, false))
        ).then(resArray => {
            dispatch(loadingEnd());
            if (resArray[0].res_code === 0 && resArray[1].res_code === 0) {
                dispatch({type: SAVE, data: {userClients: resArray[0].data, allClients: resArray[1].data}})

            }
        })
    }
};

export const queryClientAndResource = (client_id) => {
    return (dispatch) => {
        dispatch(loadingStart());
        const urls = [
            `/api/client?client_id=${client_id}`,
            `/api/resources?client_id=${client_id}&relate_role=true`
        ];
        Promise.all(
            urls.map(url => request(url, {method: 'GET'}, false))
        ).then(resArray => {
            dispatch(loadingEnd());
            if (resArray[0].res_code === 0 && resArray[1].res_code === 0) {
                dispatch({type: SAVE, data: { client: resArray[0].data, resources: resArray[1].data}})
            }
        })
    }
};

export const queryClient = (client_id) => {
    return (dispatch) => {
        dispatch(loadingStart());
        request(`/api/client?client_id=${client_id}`, {
            method: 'GET'}, false)
            .then(res =>{
                dispatch(loadingEnd());
                if(res.res_code === 0) {
                    dispatch({type: SAVE, data: { client: res.data}})

                }
            })
    }
};

export const queryResources = (client_id) => {
    return (dispatch) => {
        dispatch(loadingStart());
        request(`/api/resources?client_id=${client_id}&relate_role=true`, {
            method: 'GET'}, false)
            .then(res =>{
                dispatch(loadingEnd());
                if(res.res_code === 0) {
                    dispatch({type: SAVE, data: { resources: res.data}})

                }
            })
    }
};

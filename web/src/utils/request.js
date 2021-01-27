import fetch from 'isomorphic-fetch';
import { notification } from 'antd';

const codeMessage = {
    200: '服务器成功返回请求的数据。',
    201: '新建或修改数据成功。',
    202: '一个请求已经进入后台排队（异步任务）。',
    204: '删除数据成功。',
    400: '请求参数错误',
    401: '未登录',
    403: '无权限使用或访问该内容',
    404: '发出的请求针对的是不存在的记录，服务器没有进行操作。',
    406: '请求的格式不可得。',
    410: '请求的资源被永久删除，且不会再得到的。',
    422: '当创建一个对象时，发生一个验证错误。',
    500: '服务器发生错误，请检查服务器。',
    502: '网关错误。',
    503: '服务不可用，服务器暂时过载或维护。',
    504: '网关超时。',
};

function checkStatus(response, showErrMsg) {
    if (response.status >= 200 && response.status < 300) {
        return response;
    }
    const errortext = codeMessage[response.status] || response.statusText;
    if (response.status === 401) {
        window.location.reload();
    } else if (showErrMsg){
        notification.error({
            message: `请求错误 ${response.status}`,
            description: errortext,
        });
    }
    const error = new Error(errortext);
    error.status = response.status;
    error.response = response;
    throw error;
}
/**
 * Requests a URL, returning a promise.
 *
 * @param  {string} url       The URL we want to request
 * @param  {object} [options] The options we want to pass to "fetch"
 * @param  {boolean} showErrMsg  if show error message
 * @return {object}           An object containing either "data" or "err"
 */
export default async function request(url, options, showErrMsg = true) {
    const defaultOptions = {
        credentials: 'include',
    };
    const newOptions = { ...defaultOptions, ...options };
    if (newOptions.method === 'POST' || newOptions.method === 'PUT') {
        if (!(newOptions.body instanceof FormData)) {
            newOptions.headers = {
                Accept: 'application/json',
                'Content-Type': 'application/json; charset=utf-8',
                ...newOptions.headers,
            };
        }
    }
    let response;
    try {
        response = await fetch(url, newOptions);
    }catch (e) {
        response = {status: 504, url: url};
    }
    let ret;
    try {
        ret = checkStatus(response, showErrMsg);
        ret = await ret.json();
    } catch (e) {
        ret = { code: 500, message: e.message, data: null };
    }
    return ret;
}


export function requestPure(url, options) {
    const defaultOptions = {
        credentials: 'include',
    };
    const newOptions = { ...defaultOptions, ...options };
    return fetch(url, newOptions);
}



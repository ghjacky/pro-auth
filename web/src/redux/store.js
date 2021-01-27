import thunkMiddleware from 'redux-thunk';
import { createStore, combineReducers, applyMiddleware } from 'redux';
import { composeWithDevTools } from 'redux-devtools-extension';
import common from './reducers/common';
import client from './reducers/client';
import role from './reducers/role';
import user from './reducers/user';


const rootReducer = combineReducers({
    common,
    client,
    role,
    user,
});

const store = createStore(
    rootReducer,
    composeWithDevTools(applyMiddleware(thunkMiddleware)),
);

export default store;

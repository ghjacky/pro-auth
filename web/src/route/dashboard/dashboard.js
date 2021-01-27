import React from 'react';
import {Route, withRouter, Switch} from 'react-router-dom';
import { Layout, message } from 'antd';
import TopMenu from '../../components/dashboard/TopMenu';
import My from './my';
import Member from './member';
import Role from './role';
import Client from './client';
import './dashboard.css';
import {connect} from "react-redux";
import request from '../../utils/request';

class DashBoard extends React.Component {
    componentDidMount() {
        const {setHeader, match, history} = this.props;
        const client_id = match.params.client_id;
        const type = match.params.type;
        request(`/api/userRoles?client_id=${client_id}`, {
            method:'GET'
        },false).then(res => {
            if(res.res_code === 0) {
                const roles = res.data;
                let isRootRoleSuper = false;
                let isRoleSuperOrAdmin = false;

                for(let i =0 ; i < roles.length;i++) {
                    if(roles[i].role_type === 'super' && roles[i].parent_id === -1) {
                        isRootRoleSuper = true;
                        isRoleSuperOrAdmin = true;
                        break;
                    } else if(roles[i].role_type === 'super' || roles[i].role_type === 'admin') {
                        isRoleSuperOrAdmin = true;
                        break;
                    }
                }
                if(roles.length === 0) {
                    message.error("在该应用下没有角色");
                    history.push({pathname:'/frontend/'});
                    return;
                } else if ((type ==='user' || type === 'role') && !isRoleSuperOrAdmin) {
                    message.error("在该应用下没有管理员身份");
                    history.push({pathname:'/frontend/'});
                    return;
                } else if(type === 'client' && !isRootRoleSuper) {
                    message.error("不是该应用的超级管理员");
                    history.push({pathname:'/frontend/'});
                    return;
                }
                setHeader(<TopMenu isRootRoleSuper={isRootRoleSuper} isRoleSuperOrAdmin={isRoleSuperOrAdmin} clientId={client_id}/>)
            } else {
                message.error("获取角色失败");
                history.push({pathname:'/'});
                return;
            }
        });
    }

    componentWillUnmount() {
        const {setHeader} = this.props;
        setHeader(null);
        this.props.dispatch({type:'CLEAR_ROLE'});
        this.props.dispatch({type:'CLEAR_CLIENT'});
        this.props.dispatch({type:'CLEAR_RESOURCE'});
    }
    render() {
        return (<Layout style={{background: 'white'}}>
            <Switch>
                <Route exact path="/frontend/dashboard/my/:client_id" component={My}/>
                <Route exact path="/frontend/dashboard/user/:client_id" component={Member}/>
                <Route exact path="/frontend/dashboard/role/:client_id" component={Role}/>
                <Route exact path="/frontend/dashboard/client/:client_id" component={Client}/>
            </Switch>
        </Layout>);
    }
}

function mapStateToProps(state) {
    return {
        ...state.role,
        loading: state.common.loading
    };
}


function mapDispatchToProps(dispatch, ownProps) {
    return {
        dispatch: dispatch,
    }
}

export default withRouter(connect(mapStateToProps, mapDispatchToProps)(DashBoard));

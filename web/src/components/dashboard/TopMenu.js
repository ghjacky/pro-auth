import React from 'react';
import {Menu, Icon} from "antd";
import {Link, withRouter} from 'react-router-dom';
import './TopMenu.css';

class TopMenu extends React.Component {
    render() {
        const {location, clientId, isRootRoleSuper, isRoleSuperOrAdmin} = this.props;
        return (<Menu
            id="nav"
            mode="horizontal"
            selectedKeys={[location.pathname]}
        >
            <Menu.Item key='left'>
                <Link to="/"><Icon type="caret-left"/>返回</Link>
            </Menu.Item>
            <Menu.Item key={`/frontend/dashboard/my/${clientId}`}>
                <Link to={`/frontend/dashboard/my/${clientId}`}><Icon type="user"/>我的权限</Link>
            </Menu.Item>
            {
                isRoleSuperOrAdmin && <Menu.Item key={`/frontend/dashboard/user/${clientId}`}>
                    <Link to={`/frontend/dashboard/user/${clientId}`}><Icon type="team"/>成员管理</Link>
                </Menu.Item>
            }
            {
                isRoleSuperOrAdmin && <Menu.Item key={`/frontend/dashboard/role/${clientId}`}>
                    <Link to={`/frontend/dashboard/role/${clientId}`}><Icon type="branches"/>角色树管理</Link>
                </Menu.Item>
            }
            {
                isRootRoleSuper && <Menu.Item key={`/frontend/dashboard/client/${clientId}`}>
                    <Link to={`/frontend/dashboard/client/${clientId}`}><Icon type="setting"/>应用管理</Link>
                </Menu.Item>
            }
        </Menu>);
    }
}

export default withRouter(TopMenu);
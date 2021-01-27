import React from 'react';
import {Menu, Icon} from "antd";
import {Link, withRouter} from 'react-router-dom';
import './TopMenu.css';

class TopNav extends React.Component {
    render() {
        const {location, clientId, isRootRoleSuper, isRoleSuperOrAdmin} = this.props;
        return (<Menu
            id="nav"
            mode="horizontal"
            selectedKeys={[location.pathname]}
        >
            {
                isRoleSuperOrAdmin && <Menu.Item key={`/frontend/dashboard/user/${clientId}`}>
                    <Link to={`/frontend/dashboard/user/${clientId}`}><Icon type="team"/>用户管理</Link>
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

export default withRouter(TopNav);

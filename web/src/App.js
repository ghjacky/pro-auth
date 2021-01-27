import React, {Component} from 'react';
import './App.css';
import {Route, withRouter, Switch, Redirect, Link} from 'react-router-dom';
import {Layout, Menu, Icon, Dropdown, Button, Modal} from 'antd';
import Dashboard from './route/dashboard/dashboard';
import Client from './route/home/client'
import User from './route/home/user'
import {connect} from "react-redux";
import { queryCurrentUser, generateUserSecret } from "./redux/action/common";

const {Header, Footer} = Layout;

class App extends Component {
    state = {
        customHeader: null,
    };
    componentDidMount() {
        this.props.dispatch(queryCurrentUser())
    }

    handleLogout = () => {
        window.location.href = "/auth/logout";
    };
    setHeader = (newHeader) => {
        this.setState({customHeader: newHeader})
    };

    showModal = () => {
        Modal.info({
            title: 'Secret',
            content: (
              <div>
                <p>{this.props.user.secret}</p>
                <Button onClick={this.generateUserSecretHandle} type="primary" icon="redo">重新生成</Button>
              </div>
            ),
        })
    }

    showSecretModal = () => {
        if (this.props.user && this.props.user.secret) {
            this.showModal()
        } else {
            this.generateUserSecretHandle()
        }
    };

    generateUserSecretHandle = () => {
        Modal.confirm({
            title: "确认",
            content: "重新生成会擦除旧secret，确认操作？",
            onOk: () => {
                Modal.destroyAll();
                this.props.dispatch(generateUserSecret(this.showModal))
            },
        })
    };

    render() {
        const {location, user} = this.props;
        const { customHeader } = this.state;
        let isAdmin = false;
        let username = '未登录用户';
        if (user) {
            username = user.fullname;
            isAdmin = user.id === "admin";
        }
        const menu = (
            <Menu>
                <Menu.Item>
                    <Button style={{border: 'transparent'}} onClick={this.showSecretModal}>我的秘钥</Button>
                </Menu.Item>
                <Menu.Item>
                    <Button style={{border: 'transparent'}} onClick={this.handleLogout}>退出登录</Button>
                </Menu.Item>
            </Menu>
        );
        return (
            <Layout className="layout" style={{minHeight: '100vh'}}>
                <Header className="App-header">
                    <div className="App-header-left">
                        {/*<img alt={'logo'} className="App-logo" src="/static/logo.png" />*/}
                        { customHeader !== null && customHeader}
                    </div>
                    <div className="App-header-right">

                        <Menu id="nav" mode="horizontal" selectedKeys={[location.pathname]}>
                            <Menu.Item key={`/frontend/home/client`}>
                                <Link to={`/frontend/home/client`}><Icon type="setting"/>应用管理</Link>
                            </Menu.Item>
                            {
                                isAdmin && <Menu.Item key={`/frontend/home/user`}>
                                    <Link to={`/frontend/home/user`}><Icon type="team"/>用户管理</Link>
                                </Menu.Item>
                            }
                        </Menu>

                        <Dropdown overlay={menu} placement="bottomLeft" trigger={['click']}>
                            <Button style={{border: 'transparent', marginRight: 10}}><Icon type="user"/>
                                {username}</Button>
                        </Dropdown>
                    </div>
                </Header>
                <Layout className="App-content">
                    <Switch>
                        <Route exact path="/frontend/home/client" component={Client}/>
                        {isAdmin && <Route exact path="/frontend/home/user" component={User}/>}
                        <Route path="/frontend/dashboard/:type/:client_id" render={() =><Dashboard setHeader={this.setHeader}/>}/>
                        <Redirect from="/" to="/frontend/home/client" />
                    </Switch>
                </Layout>
                <Footer style={{textAlign: 'center'}}>
                    Auth ©{new Date().getFullYear()} xxxxx
                </Footer>
            </Layout>
        );
    }
}

function mapStateToProps(state) {
    return {
        user: state.common.user,
    };
}


function mapDispatchToProps(dispatch, ownProps) {
    return {
        dispatch: dispatch,
    }
}

export default withRouter(connect(mapStateToProps, mapDispatchToProps)(App));

import React from 'react';
import {Button, Card, Col, Form, Input, message, Modal, Row, Table, Tag} from 'antd';
import {connect} from 'react-redux';
import {withRouter} from 'react-router-dom'
import request from "../../utils/request";
import {queryUsers} from "../../redux/action/user";
import {userStatusConfig} from "../../utils/config";
import {querySystemSetting} from "../../redux/action/common";
const confirm = Modal.confirm;

const Search = Input.Search;

class User extends React.Component {
    state = {
        showForm: false,
        canSubmit: true,
        current: 1,
        pageSize: 10,
        searchText: '',
        paginator: {},
        users: [],
        usersOrigin: [],
        editUser: null,
        mode: "create",
        searchId: "",
        setting: {},
    };

    componentDidMount() {
        this.props.dispatch(queryUsers(this.state.current, this.state.pageSize, this.state.searchId));
        this.props.dispatch(querySystemSetting());
    }

    componentWillReceiveProps(nextProps) {
        const {users, paginator, setting} = nextProps;
        if (users) {
            this.setState({users: users, usersOrigin: users});
        }
        if (paginator) {
            this.setState({paginator: paginator});
        }
        if (setting) {
            this.setState({setting: setting});
        }
    }

    onSearch = (value) => {
        this.setState({searchId: value});
        this.props.dispatch(queryUsers(this.state.current, this.state.pageSize, value));
    };
    onOk = (form) => {
        form.validateFields((err, values) => {
            if (!err) {
                this.setState({canSubmit: false});
                let title, method, url;
                if (this.state.mode === "create") {
                    title = "创建用户";
                    method = "POST";
                    url = '/api/users';
                } else {
                    title = "编辑用户";
                    method = "PUT";
                    url = `/api/users/${values.id}`;
                }
                request(url,{
                    method: method,
                    body: JSON.stringify(values),
                }, false).then(res => {
                    this.setState({canSubmit: true});
                    if(res.res_code === 0) {
                        message.success(`成功${title}`);
                        this.setState({showForm: false});
                        this.props.dispatch(queryUsers(this.state.current, this.state.pageSize, this.state.searchId));
                    } else {
                        Modal.error({
                            title: `${title}失败`,
                            content: res.res_msg,
                        });
                    }
                })
            }
        });
    };

    onDeleteUser = (user)=> {
        confirm({
            title: `确定删除用户${user.id}?`,
            okText: '确定',
            okType: 'danger',
            cancelText: '取消',
            onOk() {
                request(`/api/users/${user.id}/status/delete`,{
                    method: 'PUT'
                }).then(res => {
                    if(res.res_code === 0) {
                        message.success('删除用户成功！');
                        this.props.dispatch(queryUsers(this.state.current, this.state.pageSize, this.state.searchId));
                    } else {
                        message.error(res.res_msg);
                    }
                })
            },
            onCancel() {
            },
        });
    };
    onCreateUser = (user) => {
        this.setState({showForm: true, mode: "create", editUser: null});
    };
    onCancel = (user) => {
        this.setState({showForm: false, editUser: null})
    };
    onEditUser = (user) => {
        this.setState({
            showForm: !this.state.showForm,
            mode: "edit",
            editUser: user
        });
    };
    onFrozenUser = (user) => {
        const that = this;
        confirm({
            title: `冻结用户后不能登录，确定冻结用户${user.id}?`,
            okText: '确定',
            okType: 'danger',
            cancelText: '取消',
            onOk() {
                request(`/api/users/${user.id}/status/frozen`,{
                    method: 'PUT'
                }).then(res => {
                    if(res.res_code === 0) {
                        message.success('冻结用户成功！');
                        that.props.dispatch(queryUsers(that.state.current, that.state.pageSize, this.state.searchId));
                    } else {
                        message.error(res.res_msg);
                    }
                })
            },
            onCancel() {
            },
        });
    };
    onActiveUser = (user) => {
        request(`/api/users/${user.id}/status/active`,{
            method: 'PUT'
        }).then(res => {
            if(res.res_code === 0) {
                message.success('激活用户成功！');
                this.props.dispatch(queryUsers(this.state.current, this.state.pageSize, this.state.searchId));
            } else {
                message.error(res.res_msg);
            }
        })
    };

    render() {
        const {current, pageSize, users, showForm, paginator, searchId, setting} = this.state;
        const {loading} = this.props;
        console.log(setting);
        let total = paginator.total_size;
        const paginationProps = {
            current: current,
            total: total,
            pageSize: pageSize,
            showSizeChanger: true,
            showTotal: (total, range) => `${range[0] > 0 ? range[0] : 0}-${range[1] > 0 ? range[1] : 0} of ${total} users`,
            onShowSizeChange: (current, pageSize) => {
                this.setState({current: 1, pageSize: pageSize});
                this.props.dispatch(queryUsers(current, pageSize, searchId));
            },
            onChange: (page, pageSize) => {
                this.setState({current: page, pageSize: pageSize});
                this.props.dispatch(queryUsers(page, pageSize, searchId));
            }
        };
        const columns = [
            {
                title: '账号',
                dataIndex: 'id',
                key: 'id',
            },
            {
                title: '状态',
                dataIndex: 'status',
                key: 'status',
                render: (record) => {
                    const status = userStatusConfig[record];
                    let color = "";
                    let name = "";
                    if (status != null) {
                        color = status.color;
                        name = status.name;
                    }
                    return <Tag color={color}>{name}</Tag>
                }

            },
            {
                title: '姓名',
                dataIndex: 'fullname',
                key: 'fullname',
            },
            {
                title: '邮箱',
                dataIndex: 'email',
                key: 'email',
            },
            {
                title: '电话',
                dataIndex: 'phone',
                key: 'phone',
            },
            {
                title: '微信',
                dataIndex: 'wechat',
                key: 'wechat',
            },
            {
                title: '类型',
                dataIndex: 'type',
                key: 'type',
            },
            {
                title: '组织',
                dataIndex: 'organization',
                key: 'organization',
            },
            {
                title: '操作',
                dataIndex: 'operation',
                key: 'operation',
                render: (text, record) => {
                    return (<div>
                        {record.type !== 'ldap' && record.id !== "admin" && <Button type="primary" ghost style={{margin: 6}} onClick={() => this.onEditUser(record)}>编辑</Button>}
                        {record.type !== 'ldap' && record.id !== "admin" && record.status === 'active' && <Button type="primary" ghost style={{margin: 6}} onClick={() => this.onFrozenUser(record)}>冻结</Button>}
                        {record.type !== 'ldap' && record.status === 'frozen' && <Button type="primary" ghost style={{margin: 6}} onClick={() => this.onActiveUser(record)}>激活</Button>}
                    </div>)
                }
            }
        ];
        return (<Card
            bordered={false}
            style={{overflow: 'visible'}}
            title="用户列表"
            headStyle={{border: 0}}
            bodyStyle={{padding: '0 32px 40px 32px'}}
            extra={
                <Row type="flex" justify="start" align="middle" gutter={20}>
                    <Col>
                        <Search style={{width: '500px'}} placeholder="搜索账号" onSearch={this.onSearch}
                                enterButton/>
                    </Col>
                    <Col>
                        <Button icon="plus"
                                onClick={this.onCreateUser}>创建用户</Button>
                    </Col>
                </Row>

            }
        >

            {
                showForm && <CustomizedForm onOk={this.onOk} mode={this.state.mode} onCancel={this.onCancel} canSubmit={this.state.canSubmit}
                                          editUser={this.state.editUser}/>
            }
            {
                !showForm &&  <Table
                    rowKey="id"
                    columns={columns}
                    pagination={paginationProps}
                    dataSource={users}
                    loading={loading}
                />
            }
        </Card>);
    }
}

class CustomizedForm extends React.Component {
    render() {
        const {form: {getFieldDecorator}, form, mode, onOk, onCancel, canSubmit, editUser} = this.props;
        console.log(this.props);
        const ops = mode === "create" ? "创建" : "编辑";
        const title = ops+"用户";
        return (
            <Card title={title} bordered={false}>
                <Form>
                    <Form.Item label="账号" labelCol={{span: 6}} wrapperCol={{span: 12}}>
                        {getFieldDecorator('id', {
                            rules: [
                                {required: true, message: 'Please input user id!'},
                                {
                                    pattern: new RegExp(/^[a-zA-Z0-9_]+$/),
                                    message: 'user name can only contain (A-Z a-z 0-9 _) !'
                                },
                            ],
                            initialValue: editUser ? editUser.id : "",
                        })(
                            <Input placeholder="user id" disabled={editUser}/>
                        )}
                    </Form.Item>

                    <Form.Item label="密码" labelCol={{span: 6}} wrapperCol={{span: 12}}>
                        {getFieldDecorator('password', {
                            rules: [
                                {required: mode === 'create', message: 'Please input user password!'},
                            ],
                        })(
                            <Input type="password" placeholder={mode === 'create' ? 'password' : 'password, if empty not update'}/>
                        )}
                    </Form.Item>

                    <Form.Item label="全名" labelCol={{span: 6}} wrapperCol={{span: 12}} >
                        {getFieldDecorator('fullname', {
                            rules: [
                                {required: true, message: 'Please input user fullname!'},
                            ],
                            initialValue: editUser ? editUser.fullname : "",
                        })(
                            <Input placeholder="fullname"/>
                        )}
                    </Form.Item>
                    <Form.Item label="邮箱" labelCol={{span: 6}} wrapperCol={{span: 12}}>
                        {getFieldDecorator('email', {

                            rules: [
                                {required: true, message: 'Please input user email!'},
                            ],
                            initialValue: editUser ? editUser.email : "",
                        })(
                            <Input placeholder="user email"/>
                        )}
                    </Form.Item>
                    <Form.Item label="组织" labelCol={{span: 6}} wrapperCol={{span: 12}}>
                    {getFieldDecorator('organization', {

                        rules: [
                            {required: true, message: 'Please input user organization!'},
                        ],
                        initialValue: editUser ? editUser.organization : "",
                    })(
                    <Input placeholder="user organization"/>
                    )}
                    </Form.Item>
                    <Form.Item label="电话" labelCol={{span: 6}} wrapperCol={{span: 12}}>
                        {getFieldDecorator('phone', {
                            rules: [],
                            initialValue: editUser ? editUser.phone : "",
                        })(
                            <Input placeholder="phone number" />
                        )}
                    </Form.Item>
                    <Form.Item label="微信号" labelCol={{span: 6}} wrapperCol={{span: 12}}>
                        {getFieldDecorator('wechat', {
                            rules: [],
                            initialValue: editUser ? editUser.wechat : "",
                        })(
                            <Input placeholder="wechat"/>
                        )}
                    </Form.Item>

                    <Form.Item>
                        <Row type="flex" justify="center" align="middle" gutter={20}>
                            <Button type="primary" ghost style={{margin: 10}} disabled={!canSubmit}
                                    onClick={() => onOk(form)}>
                                提交
                            </Button>
                            <Button type="primary" ghost style={{margin: 10}} onClick={onCancel}>取消</Button>
                        </Row>
                    </Form.Item>
                </Form>
            </Card>
        )
    }
}

CustomizedForm = Form.create({})(CustomizedForm);

function mapStateToProps(state) {
    return {
        ...state.user,
        loading: state.common.loading,
        setting: state.common.setting,
    };
}


function mapDispatchToProps(dispatch, ownProps) {
    return {
        dispatch: dispatch,
    }
}

export default withRouter(connect(mapStateToProps, mapDispatchToProps)(User));

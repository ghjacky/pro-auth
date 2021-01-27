import React from 'react';
import {Card, Table, Modal, Tag, Tooltip, Button, Form, Input, Col, Row, Popconfirm, message} from 'antd';
import {roleTypeConfig} from "../../utils/config";
import {withRouter} from "react-router-dom";
import {connect} from "react-redux";
import {queryUserAndClientRoles} from '../../redux/action/role';
import request from "../../utils/request";
import UserUpdateRoleTypeModal from '../../components/dashboard/UserUpdateRoleTypeModal'
const confirm = Modal.confirm;
const Search = Input.Search;

class Member extends React.Component {
    state = {
        isEdit: false,
        editUser: {},
        userList: [],
        userListOrigin: [],
        adminRoleMap: {},
        superRoleMap: {},
        currentUser: {},
        client_id: -1,
        isRootRoleSuper: -1,
    };

    componentDidMount() {
        const {match} = this.props;
        const client_id = match.params.client_id;
        this.setState({client_id});
        this.props.dispatch(queryUserAndClientRoles(['userUserRoles', 'userClientRoles'], client_id, false, true, true, 'admin'));
    }

    componentWillReceiveProps(nextProps) {
        const {userUserRoles, userClientRoles,user} = nextProps;
        if (userUserRoles && user) {
            const adminRoleMap = {};
            const superRoleMap = {};
            for (let i = 0; i < userUserRoles.length; i++) {
                if (userUserRoles[i].role_type === 'admin') {
                    adminRoleMap[userUserRoles[i].id] = userUserRoles[i];
                } else if (userUserRoles[i].role_type === 'super') {
                    superRoleMap[userUserRoles[i].id] = userUserRoles[i];
                    if(userUserRoles[i].parent_id === -1) {
                        this.setState({isRootRoleSuper: userUserRoles[i].id})
                    }
                }
            }
            this.setState({adminRoleMap, superRoleMap});
        }
        if (userClientRoles) {
            const userList = dataConvert(userClientRoles);
            this.setState({userList: userList, userListOrigin: userList});
        }
        if(user) {
            this.setState({currentUser: user});
        }
    }

    editUser = (user) => {
        this.setState({editUser: {...JSON.parse(JSON.stringify(user)),selectedRowKeys:[]}, isEdit: true})
    };
    returnToList = () => {
        this.setState({editUser: {selectedRowKeys:[]}, isEdit: false})
    };
    onSearch = (value) => {
        const {userListOrigin} = this.state;
        const userList = userListOrigin.filter(user => {
            if(user.fullname.indexOf(value) > -1) {
                return true;
            }
            if(user.id.indexOf(value) >-1) {
                return true;
            }
            if(user.dn.indexOf(value) >-1) {
                return true;
            }
            for(let i = 0; i < user.roles.length; i++) {
                if(user.roles[i].name.indexOf(value) > -1) {
                    return true
                }
            }
            return false;
        });
        this.setState({userList});
    };
    freshData = () => {
        const {client_id} = this.state;
        this.props.dispatch(queryUserAndClientRoles(['userUserRoles', 'userClientRoles'], client_id, false, true, true, 'admin'));
        this.setState({isEdit: false,editUser: {}});
    };
    onDeleteRoleUser = (user_id, role_id)=> {
        const { client_id} = this.state;
        request(`/api/roleUsers/${role_id}?client_id=${client_id}`,{
            method: 'DELETE',
            body:JSON.stringify([
                user_id
            ])
        }).then(res => {
            if(res.res_code === 0) {
                message.success('删除成员成功！');
                this.setState({selectedRowKeys: []});
                this.freshData();
            } else {
                message.error(res.res_msg);
            }
        })
    };
    onDeletePatchRoleUser = ()=> {
        const {editUser:{ selectedRowKeys}, client_id, currentRole} = this.state;
        const that = this;
        confirm({
            title: `确定删除选中${selectedRowKeys.length}个角色?`,
            okText: '确定',
            okType: 'danger',
            cancelText: '取消',
            onOk() {
                request(`/api/roleUsers/${currentRole.id}?client_id=${client_id}`,{
                    method: 'DELETE',
                    body:JSON.stringify(selectedRowKeys)
                }).then(res => {
                    if(res.res_code === 0) {
                        message.success('删除角色成功！');
                        that.setState({selectedRowKeys: []});
                        that.freshData();
                    } else {
                        message.error(res.res_msg);
                    }
                })
            },
            onCancel() {
            },
        });
    };
    render() {
        const {isEdit, editUser, userList, adminRoleMap, superRoleMap, currentUser,client_id,isRootRoleSuper} = this.state;
        const {loading} = this.props;
        const columns = [
            {
                title: '姓名',
                dataIndex: 'fullname',
                key: 'fullname',
            },
            {
                title: '账号',
                dataIndex: 'id',
                key: 'id',
            },
            {
                title: 'dn',
                dataIndex: 'dn',
                key: 'user.dn',
                render: (text, record) => {
                    if (text === undefined) {
                        return '';
                    } else {
                        return text.replace(',OU=HABROOT,DC=xxxxx,DC=corp', '').replace(record.fullname + ',', '').replace(/[(OU=|(CN=)]+/g, '').replace(/,/g, '_')
                    }
                },
            },
            {
                title: '角色',
                dataIndex: 'roles',
                key: 'roles',
                render: (text, record) => {
                    return record.roles.map((role) => {
                        return (<Tooltip key={role.id} placement="bottomLeft"
                                         title={`ta在该角色下的身份是${roleTypeConfig[role.role_type].name}`}>
                            <Tag style={{height: 20}} color={roleTypeConfig[role.role_type].color}>{role.name}</Tag>
                        </Tooltip>);
                    })
                },
                sorter: (a, b) => {
                    const rolesA = a.roles;
                    const rolesB = b.roles;
                    let superCountA = 0;
                    let adminCountA = 0;
                    let normalCountA = 0;
                    let superCountB = 0;
                    let adminCountB = 0;
                    let normalCountB = 0;
                    for (let i = 0; i < rolesA.length; i++) {
                        if (rolesA[i].role_type === 'super') {
                            superCountA++;
                        } else if (rolesA[i].role_type === 'admin') {
                            adminCountA++;
                        } else {
                            normalCountA++;
                        }
                    }
                    for (let i = 0; i < rolesB.length; i++) {
                        if (rolesB[i].role_type === 'super') {
                            superCountB++;
                        } else if (rolesB[i].role_type === 'admin') {
                            adminCountB++;
                        } else {
                            normalCountB++;
                        }
                    }
                    if (superCountA - superCountB !== 0) {
                        return superCountA - superCountB;
                    } else if (adminCountA - adminCountB !== 0) {
                        return adminCountA - adminCountB;
                    } else {
                        return normalCountA - normalCountB;
                    }
                }
            },
            {
                title: '',
                dataIndex: 'operation',
                key: 'operation',
                render: (text, record) => {
                    let disabled = record.id === currentUser.id;
                    return (<div>
                        <Button type="primary" ghost disabled={disabled} onClick={() => this.editUser(record)}>编辑</Button>
                    </div>)
                }
            }
        ];
        const column2 = [
            {
                title: '角色id',
                dataIndex: 'id',
                key: 'id',
            },
            {
                title: '角色名',
                dataIndex: 'name',
                key: 'name',
            },
            {
                title: 'ta在该角色的身份',
                dataIndex: 'role_type',
                key: 'role_type',
                render: (text, record) => {
                    return (
                        <Tooltip key={record.id} placement="bottomLeft" title={roleTypeConfig[text].effect}>
                            <Tag color={roleTypeConfig[text].color}>{roleTypeConfig[text].name}</Tag>
                        </Tooltip>
                    );
                }
            },
            {
                title: '操作',
                dataIndex: 'operation',
                key: 'operation',
                render: (text, record) => {
                    let hasRole = false;
                    if(record.role_type === 'super') {
                        hasRole = isRootRoleSuper > 0;
                    } else {
                        hasRole = adminRoleMap[record.id +'']|| superRoleMap[record.id +'']
                    }
                    if (hasRole) {
                        return (
                            <div>
                                <UserUpdateRoleTypeModal client_id={client_id} role_id={record.id} isRootRoleSuper={isRootRoleSuper > 0} user={{
                                    id: editUser.id,
                                    fullname: editUser.fullname,
                                    role_type: record.role_type
                                }} onOk={this.freshData}>
                                    <Button type="primary" size="small" ghost>修改身份</Button>
                                </UserUpdateRoleTypeModal>
                                <Popconfirm title={`确定删除${editUser.fullname}的${record.name}角色 ?`} cancelText="取消" okText="确定" okType="danger" onConfirm={()=> this.onDeleteRoleUser(editUser.id,record.id)}>
                                    <Button style={{marginLeft: 10}} type="danger" size="small">删除</Button>
                                </Popconfirm>
                            </div>
                        );
                    } else {
                        return "无权管理"
                    }
                }
            }

        ];
        const rowSelection = {
            selectedRowKeys: editUser.selectedRowKeys,
            onChange: (selectedRowKeys) => {
                const { editUser } = this.state;
                editUser['selectedRowKeys'] = selectedRowKeys;
                this.setState({editUser});
            },
            getCheckboxProps: record => {
                let hasRole = false;
                if(record.role_type === 'super') {
                    hasRole = superRoleMap[record.id +'']
                } else {
                    hasRole = adminRoleMap[record.id +'']||superRoleMap[record.id +'']
                }
                return ({
                    disabled: !hasRole,
                    name: record.name,
                });
            }
        };
        const formItemLayout = {
            labelCol: {
                span: 8
            },
            wrapperCol: {
                span: 8
            },
        };

        return (<div style={{margin: 20}}>
            {!isEdit && <Card
                bordered={false}
                style={{overflow: 'visible'}}
                title={<Row type="flex" justify="start" align="middle" gutter={20}>
                    <Col>
                        应用成员列表
                    </Col>
                    <Col>
                        <Search style={{width: '500px'}} placeholder="搜索" onSearch={this.onSearch}/>
                    </Col>
                </Row>}

                headStyle={{border: 0}}
                bodyStyle={{padding: '0 32px 40px 32px'}}
            >
                <Table
                    rowKey="id"
                    columns={columns}
                    dataSource={userList}
                    loading={loading}
                />
            </Card>
            }
            {
                isEdit && <Card
                    bordered={false}
                    style={{overflow: 'visible'}}
                    title={<Button type="primary" onClick={this.returnToList}>返回列表</Button>}
                    headStyle={{border: 0}}
                    bodyStyle={{padding: '0 32px 40px 32px'}}
                >
                    <Form>
                        <Form.Item
                            {...formItemLayout}
                            label="姓名"
                        >
                            <Input readOnly value={editUser.fullname}/>
                        </Form.Item>
                        <Form.Item
                            {...formItemLayout}
                            label="账号"
                        >
                            <Input readOnly value={editUser.id}/>
                        </Form.Item>
                        <Form.Item
                            {...formItemLayout}
                            label="dn"
                        >
                            <Input readOnly
                                   value={!editUser.dn ? "" : editUser.dn.replace(',OU=HABROOT,DC=creditease,DC=corp', '').replace(editUser.fullname + ',', '').replace(/[(OU=|(CN=)]+/g, '').replace(/,/g, '_')}/>
                        </Form.Item>
                        <Form.Item
                            {...formItemLayout}
                            label="角色">
                            {/*<Button disabled={editUser.selectedRowKeys.length === 0} style={{marginLeft: 10}} type="danger">批量删除角色</Button>*/}
                        </Form.Item>
                        <Form.Item wrapperCol={{span: 16, offset: 4}}>
                            <Table rowSelection={rowSelection} columns={column2} dataSource={editUser.roles}
                                   pagination={false}/>
                        </Form.Item>
                    </Form>

                </Card>
            }

        </div>);
    }
}

function mapStateToProps(state) {
    return {
        userUserRoles: state.role.userUserRoles,
        userClientRoles: state.role.userClientRoles,
        loading: state.common.loading,
        user: state.common.user
    };
}


function mapDispatchToProps(dispatch, ownProps) {
    return {
        dispatch: dispatch,
    }
}

export default withRouter(connect(mapStateToProps, mapDispatchToProps)(Member));


function dataConvert(data) {
    const userMap = {};
    for (let i = 0; i < data.length; i++) {
        let role = data[i];
        if (role.users === undefined) {
            continue
        }
        for (let j = 0; j < role.users.length; j++) {
            let roleUser = role.users[j];
            let user = roleUser.user;
            let obj = {
                id: role.id,
                name: role.name,
                role_type: roleUser.role_type
            };
            if (userMap[user.id] === undefined) {
                userMap[user.id] = {
                    id: user.id,
                    fullname: user.fullname,
                    dn: user.dn,
                    super: [],
                    admin: [],
                    normal: [],
                };
                userMap[user.id][obj.role_type].push(obj);
            } else {
                userMap[user.id][obj.role_type].push(obj);
            }
        }
    }
    const res = [];
    Object.keys(userMap).map((key) => {
        userMap[key].roles = [];
        userMap[key].roles.push.apply(userMap[key].roles, userMap[key].super);
        userMap[key].roles.push.apply(userMap[key].roles, userMap[key].admin);
        userMap[key].roles.push.apply(userMap[key].roles, userMap[key].normal);
        res.push(userMap[key]);
        return key
    });
    return res;
}

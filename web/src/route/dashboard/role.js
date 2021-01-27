import React from 'react';
import {Button, Card, Form, Icon, Input, Layout, message, Modal, Popconfirm, Table, Tabs, Tag, Tooltip} from 'antd';
import RoleTree from '../../components/tree/RoleTree';
import {buildTree} from "../../utils/common";
import {roleTypeConfig} from "../../utils/config";
import UserPatchAdd from '../../components/dashboard/UserPatchAdd';
import RoleResourceTransfer from '../../components/dashboard/RoleResourceTransfer';
import {withRouter} from "react-router-dom";
import {connect} from "react-redux";
import {queryUserRoles} from '../../redux/action/role';
import RoleEditModal from '../../components/dashboard/RoleEditModal';
import RoleAddModal from '../../components/dashboard/RoleAddModal';
import RoleInsertModal from '../../components/dashboard/RoleInsertModal';
import UserUpdateRoleTypeModal from '../../components/dashboard/UserUpdateRoleTypeModal';
import request from '../../utils/request';

const {Sider, Content} = Layout;
const TabPane = Tabs.TabPane;
const confirm = Modal.confirm;
const Search = Input.Search;
class Role extends React.Component {
    state = {
        selectedKeys: [],
        selectedRowKeys:[],
        roleMap: {},
        canDelete: true,
        currentRole: {},
        isEditTab1: false,
        isEditTab2: false,
        idEditTab3: false,
        isRootRoleSuper: -1,
        rootRoleId: -1,
        adminRoleMap: {},
        superRoleMap: {},
        user: {},
        client_id: -1,
        searchUserText:'',
    };

    componentDidMount() {
        const {match} = this.props;
        const client_id = match.params.client_id;
        this.setState({client_id});
        this.props.dispatch(queryUserRoles('roleUserRoles',client_id, 'admin', true, true, true, true));
    }
    componentWillReceiveProps(nextProps, nextContext) {
        const { roleUserRoles, user} = nextProps;
        const {selectedKeys} = this.state;
        if(roleUserRoles) {
            const roleMap = {};
            const adminRoleMap = {};
            const superRoleMap = {};
            let isRootRoleSuper = -1;
            for (let i = 0; i < roleUserRoles.length; i++) {
                roleMap[roleUserRoles[i].id] = roleUserRoles[i];
                if (roleUserRoles[i].parent_id === -1 ) {
                    if(roleUserRoles[i].role_type === 'super') {
                        isRootRoleSuper = roleUserRoles[i].id;
                    }
                    this.setState({ rootRoleId: roleUserRoles[i].id});
                }
                if (roleUserRoles[i].role_type === 'admin') {
                    adminRoleMap[roleUserRoles[i].id] = roleUserRoles[i];
                } else if (roleUserRoles[i].role_type === 'super') {
                    superRoleMap[roleUserRoles[i].id] = roleUserRoles[i];
                }
            }
            this.setState({roleMap, roleTree:buildTree(roleUserRoles) ,adminRoleMap, superRoleMap, isRootRoleSuper});
            if (selectedKeys.length > 0 && roleMap[selectedKeys[0]] !== undefined) {
                const currentRole = JSON.parse(JSON.stringify(roleMap[selectedKeys[0]]));
                currentRole.usersOrigin = currentRole.users;
                currentRole.resourcesOrigin = currentRole.resources;
                this.setState({currentRole});
            }
        }
        if(user) {
            this.setState({user});
        }
    }
    freshData = ()=> {
        const {client_id} = this.state;
        this.props.dispatch(queryUserRoles('roleUserRoles',client_id, 'admin', true, true, true, true));
    };
    //role tree
    onSelect = (selectedKeys) => {
        const {roleMap} = this.state;
        if (selectedKeys.length > 0 && roleMap[selectedKeys[0]] !== undefined) {
            const currentRole = JSON.parse(JSON.stringify(roleMap[selectedKeys[0]]));
            currentRole.usersOrigin = currentRole.users;
            currentRole.resourcesOrigin = currentRole.resources;
            this.setState({currentRole: currentRole});
        }
        this.setState({selectedKeys, selectedRowKeys:[], isEditTab1: false , isEditTab2: false, isEditTab3: false});
    };
    //tab1
    onDeleteRole = () => {
        const { currentRole, client_id} = this.state;
        this.setState({canDelete: false});
        request(`/api/roles/${currentRole.id}?client_id=${client_id}`, {method: 'DELETE'}, false)
            .then(res => {
                this.setState({canDelete: true});
                if(res.res_code === 0) {
                    message.success('删除角色成功');
                    this.setState({selectedKeys:[] , currentRole:{}});
                    this.freshData();
                } else {
                    message.error(res.res_msg);
                }
            })

    };
    // tab2
    onEditTab2 = ()=> {
        this.setState({isEditTab2: !this.state.isEditTab2});
    };
    onDeleteRoleUser = (user_id)=> {
        const { client_id, currentRole} = this.state;
        request(`/api/roleUsers/${currentRole.id}?client_id=${client_id}`,{
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
        const {selectedRowKeys, client_id, currentRole} = this.state;
        const that = this;
        confirm({
            title: `确定删除选中${selectedRowKeys.length}名成员的角色?`,
            okText: '确定',
            okType: 'danger',
            cancelText: '取消',
            onOk() {
                request(`/api/roleUsers/${currentRole.id}?client_id=${client_id}`,{
                    method: 'DELETE',
                    body:JSON.stringify(selectedRowKeys)
                }).then(res => {
                    if(res.res_code === 0) {
                        message.success('删除成员成功！');
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
    onSearchUser = (value)=> {
        const {currentRole} = this.state;
        currentRole.users = currentRole.usersOrigin.filter(record => {
           if(record.user.id.indexOf(value) > -1) {
               return true;
           }
           if(record.user.fullname.indexOf(value) >-1) {
               return true;
           }
           if(record.user.dn.indexOf(value) > -1) {
               return true;
           }
           if(roleTypeConfig[record.role_type].name.indexOf(value) > -1) {
               return true;
           }
           return false;
        });
        this.setState({currentRole});
    };
    //tab3
    onEditTab3 = ()=> {
        this.setState({isEditTab3: !this.state.isEditTab3});
    };
    onDeleteRoleResource = (resource_id)=>{
        const {currentRole, client_id} = this.state;
        request(`/api/roleResources/${currentRole.id}?client_id=${client_id}`,{
            method: 'DELETE',
            body: JSON.stringify([resource_id])
        }).then(res => {
            if(res.res_code === 0) {
                message.success("解除权限关联成功！");
                this.freshData();
            } else {
                message.error(res.res_msg);
            }
        })
    };
    onSearchResource = (value)=> {
        const {currentRole} = this.state;
        currentRole.resources = currentRole.resourcesOrigin.filter(record => {
            if((record.resource.id +'' ).indexOf(value) > -1) {
                return true;
            }
            if(record.resource.name.indexOf(value) >-1) {
                return true;
            }
            if(record.resource.description.indexOf(value) > -1) {
                return true;
            }
            if(record.resource.data.indexOf(value) > -1) {
                return true;
            }
            return false;
        });
        this.setState({currentRole});
    };
    //common
    isParentRoleSuper = (role_id) => {
        const {superRoleMap} = this.state;
        if (superRoleMap[role_id] !== undefined){
            return superRoleMap[role_id].parent_id === -1 ? true : superRoleMap[superRoleMap[role_id].parent_id] !== undefined;
        } else {
            return false;
        }
    };
    isRoleSuper = (role_id) => {
        const {superRoleMap} = this.state;
        return superRoleMap[role_id] !== undefined;
    };
    render() {
        const {selectedKeys,selectedRowKeys,user, client_id, isEditTab2,isEditTab3, currentRole,roleMap ,isRootRoleSuper, roleTree, rootRoleId,canDelete} = this.state;
        const columns = [
            {
                title: '姓名',
                dataIndex: 'user.fullname',
                key: 'fullname',
            },
            {
                title: '账号',
                dataIndex: 'user.id',
                key: 'id',
            },
            {
                title: 'dn',
                dataIndex: 'user.dn',
                key: 'user.dn',
                render: (text, record) => {
                    if (text === undefined) {
                        return '';
                    } else {
                        return text.replace(',OU=HABROOT,DC=creditease,DC=corp', '').replace(record.user.fullname + ',', '').replace(/[(OU=|(CN=)]+/g, '').replace(/,/g, '_')
                    }
                },
            },
            {
                title: 'ta在该角色的身份',
                dataIndex: 'role_type',
                key: 'role_type',
                sorter: (a, b) => {
                    let valueOfa = a.role_type === 'super' ? 2 : a.role_type === 'admin' ? 1 : 0;
                    let valueOfb = b.role_type === 'super' ? 2 : b.role_type === 'admin' ? 1 : 0;
                    return valueOfa - valueOfb;
                },
                render: (text, record) => {
                    return (
                        <Tooltip key={record.id} placement="bottomLeft" title={roleTypeConfig[text].effect}>
                            <Tag color={roleTypeConfig[text].color}>{roleTypeConfig[text].name}</Tag>
                        </Tooltip>
                    );
                }
            },
            {
                title: '',
                dataIndex: 'operation',
                key: 'operation',
                render: (text, record) => {
                    const {isRootRoleSuper} = this.state;
                    const disabled = (record.user.id  === user.id && currentRole.parent_id === -1) || (record.role_type === 'super' &&  isRootRoleSuper < 0);
                    return (<div>
                        <UserUpdateRoleTypeModal isRootRoleSuper={isRootRoleSuper > 0} role_id={currentRole.id} client_id={client_id} user={{role_type:record.role_type, ...record.user}} onOk={this.freshData}>
                            <Button disabled={disabled} type="primary" ghost style={{margin: '0 10px 0 10px'}}>修改身份</Button>
                        </UserUpdateRoleTypeModal>
                        {
                            disabled && <Button disabled={disabled} type="primary" ghost style={{margin: '0 10px 0 10px'}}
                            >删除</Button>
                        }
                        {
                            !disabled && <Popconfirm title={`确定删除${record.user.fullname} ?`} onConfirm={()=> this.onDeleteRoleUser(record.user.id)}>
                                <Button type="primary" ghost style={{margin: '0 10px 0 10px'}}
                                >删除</Button></Popconfirm>
                        }
                    </div>)
                }
            }
        ];
        const columns2 = [
            {
                title: 'id',
                dataIndex: 'resource.id',
                key: 'id',
                sorter: (a,b) => a.resource.id -b.resource.id
            },
            {
                title: '权限名称',
                dataIndex: 'resource.name',
                key: 'name',
            },
            {
                title: '权限说明',
                dataIndex: 'resource.description',
                key: 'description',
            },
            {
                title: '权限内容',
                dataIndex: 'resource.data',
                key: 'data',
            },
            {
                title: '操作',
                key: 'operation',
                render: (text, record) => {
                    const disabled = !this.isParentRoleSuper(currentRole.id);
                    return (<Popconfirm title={`确定解除该权限的关联(与该角色及其子角色)? `} cancelText="取消" okText="确定" onConfirm={()=> this.onDeleteRoleResource(record.resource.id)}>
                        <a disabled={disabled}>解除关联</a>
                    </Popconfirm>);

                }
            },
        ];
        if (currentRole.id === rootRoleId) {
            columns2.splice(4,1);
        }
        const formItemLayout = {
            labelCol: {
                span: 2
            },
            wrapperCol: {
                span: 8
            },
        };
        const rowSelection = {
            selectedRowKeys:selectedRowKeys,
            onChange: (selectedRowKeys) => {
                this.setState({selectedRowKeys});
            },
            getCheckboxProps: record => ({
                disabled: (record.user.id  === user.id && currentRole.parent_id === -1) || (record.role_type === "super" && isRootRoleSuper < 0)
            }),
        };
        let resourceRange = [];
        if (isRootRoleSuper > 0) {
            resourceRange = roleMap[isRootRoleSuper].resources;
        } else if(roleMap[currentRole.parent_id] && roleMap[currentRole.parent_id].resources){
            resourceRange = roleMap[currentRole.parent_id].resources;
        }
        return (<Layout style={{margin: 20}}>
            <Sider style={{background: 'white'}} width={'auto'}>
                <RoleTree title="角色树" onSelect={this.onSelect} treeNodes={roleTree}/>
            </Sider>
            <Layout>
                <Content style={{background: 'white'}}>
                    {selectedKeys.length > 0 && <Tabs defaultActiveKey="1">
                        <TabPane tab={<span><Icon type="tags"/>角色信息管理</span>} key="1">
                            <Card>
                                <Form>
                                    <Form.Item
                                        {...formItemLayout}
                                        label="id"
                                    >
                                        <p>{currentRole.id}</p>
                                    </Form.Item>
                                    <Form.Item
                                        {...formItemLayout}
                                        label="角色名"
                                    >
                                        <p>{currentRole.name}</p>
                                    </Form.Item>
                                    <Form.Item
                                        {...formItemLayout}
                                        label="角色说明"
                                    >
                                        <p>{currentRole.description}</p>
                                    </Form.Item>
                                    <Form.Item
                                        wrapperCol={{span: 16, offset: 1}}
                                    >
                                        {
                                            this.isParentRoleSuper(currentRole.id)&&<RoleEditModal client_id={client_id} currentRole={currentRole} onOk={this.freshData}>
                                                <Button style={{margin: 10}} type="primary" ghost>编辑角色</Button>
                                            </RoleEditModal>
                                        }
                                        {
                                            this.isRoleSuper(currentRole.id)&&<Tooltip key="addSubRole" placement="bottomLeft" title="为该角色新增一个子角色">
                                                <RoleAddModal parent_id={currentRole.id} client_id={client_id} onOk={this.freshData}>
                                                    <Button style={{margin: 10}} type="primary" ghost>新增子角色</Button>
                                                </RoleAddModal>
                                            </Tooltip>
                                        }
                                        {
                                            this.isRoleSuper(currentRole.id)&& currentRole.children && currentRole.children.length > 0 &&<Tooltip key="insertSubRole" placement="bottomLeft" title="在该角色与其子角色之间新增一个角色（即增加了一层）">
                                                <RoleInsertModal parentRole={currentRole} client_id={client_id} onOk={this.freshData}>
                                                    <Button style={{margin: 10}} type="primary" ghost>插入子角色</Button>
                                                </RoleInsertModal>
                                            </Tooltip>
                                        }
                                        {
                                            this.isParentRoleSuper(currentRole.id) && rootRoleId !== currentRole.id && canDelete && <Popconfirm title={`确定删除该角色? `} cancelText="取消" okText="确定" onConfirm={this.onDeleteRole}>
                                                <Button style={{margin: 10}} disabled={!canDelete} type="primary" ghost >删除角色</Button>
                                            </Popconfirm>
                                        }
                                        {
                                            this.isParentRoleSuper(currentRole.id)&& !canDelete &&  <Button style={{margin: 10}} disabled type="primary" ghost >删除角色</Button>
                                        }
                                    </Form.Item>
                                </Form>
                            </Card>
                        </TabPane>
                        <TabPane tab={<span><Icon type="usergroup-add"/>角色成员管理</span>} key="2">
                            {
                                !isEditTab2 &&  <Card>
                                    <Button type="primary" ghost style={{margin: 10}} onClick={this.onEditTab2}>批量添加成员</Button>
                                    <Button disabled={selectedRowKeys.length === 0} type="primary" ghost onClick={this.onDeletePatchRoleUser} style={{margin: 10}}>批量删除成员</Button>
                                    <div><Search style={{margin: 10, width:'60%'}} placeholder="搜索成员" onSearch={this.onSearchUser} /></div>
                                    <Table rowKey={(record) => record.user.id} rowSelection={rowSelection} columns={columns} dataSource={currentRole.users}/>
                                </Card>
                            }
                            {
                                isEditTab2 &&  <Card>
                                    <UserPatchAdd client_id={client_id} role_id={currentRole.id} isRootRoleSuper={isRootRoleSuper} existedUsers={currentRole.users} onOk={this.freshData} onCancel={this.onEditTab2} />
                                </Card>
                            }
                        </TabPane>
                        <TabPane tab={<span><Icon type="bars"/>角色权限管理</span>} key="3">
                            {
                                !isEditTab3 && <Card> {
                                    currentRole.id !== rootRoleId && <Button type="primary" ghost style={{margin: 10}} disabled={!this.isParentRoleSuper(currentRole.id)} onClick={this.onEditTab3}>关联权限</Button>
                                }
                                    <div><Search style={{margin: 10, width:'60%'}} onSearch={this.onSearchResource} placeholder="搜索权限"/></div>
                                    <Table rowKey="id" columns={columns2} dataSource={currentRole.resources}/>
                                </Card>
                            }
                            {
                                isEditTab3 && <Card bodyStyle={{textAlign: 'center'}}>
                                    <RoleResourceTransfer onCancel={this.onEditTab3} allData={resourceRange} client_id={client_id} role_id={currentRole.id} onOk={this.freshData} selectedData={currentRole.resources} />
                                </Card>
                            }

                        </TabPane>
                    </Tabs>
                    }
                    {
                        selectedKeys.length === 0 && <Card title="请选择一个角色" headStyle={{fontSize: 14}} bordered={false}/>
                    }
                </Content>
            </Layout>
        </Layout>);
    }
}



function mapStateToProps(state) {
    return {
        roleUserRoles: state.role.roleUserRoles,
        loading: state.common.loading,
        user: state.common.user
    };
}


function mapDispatchToProps(dispatch, ownProps) {
    return {
        dispatch: dispatch,
    }
}

export default withRouter(connect(mapStateToProps, mapDispatchToProps)(Role));

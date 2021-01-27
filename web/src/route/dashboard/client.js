import React from 'react';
import {Card, Button, Table, Form, Spin, Popconfirm, message, Modal, Input, Row, Col} from "antd";
import '../../asset/icon/iconfont.css';
import {queryClient, queryClientAndResource, queryResources} from "../../redux/action/client";
import {withRouter} from "react-router-dom";
import {connect} from "react-redux";
import ClientEditModal from '../../components/dashboard/ClientEditModal';
import ResourceAddModal from '../../components/dashboard/ResourceAddModal';
import ResourceEditModal from '../../components/dashboard/ResourceEditModal';
import request from "../../utils/request";
const confirm = Modal.confirm;
const Search = Input.Search;

class Client extends React.Component {
    state = {
        showSecret: false,
        client : {
        },
        selectedRowKeys:[],
        client_id : -1,
        resources: [],
        resourcesOrigin:[]
    };
    componentDidMount() {
        const {match} = this.props;
        const client_id = match.params.client_id;
        this.setState({client_id});
        this.props.dispatch(queryClientAndResource(client_id));
    }
    componentWillReceiveProps(nextProps, nextContext) {
        const {client,resources}= nextProps;
        if(client) {
            this.setState({client});
        }
        if(resources) {
            this.setState({resources, resourcesOrigin: resources})
        }

    }


    freshClient = ()=> {
        const {client_id} = this.state;
        this.props.dispatch(queryClient(client_id));
    };
    freshResources = () => {
        const {client_id} = this.state;
        this.props.dispatch(queryResources(client_id));
    };
    onChangeSecret = ()=> {
        this.setState({
            showSecret: !this.state.showSecret
        });
    };
    onEdit = () => {
        this.setState({
            isEdit: !this.state.isEdit
        });
    };
    onDeleteResource = (id)=> {
        const { client } = this.state;
        request(`/api/resources/${id}?client_id=${client.id}`,{
            method: 'DELETE',
        }).then(res => {
            if(res.res_code === 0) {
                message.success('权限删除成功！');
                this.freshResources();
            } else {
                message.error(res.res_msg);
            }
        })
    };
    onDeletePatchResource = () => {
        const {selectedRowKeys, client} = this.state;
        const that = this;
        confirm({
            title: `确定删除选中的${selectedRowKeys.length}个权限?`,
            okText: '确定',
            okType: 'danger',
            cancelText: '取消',
            onOk() {
                request(`/api/resources/${selectedRowKeys.join(',')}?client_id=${client.id}`,{
                    method: 'DELETE',
                }).then(res => {
                    if(res.res_code === 0) {
                        message.success('权限删除成功！');
                        that.setState({selectedRowKeys:[]});
                        that.freshResources();
                    } else {
                        message.error(res.res_msg);
                    }
                })
            },
            onCancel() {
            },
        });
    };
    onDeleteResourceRole = (role_id, resource_id)=>{
        const {client_id} = this.state;
        request(`/api/roleResources/${role_id}?client_id=${client_id}`,{
            method: 'DELETE',
            body: JSON.stringify([resource_id])
        }).then(res => {
            if(res.res_code === 0) {
                message.success("解除权限关联成功！");
                this.freshResources();
            } else {
                message.error(res.res_msg);
            }
        })
    };
    onSearch = (value) => {
        const {resourcesOrigin} = this.state;
        const resources = resourcesOrigin.filter( resource => {
            return JSON.stringify(resource).indexOf(value) > -1;
        });
        this.setState({resources});
    };
    render() {

        const columns = [
            {
                title: 'id',
                dataIndex: 'id',
                key: 'id',
                sorter: (a, b) => a.id -b.id
            },
            {
                title: '权限名称',
                dataIndex: 'name',
                key: 'name',
            },
            {
                title: '权限说明',
                dataIndex: 'description',
                key: 'description',
            },
            {
                title: '权限内容',
                dataIndex: 'data',
                key: 'data',
            },
            {
                title: '操作',
                key: 'operation',
                render: (text, record) => (
                    <div>
                        <ResourceEditModal onOk={this.freshResources} client_id={client.id} resource={record}>
                            <Button style={{margin: 10}} type="primary" ghost>修改</Button>
                        </ResourceEditModal>
                        <Popconfirm title={"确定删除该权限？"} onConfirm={()=> this.onDeleteResource(record.id)}>
                            <Button style={{margin: 10}} type="primary" ghost>删除</Button>
                        </Popconfirm>
                    </div>
                ),
            },
        ];
        const formItemLayout = {
            labelCol: {
                span: 2
            },
            wrapperCol: {
                span: 8
            },
        };
        const {client,showSecret, selectedRowKeys, resources} = this.state;
        const {loading} = this.props;
        const rowSelection = {
            selectedRowKeys:selectedRowKeys,
            onChange: (selectedRowKeys) => {
                this.setState({selectedRowKeys});
            },
        };
        const expandedRowRender = (resource) => {
            const columns = [
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
                    title: '角色说明',
                    dataIndex: 'description',
                    key: 'description',
                },
                {
                    title: '父角色id',
                    dataIndex: 'parent_id',
                    key: 'parent_id',
                },
                {
                    title: '操作',
                    dataIndex: 'operation',
                    key: 'operation',
                    render : (text, record)=> {
                        return <Popconfirm title={`确定解除该权限的关联(与该角色及其子角色)? `} cancelText="取消" okText="确定" onConfirm={()=>this.onDeleteResourceRole(record.id, resource.id)}>
                            <a disabled={record.parent_id === -1}>解除关联</a>
                        </Popconfirm>
                    }
                }
            ];
            return (<div>
                    <Row>
                        <Col span={6}>创建人：{resource.created_by}</Col>
                        <Col span={14} offset={1}>创建时间：{resource.created.substring(0,19)}</Col>
                    </Row>
                    <Row>
                        <Col span={6}>更新人：{resource.updated_by}</Col>
                        <Col span={14} offset={1}>更新时间：{resource.updated.substring(0,19)}</Col>
                    </Row>
                    <Table
                        title={()=> <span>关联的角色</span>}
                        columns={columns}
                        dataSource={resource.roles}
                    />
                </div>
            );
        };
        return (<div>
            <Card title="基本信息">
                <Spin spinning={loading}>
                    <Form>
                        <Form.Item
                            {...formItemLayout}
                            label="client_id"
                        >
                            {client.id}
                        </Form.Item>
                        <Form.Item
                            {...formItemLayout}
                            label="fullname"
                        >
                            {client.fullname}
                        </Form.Item>
                        <Form.Item
                            {...formItemLayout}
                            label="secret"
                        >
                            {
                                !showSecret && <div><span style={{marginLeft: 11, marginRight: 11}}>{'*'.repeat(22)}</span><i onClick={this.onChangeSecret} className="icon iconfont icon-openEye"></i></div>
                            }
                            {
                                showSecret && <div><span style={{marginLeft: 11, marginRight: 11}}>{client.secret}</span><i onClick={this.onChangeSecret} className="icon iconfont icon-yanjing"></i></div>
                            }
                        </Form.Item>
                        <Form.Item
                            {...formItemLayout}
                            label="redirect_uri"
                        >
                            {client.redirect_uri}
                        </Form.Item>
                        <Form.Item
                            wrapperCol={{offset:2 ,span: 4}}
                        >
                            <ClientEditModal onOk={this.freshClient} client={client}>
                                <Button type="primary" style={{width: '100%'}} onClick={this.onEdit}>修改</Button>
                            </ClientEditModal>
                        </Form.Item>
                    </Form>
                </Spin>
            </Card>
            <Card title="权限列表">
                <ResourceAddModal client_id={client.id} onOk={this.freshResources}>
                    <Button style={{margin: 10}} type="primary">新增</Button>
                </ResourceAddModal>
                <Button style={{margin: 10}} type="primary" disabled={selectedRowKeys.length === 0} onClick={this.onDeletePatchResource}>批量删除</Button>
                <div><Search  style={{width:'60%',margin: 10}} placeholder="搜索权限" onSearch={this.onSearch}/></div>
                <Table rowKey="id" columns={columns} dataSource={resources} loading={loading} rowSelection={rowSelection} expandedRowRender={expandedRowRender}/>
            </Card>
        </div>);
    }
}

function mapStateToProps(state) {
    return {
        client: state.client.client,
        resources: state.client.resources,
        loading: state.common.loading
    };
}


function mapDispatchToProps(dispatch, ownProps) {
    return {
        dispatch: dispatch,
    }
}

export default withRouter(connect(mapStateToProps, mapDispatchToProps)(Client));

import React from 'react';
import {
    List,
    Avatar,
    Button,
    Card,
    Radio,
    Input,
    Row,
    Col,
    Tag,
    Tooltip,
    Menu,
    Dropdown,
    Icon,
    Modal,
    Table,
    Form,
    message
} from 'antd';
import {connect} from 'react-redux';
import {withRouter, Link} from 'react-router-dom'
import {queryClients} from '../../redux/action/client'
import {roleTypeConfig} from '../../utils/config';
import request from "../../utils/request";

const RadioGroup = Radio.Group;
const RadioButton = Radio.Button;
const Search = Input.Search;

class Client extends React.Component {
    state = {
        isEdit: false,
        canSubmit: true,
        showType: 'my',
        current: 1,
        pageSize: 10,
        searchText: '',
        userClients: [],
        userClientsOrigin: [],
        allClients: [],
        allClientsOrigin: []
    };

    componentDidMount() {
        this.props.dispatch(queryClients())
    }

    componentWillReceiveProps(nextProps) {
        const {userClients, allClients} = nextProps;
        if (userClients) {
            this.setState({userClients: userClients, userClientsOrigin: userClients});
        }
        if (allClients) {
            this.setState({allClients: allClients, allClientsOrigin: allClients});
        }
    }

    onSearch = (value) => {
        const {userClientsOrigin, allClientsOrigin, showType} = this.state;
        this.setState({userClients: userClientsOrigin, allClients: allClientsOrigin});
        if (showType === 'my') {
            let userClients = userClientsOrigin.filter(client => {
                console.log(client);
                if (client.fullname.indexOf(value) > -1) {
                    return true;
                }
                for (let i = 0; i < client.roles.length; i++) {
                    if (client.roles[i].name.indexOf(value) > -1) {
                        return true;
                    }
                }
                return false;
            });
            this.setState({userClients: userClients})
        } else if (showType === 'all') {
            let allClients = allClientsOrigin.filter(client => {
                if (client.fullname.indexOf(value) > -1) {
                    return true;
                }
                for (let i = 0; i < client.users.length; i++) {
                    if (client.users[i].user_id.indexOf(value) > -1 || client.users[i].fullname.indexOf(value) > -1) {
                        return true;
                    }
                }
                return false
            });
            this.setState({allClients: allClients})
        }
    };
    onChangeShowType = (e) => {
        this.setState({
            showType: e.target.value,
            current: 1,
        });
    };
    onOk = (form) => {
        form.validateFields((err, values) => {
            if (!err) {
                this.setState({canSubmit: false});
                request('/api/client',{
                    method: 'POST',
                    body: JSON.stringify(values),
                }, false).then(res => {
                    this.setState({canSubmit: true});
                    if(res.res_code === 0) {
                        message.success("成功创建应用");
                        this.setState({isEdit: false});
                        this.props.dispatch(queryClients());
                    } else {
                        Modal.error({
                            title: '创建应用失败',
                            content: res.res_msg,
                        });
                    }
                })
            }
        });
    };
    onEdit = () => {
        this.setState({isEdit: !this.state.isEdit});
    };

    render() {
        const {showType, current, pageSize, userClients, allClients, isEdit} = this.state;
        const {loading} = this.props;
        let total = 0;
        if (showType === 'my' && userClients) {
            total = userClients.length;
        } else if (showType === 'all' && allClients) {
            total = allClients.length;
        }
        const paginationProps = {
            current,
            total: total,
            pageSize: pageSize,
            showSizeChanger: true,
            showTotal: (total, range) => `${range[0] > 0 ? range[0] : 0}-${range[1] > 0 ? range[1] : 0} of ${total} client`,
            onShowSizeChange: (current, pageSize) => {
                this.setState({current: 1, pageSize: pageSize});
            },
            onChange: (page, pageSize) => {
                this.setState({current: page});
            }
        };
        return (<Card
            bordered={false}
            style={{overflow: 'visible'}}
            title="应用列表"
            headStyle={{border: 0}}
            bodyStyle={{padding: '0 32px 40px 32px'}}
            extra={
                <Row type="flex" justify="start" align="middle" gutter={20}>
                    <Col>
                        <Search style={{width: '500px'}} placeholder="搜索应用" onSearch={this.onSearch}
                                enterButton/>
                    </Col>
                    <Col>
                        <RadioGroup value={showType} onChange={this.onChangeShowType}>
                            <RadioButton value="my">我的应用</RadioButton>
                            <RadioButton value="all">全部应用</RadioButton>
                        </RadioGroup>
                    </Col>
                </Row>

            }
        >
            {
                !isEdit && <Button
                    type="dashed"
                    style={{width: '100%', marginBottom: 8, marginTop: 8}}
                    icon="plus"
                    onClick={this.onEdit}
                >
                    创建应用
                </Button>
            }
            {
                isEdit && <CustomizedForm onOk={this.onOk} onCancel={this.onEdit} canSubmit={this.state.canSubmit}/>
            }
            <List
                key="id"
                size="large"
                loading={loading}
                pagination={paginationProps}
                dataSource={showType === 'my' ? userClients : allClients}
                renderItem={item => {
                    let actions = [];
                    if (showType === 'my') {
                        let hasAdminRole = false;
                        let hasRootSuperRole = false;
                        for (let i = 0; i < item.roles.length; i++) {
                            if (item.roles[i].role_type === 'super' && item.roles[i].parent_id === -1) {
                                hasRootSuperRole = true;
                                break;
                            } else if (item.roles[i].role_type === 'admin' || item.roles[i].role_type === "super") {
                                hasAdminRole = true;
                            }
                        }
                        let menu = <div></div>;
                        if (hasRootSuperRole) {
                            menu = <Menu>
                                <Menu.Item key="user"><Link
                                    to={`/frontend/dashboard/user/${item.id}`}>成员管理</Link></Menu.Item>
                                <Menu.Item key="role"><Link
                                    to={`/frontend/dashboard/role/${item.id}`}>角色树管理</Link></Menu.Item>
                                <Menu.Item key="client"><Link
                                    to={`/frontend/dashboard/client/${item.id}`}>应用管理</Link></Menu.Item>
                            </Menu>
                        } else if (hasAdminRole) {
                            menu = <Menu>
                                <Menu.Item key="user"><Link
                                    to={`/frontend/dashboard/user/${item.id}`}>成员管理</Link></Menu.Item>
                                <Menu.Item key="role"><Link
                                    to={`/frontend/dashboard/role/${item.id}`}>角色树管理</Link></Menu.Item>
                            </Menu>
                        }
                        actions = [
                            <Button type="primary" ghost
                                    onClick={e => {
                                        e.preventDefault();
                                    }}
                            >
                                <Link to={`/frontend/dashboard/my/${item.id}`}>我的权限</Link>
                            </Button>,
                            <Dropdown
                                disabled={!hasAdminRole && !hasRootSuperRole}
                                overlay={menu}
                            >
                                <a>
                                    更多 <Icon type="down"/>
                                </a>
                            </Dropdown>
                        ]
                    } else if (showType === 'all') {
                        actions = [
                            <Button type="primary" ghost onClick={() => {
                                const column = [
                                    {
                                        title: '姓名',
                                        dataIndex: 'fullname',
                                        key: 'fullname',
                                    },
                                    {
                                        title: '邮箱',
                                        dataIndex: 'user_id',
                                        key: 'user_id',
                                        render: (record) => record + "@xxxxx.com"
                                    },
                                    {
                                        title: '身份',
                                        dataIndex: 'role_type',
                                        key: 'role_type',
                                        render: (record) => <Tag
                                            color={roleTypeConfig[record].color}>{item.fullname + roleTypeConfig[record].name}</Tag>
                                    },

                                ];
                                Modal.info({
                                    title: '联系管理员开通权限',
                                    width: 850,
                                    content: (<div>
                                            <p>1.请先确认您的同事是否已经是具有相关权限的角色的管理员，如果是，他们可以帮您开通权限</p>
                                            <p>2.若您需要某种权限，请任选一个管理员发送开通权限的申请邮件，邮件中请说明需要哪个应用下的哪些权限和角色的身份</p>
                                            <p>3.若您需要某种权限且需要自定义权限子集的划分，请选择一个应用超级管理员发送开通权限的申请邮件，邮件中请说明需要哪个应用下的哪些权限和需要自定义权限子集的原因</p>
                                            <p>角色的身份分为：</p>
                                            <p>（1）普通成员，仅可使用该角色权限</p>
                                            <p>（2）管理员，可使用该角色权限，并可管理该角色子树的成员</p>
                                            <p>（3）超级管理员，可使用该角色权限，可管理该角色子树的成员，可自定义该角色子树的组织结构与权限关联关系</p>
                                            <Table rowKey="user_id" columns={column} dataSource={item.users}
                                                   pagination={false}/>
                                        </div>
                                    ),
                                });
                            }}>联系管理员</Button>
                        ];
                    }

                    return <List.Item
                        key={item.id}
                        actions={actions}
                    >
                        <List.Item.Meta
                            avatar={<Avatar shape="square" size={52}><span
                                style={{fontSize: 35}}>{item.fullname.substring(0, 1).toUpperCase()}</span></Avatar>}
                            title={item.fullname}
                            description={<div>
                                <Row>client_id:{item.id}</Row>
                                <Row>created by {item.created_by}</Row>
                            </div>}
                        />
                        {
                            showType === 'my' && <div style={{
                                display: 'flex',
                                flexDirection: 'column',
                                justifyContent: 'center',
                                alignItems: 'flex-start',
                                width: '100%'
                            }}>
                                <span>Your Roles</span>
                                <div>
                                    {item.roles.map((role) => {
                                        return (
                                            <Tooltip key={role.id} placement="bottomLeft"
                                                     title={`你在该角色下的身份是${roleTypeConfig[role.role_type].name}`}>
                                                <Tag style={{height: 20}}
                                                     color={roleTypeConfig[role.role_type].color}>{role.name}</Tag>
                                            </Tooltip>);
                                    })}
                                </div>
                            </div>
                        }
                        {
                            showType === 'all' && <div style={{
                                display: 'flex',
                                flexDirection: 'column',
                                justifyContent: 'center',
                                alignItems: 'center',
                                width: '100%'
                            }}>
                                <span>Administrator</span>
                                <div>
                                    {
                                        item.users.map((user) => {
                                            return <Tooltip key={user.id} placement="bottomLeft"
                                                            title={`${user.fullname}是${item.fullname}的${roleTypeConfig[user.role_type].name}`}><Avatar
                                                shape="round" style={{
                                                marginLeft: -12,
                                                border: '1px solid white',
                                                backgroundColor: roleTypeConfig[user.role_type].color,
                                                fontWeight: 500
                                            }} size={40}><span
                                                style={{fontSize: 20}}>{user.fullname}</span></Avatar></Tooltip>
                                        })
                                    }
                                </div>
                            </div>
                        }
                    </List.Item>
                }}
            />

        </Card>);
    }
}

class CustomizedForm extends React.Component {
    render() {
        const {form: {getFieldDecorator}, form, onOk, onCancel, canSubmit} = this.props;
        return (
            <Card title="创建应用" bordered={false}>
                <Form>
                    <Form.Item label="应用名" labelCol={{span: 6}} wrapperCol={{span: 12}}>
                        {getFieldDecorator('fullname', {
                            rules: [
                                {required: true, message: 'Please input your client name!'},
                                {
                                    pattern: new RegExp(/^[a-zA-Z0-9_]+$/),
                                    message: 'client name can only contain (A-Z a-z 0-9 _) !'
                                },
                            ],
                        })(
                            <Input placeholder="client name"/>
                        )}
                    </Form.Item>
                    <Form.Item label="重定向地址" labelCol={{span: 6}} wrapperCol={{span: 12}}>
                        {getFieldDecorator('redirect_uri', {
                            rules: [
                                {required: true, message: 'Please input your client uri!'},
                                {
                                    pattern: new RegExp(/^([hH][tT]{2}[pP]:\/\/|[hH][tT]{2}[pP][sS]:\/\/)/),
                                    message: 'client redirect_uri must start with http:// or https !'
                                },
                            ],
                            initialValue: 'https://'
                        })(
                            <Input placeholder="client redirect_uri"/>
                        )}
                    </Form.Item>
                    <Form.Item>
                        <Row type="flex" justify="center" align="middle" gutter={20}>
                            <Button type="primary" ghost style={{margin: 10}} disabled={!canSubmit}
                                    onClick={() => onOk(form)}>创建</Button>
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
        ...state.client,
        loading: state.common.loading
    };
}


function mapDispatchToProps(dispatch, ownProps) {
    return {
        dispatch: dispatch,
    }
}

export default withRouter(connect(mapStateToProps, mapDispatchToProps)(Client));

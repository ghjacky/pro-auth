import React from 'react';
import {Card, Col, Row, Table, Input, Tag, Tooltip, Radio} from 'antd';
import {roleTypeConfig} from '../../utils/config';
import {withRouter} from "react-router-dom";
import {connect} from "react-redux";
import {queryUserRoles} from '../../redux/action/role';

const Search = Input.Search;
const RadioGroup = Radio.Group;
const RadioButton = Radio.Button;

class My extends React.Component {

    state = {
        showType: 'all',
        roles: [],
        rolesOrigin:[],
        resources: [],
        resourcesOrigin:[],
    };

    componentDidMount() {
        const {match} = this.props;
        const client_id = match.params.client_id;
        this.props.dispatch(queryUserRoles('myUserRoles', client_id, 'normal', true));
    }

    componentWillReceiveProps(nextProps) {
        const {myUserRoles} = nextProps;
        if (myUserRoles) {
            const resourcesMap = {};
            const resources = [];

            for (let i = 0; i < myUserRoles.length; i++) {
                if (myUserRoles[i].resources) {
                    for (let j = 0; j < myUserRoles[i].resources.length; j++) {
                        resourcesMap[myUserRoles[i].resources[j].resource.id] = myUserRoles[i].resources[j].resource;
                    }
                }
            }
            Object.values(resourcesMap).map(value => {
                resources.push(value);
                return value;
            });
            this.setState({roles: myUserRoles,rolesOrigin: myUserRoles, resources: resources, resourcesOrigin: resources});
        }

    }

    onChangeShowType = (value) => {
        this.setState({showType: value})
    };

    onSearch = (value) => {
        const {rolesOrigin, resourcesOrigin, showType} = this.state;
        this.setState({roles: rolesOrigin, resources: resourcesOrigin});
        if (showType === 'all') {
            let resources = resourcesOrigin.filter(resource => {
                if (resource.name.indexOf(value) > -1) {
                    return true;
                }
                if (resource.description.indexOf(value) > -1) {
                    return true;
                }
                if (resource.data.indexOf(value) > -1) {
                    return true;
                }
                return false;
            });
            this.setState({resources: resources})
        } else if (showType === 'by_role') {
            let roles = rolesOrigin.filter(role => {
                if (role.name.indexOf(value) > -1) {
                    return true;
                }
                if (role.description.indexOf(value) > -1) {
                    return true;
                }
                if ((role.role_type === 'super' && '超级管理员'.indexOf(value) > -1) || (role.role_type === 'admin' && '管理员'.indexOf(value) > -1) || (role.role_type === 'normal' && '普通成员'.indexOf(value) > -1)) {
                    return true;
                }
                if(role.resources) {
                    for (let i = 0; i < role.resources.length; i++) {
                        let resource = role.resources[i].resource;
                        if (resource.name.indexOf(value) > -1 ||resource.description.indexOf(value) > -1 ||resource.data.indexOf(value) > -1) {
                            return true;
                        }
                    }
                }
                return false
            });
            this.setState({roles: roles})
        }
    };

    render() {
        const {roles, resources, showType} = this.state;
        const {loading} = this.props;
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
                title: '你在该角色的身份',
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

        ];
        const columns2 = [
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
        ];
        const expandedRowRender = (record) => {
            const columns = [
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
            ];

            return (
                <Table
                    title={() => `${record.name}的权限`}
                    columns={columns}
                    dataSource={record.resources}
                    pagination={false}
                />
            );
        };
        const extraContent = (
            <Row type="flex" justify="start" align="middle" gutter={20}>
                <Col>
                    <Search style={{width: 500}} placeholder="搜索" onSearch={this.onSearch}/>
                </Col>
                <Col>
                    <RadioGroup defaultValue="all" onChange={(e) => this.onChangeShowType(e.target.value)}>
                        <RadioButton value="all">全部权限</RadioButton>
                        <RadioButton value="by_role">按角色查看</RadioButton>
                    </RadioGroup>
                </Col>
            </Row>
        );
        return (
            <div style={{margin: 20}}>
                <Card
                    bordered={false}
                    style={{overflow: 'visible'}}
                    title="我的权限"
                    headStyle={{border: 0}}
                    bodyStyle={{padding: '0 32px 40px 32px'}}
                    extra={extraContent}
                >
                    {
                        showType === 'all' && <Table
                            loading={loading}
                            columns={columns2}
                            dataSource={resources}
                        />
                    }
                    {
                        showType === 'by_role' && <Table
                            loading={loading}
                            columns={columns}
                            expandedRowRender={expandedRowRender}
                            dataSource={roles}
                        />
                    }

                </Card>
            </div>
        );
    }
}

function mapStateToProps(state) {
    return {
        myUserRoles: state.role.myUserRoles,
        loading: state.common.loading
    };
}


function mapDispatchToProps(dispatch, ownProps) {
    return {
        dispatch: dispatch,
    }
}

export default withRouter(connect(mapStateToProps, mapDispatchToProps)(My));
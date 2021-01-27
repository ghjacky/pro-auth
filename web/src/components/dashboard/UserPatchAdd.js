import React from 'react';
import {Button, Table, Select, Modal, Row, message, Input, Alert} from "antd";
import {roleTypeConfig} from "../../utils/config";
import {queryUserList} from '../../redux/action/common';
import request from "../../utils/request";
import {connect} from "react-redux";

const Option = Select.Option;
const Search= Input.Search;
class UserPatchAdd extends React.Component {
    state = {
        users: [],
        userList:[],
        showModal: false,
        selectedRowKeys:[],
        selectedRows:[],
        currentUser: {},
        existedUserMap: {},
        errMsg:'',
        canSubmit: true,
        searchText: '',
    };
    componentDidMount() {
        const {client_id} = this.props;
        this.props.dispatch(queryUserList(client_id));
    }
    componentWillReceiveProps(nextProps, nextContext) {
        const { userList, user, existedUsers} = nextProps;
        if(userList) {
            this.setState({userList});
        }
        if(user) {
            this.setState({user});
        }
        if(existedUsers) {
            const existedUserMap = {};
            for( let i = 0 ; i < existedUsers.length; i++ ) {
                existedUserMap[existedUsers[i].user.id] = existedUsers[i].user;
            }
            this.setState({existedUserMap});
        }
    }
    onColumnChange = (index, key, value) => {
        const {users} = this.state;
        users[index][key] = value;
        this.setState({users});
    };
    onDelete = (index) => {
        const {users} = this.state;
        users.splice(index, 1);
        this.setState({users});
    };
    onAddByMail = () => {
        const {users} = this.state;
        const newUser = {
            user_id: '',
            fullname: '未知',
            readOnly: false,
            role_type: 'normal'
        };
        users.push(newUser);
        this.setState({users});
    };
    changeModalVisible = () => {
        this.setState({showModal : !this.state.showModal});
    };
    onAddByUserList = () => {
        const { selectedRows, users} = this.state;
        console.log(selectedRows);
        for(let i = 0; i < selectedRows.length;i++) {
            let obj = {
                user_id: selectedRows[i].id,
                fullname: selectedRows[i].fullname,
                readOnly: true,
                role_type: 'normal',
            };
            users.push(obj);
        }
        this.setState({users, showModal: false});
    };
    okHandler = () => {
        const {onOk, onCancel, client_id, role_id} = this.props;
        const {users, existedUserMap} = this.state;
        let errMsg = '';
        for (let i = 0 ; i<users.length;i++) {
            if(existedUserMap[users[i].user_id] ) {
                errMsg += `${users[i].user_id}已经是该角色的成员\n`
            }
        }
        if (errMsg !== '') {
            this.setState({errMsg})
        } else {
            this.setState({canSubmit: false});
            request(`/api/roleUsers/${role_id}?client_id=${client_id}`,{
                method: 'POST',
                body: JSON.stringify(users),
            }).then(res => {
                this.setState({canSubmit: true});

                if(res.res_code === 0) {
                    message.success("添加成员成功！");
                    if(onOk) {
                        onOk();
                        onCancel();
                    }
                } else {
                    this.setState({errMsg: res.res_msg})
                }
            })
        }
    };
    onSearch = (value) => {
        this.setState({searchText: value, selectedRowKeys:[], selectedRows: []});
    };

    render() {
        const roleTypeRange = {
            super: true,
            admin: true,
            normal: true
        };
        const {users, userList,showModal,selectedRowKeys, canSubmit,errMsg, searchText} = this.state;
        const columns = [
            {
                title: '账号',
                dataIndex: 'user_id',
                key: 'user_id',
                render: (text, record, index) => {
                    return (<Input disabled={record.readOnly} placeholder="请输入用户账号" value={text}
                                   onChange={(e) => this.onColumnChange(index, 'user_id', e.target.value)}/>);
                }
            },
            {
                title: '姓名',
                dataIndex: 'fullname',
                key: 'fullname',
            },
            {
                title: '在该角色中的身份',
                dataIndex: 'role_type',
                key: 'role_type',
                render: (text, record, index) => {
                    return (<Select style={{width: '100%'}} value={text}
                                    onChange={(value) => this.onColumnChange(index, 'role_type', value)}>
                        {
                            Object.keys(roleTypeConfig).filter(key => roleTypeRange[key] === true).map(key => {
                                if(key !== 'super' || (this.props.isRootRoleSuper > 0 && key === 'super')) {
                                    return <Option key={key} value={key}>
                                        <div>{roleTypeConfig[key].name}</div>
                                    </Option>
                                } else {
                                    return null
                                }
                            })
                        }
                    </Select>);
                }
            },
            {
                title: '操作',
                key: 'operation',
                render: (text, record, index) => (
                    <Button type="danger" onClick={() => this.onDelete(index)}>删除</Button>
                ),
            },
        ];
        const columns2 = [
            {
                title: '姓名',
                dataIndex: 'fullname',
                key: 'fullname',
                filteredValue: [searchText],
                onFilter: (value, record) => {
                    return record.fullname.includes(value) || record.id.includes(value) || record.dn.includes(value);
                },
            },
            {
                title: '账号',
                dataIndex: 'id',
                key: 'id',
            },
            {
                title: 'dn',
                dataIndex: 'dn',
                key: 'dn',
                render: (text, record) => {
                    if (text === undefined) {
                        return '';
                    } else {
                        return text.replace(',OU=HABROOT,DC=creditease,DC=corp', '').replace(record.fullname+ ',', '').replace(/[(OU=|(CN=)]+/g,'').replace(/,/g,'_')
                    }
                },
            }
        ];
        const rowSelection = {
            selectedRowKeys,
            onChange: (selectedRowKeys, selectedRows) => {
                this.setState({selectedRowKeys, selectedRows});
            },
        };


        return (<div>
            <Button type="dashed"
                    style={{width: '45%', margin: 8}}
                    icon="plus"
                    onClick={this.changeModalVisible}
            >从用户列表添加</Button>
            <Button
                type="dashed"
                style={{width: '45%',margin:8}}
                icon="plus"
                onClick={this.onAddByMail}
            >手动填写账号添加</Button>
            <Table rowKey={(record)=> record.id} dataSource={users} columns={columns} pagination={false}/>
            {
                errMsg !== '' && <Alert message={errMsg.split('\n').map((err, index) => <p key={index}>{err}</p>)} type="error" />
            }
            <Row type="flex" justify="center">
                <Button type="primary" ghost style={{margin: 10}} onClick={this.okHandler} disabled={users.length === 0 || !canSubmit}>添加以上成员</Button>
                <Button type="primary" ghost style={{margin: 10}} onClick={this.props.onCancel}>取消</Button>
            </Row>
            <Modal
                visible={showModal}
                title="用户列表"
                onOk={this.onAddByUserList}
                onCancel={this.changeModalVisible}
                maskClosable={false}
                width={800}
                okText={'添加选中的用户'}
                bodyStyle={{padding:0}}
                okButtonProps={{ disabled: selectedRowKeys.length === 0 }}
            >
                <Search style={{margin: 10, width: '80%'}} onSearch={this.onSearch} placeholder={'搜索用户'}/>
                <Table rowKey={'id'}
                       rowSelection={rowSelection}
                       size="small" columns={columns2}
                       dataSource={userList}
                       pagination={{showTotal: (total) => `select ${selectedRowKeys.length} of ${total} users`}} />
            </Modal>
        </div>);
    }
}

function mapStateToProps(state) {
    return {
        userList: state.common.userList,
        user: state.common.user
    };
}


function mapDispatchToProps(dispatch, ownProps) {
    return {
        dispatch: dispatch,
    }
}

export default connect(mapStateToProps, mapDispatchToProps)(UserPatchAdd);

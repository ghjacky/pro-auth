import React from 'react';
import {Button, Col, Icon, Input, message, Row, Table} from 'antd';
import request from '../../utils/request';

const Search = Input.Search;
class RoleResourceTransfer extends React.Component {
    state = {
        leftData: [],
        leftSelectedRowKeys:[],
        leftSelectedRows:[],
        leftSearchText: '',

        rightData:[],
        rightSelectedRowKeys:[],
        rightSelectedRows:[],
        rightSearchText: '',

        canSubmit: true,
    };
    componentDidMount() {
        const {allData, selectedData} = this.props;
        this.setState({
            rightData: JSON.parse(JSON.stringify(selectedData)),
            leftData: JSON.parse(JSON.stringify(getSubtraction(allData,selectedData))),
        });
    }
    componentWillReceiveProps(nextProps, nextContext) {
        const {allData, selectedData} = nextProps;
        this.setState({
            rightData: JSON.parse(JSON.stringify(selectedData)),
            leftData: JSON.parse(JSON.stringify(getSubtraction(allData,selectedData))),
        });
    }
    okHandler = () => {
        const {onOk,onCancel, role_id, client_id} = this.props;
        const { rightData} = this.state;
        this.setState({canSubmit:false});
        request(`/api/roleResources/${role_id}?client_id=${client_id}`, {
            method: 'PUT',
            body: JSON.stringify(rightData.map(record => record.resource.id)),
        }).then(res => {
            this.setState({canSubmit: true});
            if(res.res_code === 0) {
                if(onOk) {
                    onOk();
                    onCancel()
                }
                message.success("关联权限成功！");
            } else {
                message.error(res.res_msg);
            }
        });
    };
    moveDataLeftToRight = () => {
        const {
            leftData,
            leftSelectedRows,
            rightData,
        } = this.state;
        const newLeftData = getSubtraction(leftData,leftSelectedRows);
        const newRightData = [];
        newRightData.push(...rightData, ...leftSelectedRows);
        this.setState({
            leftData: newLeftData,
            leftSelectedRowKeys: [],
            leftSelectedRows:[],
            rightData: newRightData,
        });
    };
    moveDataRightToLeft = () => {
        const {
            leftData,
            rightData,
            rightSelectedRows
        } = this.state;
        const newLeftData = [];
        const newRightData = getSubtraction(rightData,rightSelectedRows)
        newLeftData.push(...leftData,...rightSelectedRows );

        this.setState({
            leftData: newLeftData,
            rightSelectedRowKeys: [],
            rightSelectedRows:[],
            rightData: newRightData,
        });
    };

    onLeafSearch = (value) => {
        this.setState({ leftSearchText: value, leftSelectedRowKeys:[] , leftSelectedRows:[]})
    };

    onRightSearch = (value) => {
        this.setState({ rightSearchText: value, rightSelectedRowKeys:[] , rightSelectedRows:[]})
    };
    render() {
        const {
            canSubmit,
            leftData,
            leftSearchText,
            leftSelectedRowKeys,
            rightData,
            rightSearchText,
            rightSelectedRowKeys,
        } = this.state;
        const columns1 = [
            {
                title: 'id',
                dataIndex: 'resource.id',
            },
            {
                title: '可关联的权限',
                dataIndex: 'resource.name',
                filtered:true,
                filteredValue: [leftSearchText],
                onFilter: (value, record) => record.resource.name.toString().toLowerCase().includes(value.toLowerCase()),
            },
            {
                title: '说明',
                dataIndex: 'resource.description',
            }
        ];
        const columns2 = [
            {
                title: 'id',
                dataIndex: 'resource.id',
            },
            {
                title: '已关联的权限',
                dataIndex: 'resource.name',
                filtered:true,
                filteredValue: [rightSearchText],
                onFilter: (value, record) => record.resource.name.toString().toLowerCase().includes(value.toLowerCase()),
            },
            {
                title: '说明',
                dataIndex: 'resource.description',
            }
        ];
        const rowSelection1 = {
            selectedRowKeys: leftSelectedRowKeys,
            hideDefaultSelections: true,
            onChange:  (selectedRowKeys, selectedRows) => {
                this.setState({leftSelectedRowKeys:selectedRowKeys , leftSelectedRows:selectedRows});
            },
            selections: [
                {
                    key: 'all-data',
                    text: '选择全部',
                    onSelect: () => {
                        let keys = [];
                        let data = leftData.filter(record => {
                            if (record.resource.name.includes(leftSearchText)) {
                                keys.push(record.resource.id);
                                return true;
                            } else {
                                return false;
                            }
                        });
                        this.setState({
                            leftSelectedRowKeys: keys,
                            leftSelectedRows: data,
                        });
                    },
                },
                {
                    key: 'clear-data',
                    text: '清除全部',
                    onSelect: () => {
                        this.setState({
                            leftSelectedRowKeys: [],
                            leftSelectedRows:[],
                        });
                    },
                },
            ],
            onSelection: this.onSelection,
        };
        const rowSelection2 = {
            selectedRowKeys: rightSelectedRowKeys,
            hideDefaultSelections: true,
            onChange:  (selectedRowKeys, selectedRows) => {
                this.setState({rightSelectedRowKeys:selectedRowKeys, rightSelectedRows:selectedRows});
            },
            selections: [
                {
                    key: 'all-data',
                    text: '选择全部',
                    onSelect: () => {
                        let keys = [];
                        let data = rightData.filter(record => {
                           if (record.resource.name.includes(rightSearchText)) {
                               keys.push(record.resource.id)
                               return true;
                           } else {
                               return false;
                           }

                        });
                        this.setState({
                            rightSelectedRowKeys: keys,
                            rightSelectedRows: data,
                        });
                    },
                },
                {
                    key: 'clear-data',
                    text: '清除全部',
                    onSelect: () => {
                        this.setState({
                            rightSelectedRowKeys: [],
                            rightSelectedRows:[],
                        });
                    },
                },
            ],
            onSelection: this.onSelection,
        };
        return (
            <div>
                <Row type="flex" align="middle" justify="center" gutter={10} >
                    <Col span={11}>
                        <Table rowKey={(record)=> record.resource.id}
                               title={()=> <Search placeholder="搜索可关联的权限" onSearch={this.onLeafSearch}/> }
                               rowSelection={rowSelection1}
                               columns={columns1}
                               dataSource={leftData}
                               size="middle"
                               pagination={{showTotal: (total) => `select ${leftSelectedRowKeys.length} of ${total}`}}
                        />
                    </Col>
                    <Col span={2} style={{display: 'flex', flexDirection: 'column', justifyContent: 'center',alignItems: 'center'}}>
                        <Button style={{ margin: 10}}
                                type="primary"
                                disabled={leftSelectedRowKeys.length === 0}
                                size="small"
                                onClick={this.moveDataLeftToRight}
                        >
                            <Icon type="right" />
                        </Button>
                        <Button style={{ margin: 10}}
                                type="primary"
                                disabled={rightSelectedRowKeys.length === 0}
                                size="small"
                                onClick={this.moveDataRightToLeft}
                        >
                            <Icon type="left" />
                        </Button>
                    </Col>
                    <Col span={11}>
                        <Table rowKey={(record)=> record.resource.id}
                               title={()=> <Search placeholder="搜索已关联的权限" onSearch={this.onRightSearch}/> }
                               rowSelection={rowSelection2}
                               columns={columns2}
                               dataSource={rightData}
                               size="middle"
                               pagination={{showTotal: (total) => `select ${rightSelectedRowKeys.length} of ${total}`}}
                        />
                    </Col>
                </Row>
                <div style={{margin: 10}}>注意：如果删除了关联的权限，那么其子角色也会删除相应的权限</div>
                <Row type="flex" justify="center">
                    <Button type="primary" ghost style={{margin: 10}} disabled={!canSubmit} onClick={this.okHandler}>保存</Button>
                    <Button type="primary" ghost style={{margin: 10}} onClick={this.props.onCancel}>取消</Button>
                </Row>
            </div>
        );
    }
}

function getSubtraction(allset, subset) {
   let connectMap = {};
   let res = [];
   for (let i = 0; i< subset.length; i++) {
       connectMap[subset[i].resource.id] = true
   }
    for (let i = 0; i< allset.length; i++) {
       if (connectMap[allset[i].resource.id] !== true) {
           res.push(allset[i]);
       }
    }
   return res
}

export default RoleResourceTransfer;

import React from 'react';
import {Tree, Input, message, Button, Row, Tooltip} from 'antd';
import '../../asset/icon/iconfont.css';
import {roleTypeConfig} from "../../utils/config";

const TreeNode = Tree.TreeNode;
const Search = Input.Search;

const getParentKey = (key, tree) => {
    let parentKey;
    for (let i = 0; i < tree.length; i++) {
        const node = tree[i];
        if (node.children) {
            if (node.children.some(item => item.key === key)) {
                parentKey = node.key;
            } else if (getParentKey(key, node.children)) {
                parentKey = getParentKey(key, node.children);
            }
        }
    }
    return parentKey;
};

const getKey = (tree, dataList, expandedKeys, depth) => {
    if (tree === undefined) {
        return
    }
    tree.map((item) => {
        if (dataList !== undefined) {
            dataList.push(item);
        }
        if (expandedKeys !== undefined && depth > 0) {
            expandedKeys.push(item.id +'');
        }
        if (item.children) {
            getKey(item.children, dataList, expandedKeys, item.role_type ? depth - 1: depth );
        }
        return item;
    });
};


class RoleTree extends React.Component {
    state = {
        expandedKeys: [],
        searchValue: '',
        autoExpandParent: true,
        treeNodes: [],
        dataList: [],
        openAll: false,
        selectedKeys:[]
    };

    componentWillReceiveProps(nextProps) {
        const {treeNodes} = nextProps;
        if (JSON.stringify(treeNodes) !== JSON.stringify(this.props.treeNodes)) {
            const dataList = [];
            const expandedKeys = [];
            getKey(treeNodes, dataList, expandedKeys, 1);
            if (this.state.expandedKeys.length === 0) {
                this.setState({expandedKeys});
            }
            this.setState({dataList, treeNodes});

        }

    }
    onExpand = (expandedKeys) => {
        this.setState({
            expandedKeys,
            autoExpandParent: false,
        });
    };
    onSearch = (value) => {
        const {dataList} = this.state;
        const expandedKeys = dataList.map((item) => {
            if (item.name.indexOf(value) > -1 && item.role_type !== undefined) {
                return item.id +'';
            } else {
                return null;
            }
        }).filter((key) => key !== null);
        if (expandedKeys.length > 0) {
            this.setState({
                expandedKeys,
                searchValue: value,
                autoExpandParent: true,
            });
        } else {
            message.info('未找到符合的角色或无对应角色权限');
        }
    };
    openAll = () => {
        const {dataList} = this.state;
        const expandedKeys = dataList.map(item => item.id +'');
        this.setState({expandedKeys});
    };
    cancelOpenAll = ()=> {
        const {treeNodes} = this.state;
        const expandedKeys = [];
        getKey(treeNodes, undefined, expandedKeys, 1);
        this.setState({expandedKeys});
    };
    renderTreeNodes = (treeNode) => {
        if (treeNode === undefined) {
            return null
        }
        return treeNode.map((item) => {
            const index = item.name.indexOf(this.state.searchValue);
            const beforeStr = item.name.substr(0, index);
            const afterStr = item.name.substr(index + this.state.searchValue.length);
            const title = index > -1 && item.role_type !== undefined ? (
                <span>
          {beforeStr}
                    <span style={{color: '#108ee9'}}>{this.state.searchValue}</span>
                    {afterStr}
        </span>
            ) : <span>{item.name}</span>;
            let icon = <i className="icon iconfont icon-mwuquanxian"></i>;
            if (item.role_type === 'super') {
                icon = <i className="icon iconfont icon-huangguan"></i>;
            } else if (item.role_type === 'admin') {
                icon = <i className="icon iconfont icon-guanliyuan"></i>;
            } else if(item.role_type === 'normal') {
                icon = <i className="icon iconfont icon-chengyuan"></i>;
            }
            const showText = roleTypeConfig[item.role_type]? `你在该角色下的身份是${roleTypeConfig[item.role_type].name}，${roleTypeConfig[item.role_type].effect}`: '没有权限';
            if (item.children) {
                return (
                    <TreeNode icon={icon} role_type={item.role_type} title={
                        <Tooltip mouseEnterDelay={1} placement="bottomLeft"
                                 title={showText}>
                            {title}
                        </Tooltip>
                    } key={item.id + ''} >
                        {this.renderTreeNodes(item.children)}
                    </TreeNode>

                );
            }
            return (
                <TreeNode icon={icon} role_type={item.role_type} title={<Tooltip mouseEnterDelay={1} placement="bottomLeft" title={showText}>{title}</Tooltip>} key={item.id + ''} disabled={item.role_type === undefined}/>
           );
        });
    };

    render() {
        const {title, treeNodes} = this.props;
        const {expandedKeys, autoExpandParent, selectedKeys} = this.state;
        return (
            <div style={{marginRight: '20px', minWidth: '200px'}}>
                <div style={{
                    textAlign: 'center',
                    fontSize: 16,
                    fontWeight: 500,
                    color: 'rgba(0,0,0,0.85)'
                }}>{title}</div>
                <Row type="flex" justify="space-between" align="middle">
                    <Button type="primary" size="small" style={{width: '45%'}} onClick={this.openAll}
                            ghost>展开全部</Button>
                    <Button type="primary" size="small" style={{width: '45%'}} onClick={this.cancelOpenAll}
                            ghost>恢复默认</Button>
                </Row>
                <Search style={{marginBottom: 8, marginTop: 8}} placeholder="Search" onSearch={this.onSearch}/>
                <Tree
                    showIcon
                    selectedKeys={selectedKeys}
                    onExpand={this.onExpand}
                    expandedKeys={expandedKeys}
                    autoExpandParent={autoExpandParent}
                    onSelect={
                        (selectedKeys, e) => {
                            if(e.node.props.role_type === undefined) {
                                this.setState({selectedKeys:[]});
                                this.props.onSelect([]);
                            } else {
                                this.setState({selectedKeys});
                                this.props.onSelect(selectedKeys);
                            }
                        }
                    }
                >
                    {this.renderTreeNodes(treeNodes)}
                </Tree>
            </div>
        );
    }

}

export default RoleTree;

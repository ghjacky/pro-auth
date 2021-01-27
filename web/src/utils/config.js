module.exports = {
    roleTypeConfig: {
        "super":{
            name: '超级管理员',
            color: '#87CEFA',
            effect: '能够管理该角色子树的组织结构并使用该角色的权限'
        },
        "admin":{
            name:'管理员',
            color: '#ADD8E6',
            effect: '能够管理该角色子树的成员并使用该角色的权限'
        },
        "normal":{
            name: '普通成员',
            color: '',
            effect: '能够使用该角色的权限'
        }
    },

    userStatusConfig: {
        "active": {
            name: '正常',
            color: '#87CEFA',
            effect: '用户账号正常登录和使用'
        },
        "frozen": {
            name: '冻结',
            color: '#b3b2b4',
            effect: '账号已被冻结请联系管理员解冻'
        },
        "delete": {
            name: '删除',
            color: '#e65e41',
            effect: '账号已被删除无法继续使用'
        }
    }
};



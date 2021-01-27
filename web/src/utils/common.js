export function buildTree(roleList, roleMap = {}) {
    if(roleList === undefined) {
        return [];
    }
    roleList.sort((x, y) => x.id - y.id);
    const resRoleMap = {};
    for (let i = 0 ; i < roleList.length; i++) {
        if(roleList[i].resources === undefined) {
            roleList[i].resources = [];
        }
        if(roleList[i].users === undefined) {
            roleList[i].usrs = [];
        }
        roleList[i].children = [];
        roleMap[roleList[i].id] = roleList[i];
        resRoleMap[roleList[i].id] = roleList[i];
    }
    let rootList =[];
    for (let i = 0 ; i < roleList.length; i++) {
        if (roleMap[roleList[i].parent_id] === undefined) {
            rootList.push(roleList[i]);
            delete resRoleMap[roleList[i].id];
        }
    }
    return doBuildTree(rootList, resRoleMap);
}


function doBuildTree(rootList, resRoleMap) {
    for(let i =0 ; i <rootList.length; i++) {
        let children = [];
        for(let k in resRoleMap) {
            if (resRoleMap[k].parent_id === rootList[i].id) {
                children.push(resRoleMap[k]);
                delete resRoleMap[k];
            }
        }
        if (children.length > 0) {
            rootList[i].children = doBuildTree(children, resRoleMap);
        }
    }
    return rootList;
}

import React from 'react';
import Exception from 'ant-design-pro/lib/Exception';
import {Link} from 'react-router-dom';

import 'ant-design-pro/dist/ant-design-pro.css';
import {Button} from 'antd';

class NoAuthorized extends React.Component {
    render() {
        const actions = (
            <div>
                <Button type="primary"><Link to={"/"}>Home</Link></Button>
            </div>
        );
        return <Exception type="403" className="exception-content" actions={actions}/>
    }
}

export default NoAuthorized;
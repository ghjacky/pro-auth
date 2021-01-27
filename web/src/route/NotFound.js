import React from 'react';
import Exception from 'ant-design-pro/lib/Exception';
import 'ant-design-pro/dist/ant-design-pro.css';

class NotFound extends React.Component {
    render() {
        return (
            <div>
                <Exception className="exception-content" type="404"/>
            </div>
        );
    }
}

export default NotFound;
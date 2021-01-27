import React from 'react';
import {Modal, Form, Input, message, Alert, Select} from 'antd';
import request from '../../utils/request';
import {roleTypeConfig} from '../../utils/config';
const FormItem = Form.Item;
const Option = Select.Option;
class UserUpdateRoleTypeModal extends React.Component {
    state = {
        visible: false,
        canSubmit: true,
        errMsg: '',
    };
    showModelHandler = (e) => {
        if (e) e.stopPropagation();
        this.props.form.resetFields();
        this.setState({
            visible: true,
        });
    };

    hideModelHandler = () => {
        this.setState({
            visible: false,
        });
    };
    okHandler = () => {
        const {onOk, role_id, client_id, user} = this.props;
        this.props.form.validateFields((err, values) => {
            if (!err) {
                this.setState({errMsg: '', canSubmit: false});
                request(`/api/roleUsers/${role_id}?client_id=${client_id}`, {
                    method: 'PUT',
                    body: JSON.stringify({
                        user_id: user.id,
                        role_type: values.role_type
                    })
                }, false).then(res => {
                    this.setState({canSubmit: true});
                    if (res.res_code === 0) {
                        message.success("更新成员身份成功！");
                        this.hideModelHandler();
                        if (onOk) {
                            onOk();
                        }
                    } else {
                        this.setState({errMsg: res.res_msg});
                    }
                });
            }
        });
    };

    render() {
        const {user, children, form: {getFieldDecorator}, isRootRoleSuper} = this.props;
        const {canSubmit, errMsg} = this.state;
        return (
            <span>
        <span onClick={this.showModelHandler}>
          {children}
        </span>
        <Modal
            maskClosable={false}
            title="修改角色信息"
            visible={this.state.visible}
            onCancel={this.hideModelHandler}
            onOk={this.okHandler}
            okButtonProps={{disabled: !canSubmit}}
            cancelButtonProps={{disabled: !canSubmit}}
        >
          <Form layout="inline">
            <FormItem>
                <Input disabled value={user.fullname + `(${user.id})`}/>
            </FormItem>
              <FormItem>
                  {getFieldDecorator('role_type', {
                      initialValue: user.role_type,
                      rules:[{required: true, message: 'Please select one identity !'}]
                  })(
                      <Select style={{width: 120}}>
                          { Object.keys(roleTypeConfig).map(key =>{
                                if(key === 'super') {
                                    if (isRootRoleSuper) {
                                        return <Option key={key} value={key}> {roleTypeConfig[key].name}</Option>;
                                    } else {
                                        return null;
                                    }
                                } else {
                                    return <Option key={key} value={key}> {roleTypeConfig[key].name}</Option>
                                }
                          } )}
                      </Select>
                  )}
              </FormItem>
          </Form>
            {
                errMsg !== '' && <Alert message={errMsg} type="error"/>
            }
        </Modal>
      </span>
        );
    }
}

export default Form.create()(UserUpdateRoleTypeModal);



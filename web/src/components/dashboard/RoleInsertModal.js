import React from 'react';
import {Modal, Form, Input, Select, message, Alert} from 'antd';
import request from "../../utils/request";

const FormItem = Form.Item;
const TextArea = Input.TextArea;
const Option = Select.Option;
class RoleInsertModal extends React.Component {
    state = {
        visible: false,
        canSubmit: true,
        errMsg: ''
    };
    showModelHandler = (e) => {
        if (e) e.stopPropagation();
        this.setState({
            visible: true,
        });
    };

    hideModelHandler = () => {
        if (!this.state.canSubmit) {
            return
        }
        this.setState({
            visible: false,
        });
    };
    okHandler = () => {
        const {onOk, client_id, parentRole} = this.props;
        this.props.form.validateFields((err, values) => {
            if (!err) {
                this.setState({errMsg:'', canSubmit: false});
                request(`/api/roles/${values.children_ids.join(',')}?client_id=${client_id}`,{
                    method: 'POST',
                    body: JSON.stringify({
                        name: values.name,
                        description: values.description,
                        parent_id: parentRole.id
                    })
                }, false).then(res =>{
                    this.setState({canSubmit: true});
                    if(res.res_code === 0) {
                        message.success("插入子角色成功！");
                        this.hideModelHandler();
                        if(onOk) {
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
        const {children, form: {getFieldDecorator}, parentRole} = this.props;
        const {errMsg, canSubmit} = this.state;
        let roleList = [];
        if(parentRole.children) {
            roleList = parentRole.children;
        }
        return (
            <span>
        <span onClick={this.showModelHandler}>
          {children}
        </span>
        <Modal
            maskClosable={false}
            title="插入子角色"
            visible={this.state.visible}
            onCancel={this.hideModelHandler}
            onOk={this.okHandler}
            okButtonProps={{ disabled: !canSubmit }}
            cancelButtonProps={{ disabled: !canSubmit }}
        >
          <Form>
            <FormItem>
                在{parentRole.name}与以下子角色之间插入新角色
            </FormItem>
            <FormItem>
                 {getFieldDecorator('children_ids', {
                     rules:[{required: true, message: 'Please select one role !'}]
                 })(
                     <Select
                         mode="multiple"
                         placeholder="选择子角色"
                     >
                         { roleList.map(role => <Option key={role.id} value={role.id}> {role.name}</Option>)}
                     </Select>
                 )}
            </FormItem>
            <FormItem label="角色名">
              {getFieldDecorator('name', {
                  rules: [{required: true, message: 'Please input the data of roleName !'}],
              })(
                  <Input/>)}
            </FormItem>
            <FormItem label="角色说明">
              {getFieldDecorator('description', {
                  rules: [{required: true, message: 'Please input the data of description !'}],
              })(
                  <TextArea autosize={{minRows: 2, maxRows: 6}}/>)}
            </FormItem>
          </Form>
            {
                errMsg !== '' && <Alert message={errMsg} type="error" />
            }
        </Modal>
      </span>
        );
    }
}

export default Form.create()(RoleInsertModal);



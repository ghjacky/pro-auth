import React from 'react';
import {Modal, Form, Input, message, Alert} from 'antd';
import request from "../../utils/request";

const FormItem = Form.Item;
const TextArea = Input.TextArea;

class ResourceAddModal extends React.Component {
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
        const {onOk, client_id} = this.props;
        this.props.form.validateFields((err, values) => {
            if (!err) {
                this.setState({errMsg: '', canSubmit: false});
                request(`/api/resources?client_id=${client_id}`, {
                    method: 'POST',
                    body: JSON.stringify([values])
                }, false).then(res => {
                    this.setState({canSubmit: true});
                    if (res.res_code === 0) {
                        message.success("新增权限成功！");
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
        const {children, form: {getFieldDecorator}} = this.props;
        const {canSubmit, errMsg} = this.state;
        return (
            <span>
        <span onClick={this.showModelHandler}>
          {children}
        </span>
        <Modal
            maskClosable={false}
            title="新增权限"
            visible={this.state.visible}
            onCancel={this.hideModelHandler}
            onOk={this.okHandler}
            okButtonProps={{disabled: !canSubmit}}
            cancelButtonProps={{disabled: !canSubmit}}
        >
          <Form>
            <FormItem label="权限名">
              {getFieldDecorator('name', {
                  rules: [{required: true, message: 'Please input the data of name !'}],
              })(
                  <Input/>)}
            </FormItem>
            <FormItem label="权限说明">
              {getFieldDecorator('description', {
                  rules: [{required: true, message: 'Please input the data of description !'}],
              })(
                  <Input />)}
            </FormItem>
            <FormItem label="权限内容">
              {getFieldDecorator('data', {
                  rules: [{required: true, message: 'Please input the data of content !'}],
              })(
                  <TextArea autosize={{minRows: 2, maxRows: 6}}/>)}
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

export default Form.create()(ResourceAddModal);



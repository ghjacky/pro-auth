import React from 'react';
import {Modal, Form, Input, message,Alert} from 'antd';
import request from '../../utils/request';
const FormItem = Form.Item;

class ClientEditModal extends React.Component {
    state = {
        visible: false,
        canSubmit: true,
        errMsg:'',
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
        const {onOk, client} = this.props;
        this.props.form.validateFields((err, values) => {
            if (!err) {
                this.setState({errMsg:'', canSubmit: false});
                request(`/api/client?client_id=${client.id}`,{
                    method: 'PUT',
                    body: JSON.stringify(values)
                }, false).then(res =>{
                    this.setState({canSubmit: true});
                    if(res.res_code === 0) {
                        message.success("更新应用成功！");
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
        const {children, client, form: {getFieldDecorator} } = this.props;
        const { canSubmit, errMsg } = this.state;
        return (
            <span>
        <span onClick={this.showModelHandler}>
          {children}
        </span>
        <Modal
            maskClosable={false}
            title="修改应用信息"
            visible={this.state.visible}
            onCancel={this.hideModelHandler}
            onOk={this.okHandler}
            okButtonProps={{ disabled: !canSubmit }}
            cancelButtonProps={{ disabled: !canSubmit }}
        >
          <Form>
            <FormItem label="client_id">
                <Input disabled value={client.id}/>
            </FormItem>
            <FormItem label="应用名">
              {getFieldDecorator('fullname', {
                  initialValue: client.fullname,
                  rules: [{required: true, message: 'Please input the fullname!'}],
              })(
                  <Input/>)}
            </FormItem>
            <FormItem label="redirect_uri">
              {getFieldDecorator('redirect_uri', {
                  initialValue: client.redirect_uri,
                  rules: [{required: true, message: 'Please input the redirect_uri !'}],
              })(
                  <Input/>)}
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

export default Form.create()(ClientEditModal);



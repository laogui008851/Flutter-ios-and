import { update_info } from '@/services/server/user_manage';
import { PageContainer } from '@ant-design/pro-components';
import { useModel } from '@umijs/max';
import { Button, Card, Form, Input, message } from 'antd';
import md5 from 'md5';
import { useEffect, useState } from 'react';

type FormFields = {
  old_password: string;
  password: string;
  password2: string;
};

const Profile = () => {
  const [form] = Form.useForm<FormFields>();
  const { initialState, setInitialState } = useModel('@@initialState');
  console.log(initialState?.currentUser);

  const [userInfo] = useState({
    nickname: initialState?.currentUser?.nickname ?? '',
    faceURL: initialState?.currentUser?.faceURL ?? '',
    password: initialState?.currentUser?.password ?? '',
  });

  const [mathPassword, setMathPassword] = useState(true);
  const password = Form.useWatch('password', form);
  const password2 = Form.useWatch('password2', form);
  useEffect(() => {
    if (password === password2) {
      setMathPassword(true);
    } else {
      setMathPassword(false);
    }
  }, [password, password2]);

  const onFinish = async (e: FormFields) => {
    try {
      await update_info({
        ...userInfo,
        userID: localStorage.getItem('IMAdminUserID')!,
        password: md5(e.password),
      });
      await setInitialState((s) => ({
        ...s,
        currentUser: {
          ...s?.currentUser,
          nickname: userInfo.nickname,
          faceURL: userInfo.faceURL,
          password: e.password,
        },
      }));
      message.success('修改成功');
    } catch (error) {
      message.error('修改失败！');
    }
  };

  return (
    <PageContainer>
      <Card>
        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          autoComplete="off"
          className="mt-4"
          style={{ width: '300px' }}
        >
          <Form.Item label="新密码" name="password">
            <Input.Password allowClear placeholder="请输入您的密码" />
          </Form.Item>

          <Form.Item
            label="确认密码"
            name="password2"
            validateStatus={mathPassword ? '' : 'error'}
            help={mathPassword ? '' : '两次输入的密码不一致!'}
            className="mb-8"
          >
            <Input.Password allowClear placeholder="请再次确认您的密码" />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" block disabled={!mathPassword}>
              保存
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </PageContainer>
  );
};

export default Profile;

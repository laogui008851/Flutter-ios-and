import { generateInvitationCode, getInvitationCode, setInvitationCodeAllowResetPassword } from '@/services/server/register_manage';
import { copy2Text } from '@/utils/common';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { PageContainer, ProTable } from '@ant-design/pro-components';
import { Button, InputNumber, message, Modal, Switch } from 'antd';
import { useRef, useState } from 'react';
import type { InvitationCodeItem } from './data';

const InvitionCode = () => {
  const actionRef = useRef<ActionType>();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [genNumber, setGenNumber] = useState(10);
  const [dataSource, setDataSource] = useState<InvitationCodeItem[]>([]);

  const copyAllCodes = () => {
    if (dataSource.length === 0) {
      message.warning('暂无邀请码可复制');
      return;
    }
    const allCodes = dataSource.map((item) => item.invitationCode).join('\n');
    if (navigator.clipboard) {
      navigator.clipboard.writeText(allCodes).then(() => {
        message.success(`已复制 ${dataSource.length} 个邀请码`);
      }).catch(() => {
        message.error('复制失败');
      });
    } else {
      const textarea = document.createElement('textarea');
      document.body.appendChild(textarea);
      textarea.style.position = 'fixed';
      textarea.style.clip = 'rect(0 0 0 0)';
      textarea.style.top = '10px';
      textarea.value = allCodes;
      textarea.select();
      document.execCommand('copy', true);
      document.body.removeChild(textarea);
      message.success(`已复制 ${dataSource.length} 个邀请码`);
    }
  };

  const handleAllowResetPasswordChange = async (code: string, checked: boolean) => {
    try {
      await setInvitationCodeAllowResetPassword(code, checked ? 1 : 0);
      message.success('设置成功');
      actionRef.current?.reload();
    } catch (error) {
      message.error('设置失败');
    }
  };

  const columns: ProColumns<InvitationCodeItem>[] = [
    {
      key: 'index',
      title: '序号',
      dataIndex: 'index',
      valueType: 'indexBorder',
      width: 48,
      align: 'center',
    },
    {
      title: '邀请码',
      dataIndex: 'invitationCode',
      key: 'invitationCode',
      align: 'center',
    },
    {
      title: '使用人昵称',
      key: 'usedUserName',
      dataIndex: ['usedUser', 'nickname'],
      hideInSearch: true,
      align: 'center',
      render: (_, record: any) => record.usedUser?.nickname || '-',
    },
    {
      title: '使用状态',
      key: 'status',
      dataIndex: 'status',
      valueType: 'select',
      hideInTable: true,
      initialValue: '全部',
      valueEnum: {
        2: '未使用',
        1: '已使用',
        0: '全部',
      },
    },
    {
      title: '使用人ID',
      key: 'usedUserID',
      dataIndex: 'usedUserID',
      hideInSearch: true,
      align: 'center',
      render: (_, record: any) => record.usedUserID || '-',
    },
    {
      title: '允许找回密码',
      key: 'allowResetPassword',
      dataIndex: 'allowResetPassword',
      hideInSearch: true,
      align: 'center',
      render: (_, record: any) => {
        // 只有已使用的邀请码才能设置为允许找回密码
        const isUsed = !!record.usedUserID;
        return (
          <Switch
            checked={record.allowResetPassword === 1}
            disabled={!isUsed}
            onChange={(checked) => handleAllowResetPasswordChange(record.invitationCode, checked)}
          />
        );
      },
    },
    {
      title: '操作',
      valueType: 'option',
      key: 'option',
      render: (text, record) => {
        return [
          <a href="#" key="copy" onClick={() => copy2Text(record.invitationCode)}>
            复制
          </a>,
        ];
      },
    },
  ];

  const showModal = () => {
    setIsModalOpen(true);
  };

  const genCode = () => {
    generateInvitationCode(genNumber).then(() => {
      message.success('生成成功！');
      actionRef.current?.reload();
    });
    // .catch(() => message.error('生成失败！'));
  };

  const handleOk = () => {
    setIsModalOpen(false);
    genCode();
  };

  const handleCancel = () => {
    setIsModalOpen(false);
  };

  return (
    <PageContainer>
      <ProTable<InvitationCodeItem>
        columns={columns}
        actionRef={actionRef}
        cardBordered
        columnsState={{
          defaultValue: {
            option: {
              fixed: 'right',
            },
          },
        }}
        request={async (params = {}, sort, filter) => {
          console.log(params, sort, filter);
          if (!params.status) {
            params.status = 0;
          }

          if (params.invitationCode) {
            params.keyword = params.invitationCode;
          }

          const { data } = await getInvitationCode({
            ...(params as any),
            status: Number(params.status) ?? 0,
            userIDs: [],
            pagination: {
              pageNumber: params.current as number,
              showNumber: params.pageSize,
            },
          });
          // const tmpData = data.chat_logs ?? [];
          setDataSource(data.list || []);
          return {
            data: data.list,
            success: true,
            total: data.total,
          };
        }}
        rowKey="invitationCode"
        search={{
          labelWidth: 'auto',
        }}
        pagination={{
          pageSize: 10,
          showSizeChanger: true,
          pageSizeOptions: [10, 20, 50, 100],
          showTotal: (total) => `共 ${total} 条`,
        }}
        dateFormatter="string"
        scroll={{ x: 'max-content' }}
        toolbar={{
          actions: [
            <Button key="copyAll" onClick={copyAllCodes}>
              整体复制
            </Button>,
            <Button key="create" type="primary" onClick={showModal}>
              生成邀请码
            </Button>,
          ],
          settings: [],
        }}
      />
      <Modal title="请选择生成数量" open={isModalOpen} onOk={handleOk} onCancel={handleCancel}>
        <div className="w-full flex justify-center">
          <InputNumber
            min={1}
            max={100}
            value={genNumber}
            onChange={(value) => {
              setGenNumber(value as number);
            }}
          />
        </div>
      </Modal>
    </PageContainer>
  );
};

export default InvitionCode;

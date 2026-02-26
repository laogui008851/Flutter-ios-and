import OIMAvatar from '@/components/OIMAvatar';
import { API_URL, WS_URL } from '@/config';
import { getUserToken } from '@/services/server/user_manage';
import { CloseOutlined, SearchOutlined } from '@ant-design/icons';
import { useLatest } from '@ant-design/pro-components';
import { Button, DatePicker, Input, Modal, Space, Spin } from 'antd';
import classNames from 'classnames';
import { throttle } from 'lodash';
import moment, { Moment } from 'moment';
import { getSDK } from 'open-im-sdk-wasm';
import { CbEvents } from 'open-im-sdk-wasm/lib/constant';
import { ConversationItem, MessageItem } from 'open-im-sdk-wasm/lib/types/entity';
import { SessionType } from 'open-im-sdk-wasm/lib/types/enum';
import { ListRef } from 'rc-virtual-list';
import { FC, useEffect, useRef, useState } from 'react';
import { Virtuoso, VirtuosoHandle } from 'react-virtuoso';
import HistoryMessageItem from './HistoryMessageItem';
import styles from './style.less';
import { canSearchMessageTypes } from './utils';

const openIM = getSDK();

type PreviewMessageModalProps = {
  closeModal: () => void;
  userID: string;
};

const START_INDEX = 10000;
const INITIAL_ITEM_COUNT = 20;

const PreviewMessageModal: FC<PreviewMessageModalProps> = ({ closeModal, userID }) => {
  const virtuoso = useRef<VirtuosoHandle>(null);
  const virtuosoHeight = useRef(0);
  const [firstItemIndex, setFirstItemIndex] = useState(START_INDEX);
  const [syncLoading, setSyncLoading] = useState(false);
  const [getMessageLoading, setGetMessageLoading] = useState(false);
  const [conversationSearchStr, setConversationSearchStr] = useState('');
  const [isSearching, setIsSearching] = useState(false);
  const [conversationList, setConversationList] = useState([] as ConversationItem[]);
  const [currentConversation, setCurrentConversation] = useState<ConversationItem>();
  const [messageList, setMessageList] = useState([] as MessageItem[]);
  const [offSetParams, setOffSetParams] = useState({
    startClientMsgID: '',
    lastMinSeq: 0,
    hasMore: true,
  });
  const latestOffsetParams = useLatest(offSetParams);
  const [searchParams, setSearchParams] = useState({
    keywords: '',
    timeRange: [] as number[],
    compRange: [],
    timeButton: '',
  });
  const latestSearchParams = useLatest(searchParams);
  const [searchOffset, setSearchOffset] = useState({
    pageIndex: 1,
    loading: false,
    hasMore: true,
  });
  const latestSearchOffset = useLatest(searchOffset);
  const vRef = useRef<ListRef>(null);

  const loginIM = (userID: string) => {
    getUserToken(userID).then(({ data }) => {
      openIM
        .login({
          userID,
          token: data.token,
          apiAddr: API_URL,
          wsAddr: WS_URL,
          platformID: 10,
        })
        .then((res) => console.log(res))
        .catch((err) => console.log(err));
    });
  };

  const onCancel = async () => {
    openIM.logout();
    closeModal();
  };

  const syncStartHandler = () => {
    setSyncLoading(true);
  };
  const syncFinishHandler = () => {
    openIM.getAllConversationList().then(({ data }) => {
      setConversationList(data);
      setSyncLoading(false);
    });
  };

  const setIMListener = () => {
    openIM.on(CbEvents.OnSyncServerStart, syncStartHandler);
    openIM.on(CbEvents.OnSyncServerFinish, syncFinishHandler);
  };

  const disposeIMListener = () => {
    openIM.off(CbEvents.OnSyncServerStart, syncStartHandler);
    openIM.off(CbEvents.OnSyncServerFinish, syncFinishHandler);
  };

  useEffect(() => {
    setIMListener();
    loginIM(userID);
    return () => {
      disposeIMListener();
    };
  }, []);

  const scrollToMsg = (id?: string): Promise<void> => {
    return new Promise((resolve) => {
      if (!id) {
        resolve();
        return;
      }
      setTimeout(() => {
        vRef.current?.scrollTo({
          key: id!,
          align: 'top',
        });
        resolve();
      });
    });
  };

  const scrollToBottom = (id: string): Promise<void> => {
    return new Promise((resolve) => {
      setTimeout(() => {
        vRef.current?.scrollTo({
          key: id,
          align: 'bottom',
        });
        resolve();
      });
    });
  };

  const getMessageNomaly = (item?: ConversationItem) => {
    setGetMessageLoading(true);
    openIM
      .getAdvancedHistoryMessageList({
        userID: '',
        groupID: '',
        count: 20,
        startClientMsgID: latestOffsetParams.current.startClientMsgID,
        conversationID: item ? item.conversationID : currentConversation?.conversationID ?? '',
        lastMinSeq: latestOffsetParams.current.lastMinSeq,
      })
      .then(({ data }) => {
        const historyData = data;
        console.log(historyData);
        setFirstItemIndex((idx) => idx - (historyData.messageList.length as number));
        const messages: MessageItem[] = historyData.messageList;
        setMessageList([
          ...messages,
          ...(latestOffsetParams.current.startClientMsgID ? messageList : []),
        ]);
        setOffSetParams({
          startClientMsgID: messages[0]?.clientMsgID ?? '',
          lastMinSeq: historyData.lastMinSeq,
          hasMore: !historyData.isEnd && messages.length === 20,
        });
        if (!item) {
          scrollToMsg(messageList[0]?.clientMsgID).then(() => {
            setTimeout(() => {
              setGetMessageLoading(false);
            });
          });
        }
        if (item && messages.length > 0) {
          scrollToBottom(messages[messages.length - 1].clientMsgID).then(() => {
            setTimeout(() => {
              setGetMessageLoading(false);
            });
          });
        }
      });
  };

  const selectConversation = (item: ConversationItem) => {
    setCurrentConversation(item);
    setOffSetParams({
      startClientMsgID: '',
      lastMinSeq: 0,
      hasMore: true,
    });
    setTimeout(() => {
      getMessageNomaly(item);
    });
  };

  const searchMessageByTime = () => {
    const options = {
      conversationID: currentConversation!.conversationID,
      keywordList: [],
      keywordListMatchType: 0,
      senderUserIDList: [],
      messageTypeList: canSearchMessageTypes,
      searchTimePosition: latestSearchParams.current.timeRange[1],
      searchTimePeriod:
        latestSearchParams.current.timeRange[1] - latestSearchParams.current.timeRange[0],
      pageIndex: latestSearchOffset.current.pageIndex,
      count: 20,
    };

    openIM.searchLocalMessages(options).then(({ data }) => {
      const tmpData = data;
      const messages = tmpData.searchResultItems ? tmpData.searchResultItems[0].messageList : [];
      setFirstItemIndex((idx) => idx - (messages.length as number));

      setMessageList([
        ...messages.reverse(),
        ...(latestSearchOffset.current.pageIndex === 1 ? [] : messageList),
      ]);
      if (latestSearchOffset.current.pageIndex !== 1) {
        scrollToMsg(messageList[0]?.clientMsgID).then(() => {
          setTimeout(() => {
            setSearchOffset({
              loading: false,
              pageIndex: latestSearchOffset.current.pageIndex + 1,
              hasMore: messages.length === 20,
            });
          });
        });
      }

      if (latestSearchOffset.current.pageIndex === 1 && messages.length > 0) {
        scrollToBottom(messages[messages.length - 1].clientMsgID).then(() => {
          setTimeout(() => {
            setSearchOffset({
              loading: false,
              pageIndex: latestSearchOffset.current.pageIndex + 1,
              hasMore: messages.length === 20,
            });
          });
        });
      }
    });
  };

  const datePackerChanged = (dates?: Moment[]) => {
    console.log(dates);

    if (!dates) {
      setSearchParams({
        keywords: '',
        timeButton: '',
        compRange: [],
        timeRange: [],
      });

      selectConversation({ ...currentConversation! });
      return;
    }
    setSearchParams({
      ...searchParams,
      keywords: '',
      compRange: dates as any,
      timeRange: dates.map((item) => item.unix()),
    });
    setSearchOffset({
      pageIndex: 1,
      loading: true,
      hasMore: true,
    });
    setTimeout(() => {
      searchMessageByTime();
    });
  };

  const searchMessageByKeyWords = () => {
    const options = {
      conversationID: currentConversation!.conversationID,
      keywordList: [latestSearchParams.current.keywords],
      keywordListMatchType: 0,
      senderUserIDList: [],
      messageTypeList: [],
      searchTimePosition: 0,
      searchTimePeriod: 0,
      pageIndex: latestSearchOffset.current.pageIndex,
      count: 20,
    };

    openIM.searchLocalMessages(options).then(({ data }) => {
      const tmpData = data;
      console.log(tmpData);

      const messages = tmpData.searchResultItems ? tmpData.searchResultItems[0].messageList : [];
      setFirstItemIndex((idx) => idx - (messages.length as number));

      setMessageList([
        ...messages.reverse(),
        ...(latestSearchOffset.current.pageIndex === 1 ? [] : messageList),
      ]);
      if (latestSearchOffset.current.pageIndex !== 1) {
        scrollToMsg(messageList[0]?.clientMsgID).then(() => {
          setTimeout(() => {
            setSearchOffset({
              loading: false,
              pageIndex: latestSearchOffset.current.pageIndex + 1,
              hasMore: messages.length === 20,
            });
          });
        });
      }

      if (latestSearchOffset.current.pageIndex === 1 && messages.length > 0) {
        scrollToBottom(messages[messages.length - 1].clientMsgID).then(() => {
          setTimeout(() => {
            setSearchOffset({
              loading: false,
              pageIndex: latestSearchOffset.current.pageIndex + 1,
              hasMore: messages.length === 20,
            });
          });
        });
      }
    });
  };

  const onRightScroll = () => {
    const isSearch =
      latestSearchParams.current.keywords || latestSearchParams.current.timeRange.length > 0;
    const loadMoreFlag = true;

    if (loadMoreFlag) {
      if (isSearch) {
        if (latestSearchParams.current.keywords) {
          searchMessageByKeyWords();
        } else {
          searchMessageByTime();
        }
      } else {
        getMessageNomaly();
      }
    }
  };

  const throttleScroll = throttle(onRightScroll, 250);

  const enterSearchByKeywords = () => {
    setSearchOffset({
      pageIndex: 1,
      loading: true,
      hasMore: true,
    });
    setTimeout(() => {
      searchMessageByKeyWords();
    });
  };

  const setTimeButton = (type: string) => {
    let tmpArr = [moment().subtract(1, type as any), moment()];

    setSearchParams({
      keywords: '',
      timeButton: type,
      compRange: tmpArr as any,
      timeRange: tmpArr.map((item) => item.unix()),
    });
    setSearchOffset({
      pageIndex: 1,
      loading: true,
      hasMore: true,
    });
    setTimeout(() => {
      searchMessageByTime();
    });
  };

  const searchConversation = () => {
    if (conversationSearchStr) {
      setIsSearching(true);
    }
  };

  const getRenderConversationList = isSearching
    ? conversationList.filter(
        (item) =>
          item.showName.includes(conversationSearchStr) ||
          item.userID.includes(conversationSearchStr) ||
          item.groupID.includes(conversationSearchStr),
      )
    : conversationList;

  return (
    <Modal
      destroyOnClose
      footer={null}
      closable={false}
      className={styles.preview_modal}
      width={1068}
      open
      onCancel={onCancel}
    >
      <div className="h-[35px] leading-[35px] w-full bg-[#F0F8FF] text-right pr-3">
        <CloseOutlined onClick={onCancel} className="text-xl text-[#999] cursor-pointer" />
      </div>
      <Spin spinning={syncLoading}>
        <div className="flex h-[600px] bg-[#f0f2f5]">
          <div className="w-[320px] bg-white mr-3">
            <div className="flex mx-3 my-3">
              <Space size={'middle'}>
                <Input
                  value={conversationSearchStr}
                  allowClear
                  onChange={(e) => {
                    setConversationSearchStr(e.target.value);
                    if (!e.target.value) {
                      setIsSearching(false);
                    }
                  }}
                  prefix={<SearchOutlined />}
                  placeholder="请输入联系人ID或群组ID"
                />
                <Button onClick={searchConversation} size="small" type="primary">
                  查询
                </Button>
              </Space>
            </div>
            <Virtuoso
              className="h-[542px] overflow-x-hidden"
              data={getRenderConversationList}
              startReached={throttleScroll}
              itemContent={(idx, item) => {
                return (
                  <div
                    key={idx}
                    onClick={() => selectConversation(item)}
                    className={classNames(
                      'flex items-center px-3 py-2 hover:bg-[#f4f8fd] rounded',
                      {
                        'bg-[#f4f8fd]': currentConversation?.conversationID === item.conversationID,
                      },
                    )}
                  >
                    <OIMAvatar
                      isgroup={item.conversationType === SessionType.WorkingGroup}
                      src={item.faceURL}
                      text={item.showName}
                    />
                    <div className="ml-3 truncate max-w-[120px]">{item.showName}</div>
                  </div>
                );
              }}
            />
          </div>
          <div className="flex flex-col w-full">
            <div className="bg-white mb-3 py-4 px-3 flex justify-between">
              <div>
                <DatePicker.RangePicker
                  disabled={currentConversation === undefined}
                  onChange={datePackerChanged as any}
                  value={searchParams.compRange as any}
                />
                <Button
                  disabled={currentConversation === undefined}
                  onClick={() => setTimeButton('day')}
                  type={searchParams.timeButton === 'day' ? 'link' : 'text'}
                >
                  本日
                </Button>
                <Button
                  disabled={currentConversation === undefined}
                  onClick={() => setTimeButton('week')}
                  type={searchParams.timeButton === 'week' ? 'link' : 'text'}
                >
                  本周
                </Button>
                <Button
                  disabled={currentConversation === undefined}
                  onClick={() => setTimeButton('month')}
                  type={searchParams.timeButton === 'month' ? 'link' : 'text'}
                >
                  本月
                </Button>
              </div>
              <Space size={'middle'}>
                <Input
                  allowClear
                  disabled={currentConversation === undefined}
                  onChange={(e) => {
                    setSearchParams({
                      ...searchParams,
                      keywords: e.target.value,
                    });
                    if (!e.target.value) {
                      selectConversation({ ...currentConversation! });
                    }
                  }}
                  value={searchParams.keywords}
                  prefix={<SearchOutlined />}
                  placeholder="请输入关键词"
                />
                <Button
                  disabled={currentConversation === undefined}
                  onClick={enterSearchByKeywords}
                  size="small"
                  type="primary"
                >
                  查询
                </Button>
              </Space>
            </div>
            <Spin spinning={getMessageLoading && messageList.length > 0}>
              <div className="bg-white h-[522px]">
                <Virtuoso
                  ref={virtuoso}
                  firstItemIndex={firstItemIndex}
                  initialTopMostItemIndex={INITIAL_ITEM_COUNT - 1}
                  className="h-full overflow-x-hidden"
                  data={messageList}
                  totalListHeightChanged={(height) => (virtuosoHeight.current = height)}
                  startReached={throttleScroll}
                  itemContent={(idx, message) => {
                    return (
                      <HistoryMessageItem
                        key={idx}
                        message={message}
                        isSelf={message.sendID === userID}
                      />
                    );
                  }}
                />
              </div>
            </Spin>
          </div>
        </div>
      </Spin>
    </Modal>
  );
};

export default PreviewMessageModal;

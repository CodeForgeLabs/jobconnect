"use client";
import {
  Check,
  CheckCheck,
  Send,
  Paperclip,
  Image as ImageIcon,
  MessageCircle,
  Smile,
  Edit
} from "lucide-react";
import { useState, useEffect } from "react";
import Image from "next/image";
import {
  useGetMessagesQuery,
  useGetConversationsQuery,
  useMarkAsSeenMutation,
  useSendMessageMutation,
  ConversationItem as ConversationItemType,
  LastMessage,
} from "@/api/messageapi";
import { useGetMeQuery } from "@/api/userapi";
import { useChatSocket } from "@/hooks/useChatSocket";

function isNewMessagePayload(
  value: unknown,
): value is { type: "new_message"; data: LastMessage } {
  if (typeof value !== "object" || value === null) return false;

  const candidate = value as { type?: unknown; data?: unknown };

  return candidate.type === "new_message" && typeof candidate.data === "object";
}

export default function ChatPage() {
  const { data: userData } = useGetMeQuery();
  const userId = userData?.id || 0;

  const [activeChat, setActiveChat] = useState<number>(-1);
  const [messageText, setMessageText] = useState("");
  const [liveMessages, setLiveMessages] = useState<LastMessage[]>([]);
  const { data: temp, refetch: refetchConversations } =
    useGetConversationsQuery();
  const conversations = temp?.conversations || [];
  const [sendMessage, { isLoading: isSending }] = useSendMessageMutation();
  const [markAsSeen] = useMarkAsSeenMutation();

  // Fetch messages only when activeChat is selected (activeChat stores conversationID)
  const shouldFetchMessages = activeChat !== -1;
  const { data: messagesData, refetch: refetchMessages } = useGetMessagesQuery(
    { conversationId: activeChat },
    { skip: !shouldFetchMessages },
  );

  const renderedMessages = [...(messagesData?.messages ?? []), ...liveMessages];

  useEffect(() => {
    if (!shouldFetchMessages) return;

    const seenTimer = window.setTimeout(() => {
      markAsSeen(activeChat)
        .unwrap()
        .then(() => {
          refetchConversations();
        })
        .catch((error) => {
          console.error("Failed to mark conversation as seen", error);
        });
    }, 250);

    return () => window.clearTimeout(seenTimer);
  }, [activeChat, shouldFetchMessages, markAsSeen, refetchConversations]);

  // Setup WebSocket for real-time messages
  useChatSocket(userId, (msg) => {
    if (isNewMessagePayload(msg)) {
      const newMsg = msg.data;

      // Only add if it's for the currently active chat
      if (newMsg.ConversationID === activeChat) {
        setLiveMessages((prev) =>
          prev.some((message) => message.Id === newMsg.Id)
            ? prev
            : [...prev, newMsg],
        );

        // Refetch to ensure sync with server
        refetchMessages();

        markAsSeen(activeChat)
          .unwrap()
          .catch((error) => {
            console.error("Failed to mark incoming message as seen", error);
          });
      }

      // Always refetch conversations to update the list with the latest message
      refetchConversations();
    }
  });

  const handleSendMessage = async () => {
    const text = messageText.trim();

    if (!text || activeChat === -1 || !userData) return;

    const activeConversation = conversations.find(
      (conversation: ConversationItemType) =>
        (conversation.LastMessage?.ConversationID ||
          conversation.OtherUserID) === activeChat,
    );

    if (!activeConversation?.User) return;

    setMessageText("");
    console.log("send to user", activeConversation.User.id, "message:", text);

    const payload = {
      caption: "",
      image_url: "",
      receiver_id: activeConversation.User.id,
      sender_id: userId,
      text: text,
      type: "text",
      video_url: "",
    };

    try {
      await sendMessage(payload).unwrap();

      await Promise.all([refetchMessages(), refetchConversations()]);
    } catch (error) {
      console.error("Failed to send message", error);
      setMessageText(text);
    }
  };

  return (
    <div className="flex  flex-col h-[90vh] bg-surface text-on-surface selection:bg-primary-fixed selection:text-primary">
      <div className="flex flex-1 overflow-hidden">
        {/* Sidebar: Conversation List */}
        <aside className="w-80 md:w-96 shrink-0 border-r border-outline-variant/30 bg-surface-container-low flex flex-col">
          <div className="p-6 flex items-center justify-between">
            <h1 className="text-2xl font-extrabold font-headline text-primary tracking-tight">
              Messages
            </h1>
            <button className="bg-primary text-white p-2.5 rounded-xl hover:opacity-90 transition-all flex items-center">
              <span className="material-symbols-outlined text-xl">
                <Edit size={18} />
              </span>
            </button>
          </div>

          <div className="flex-1 overflow-y-auto px-3 space-y-1 pb-4 custom-scrollbar">
            {conversations && conversations.length > 0 ? (
              conversations.map((conv: ConversationItemType) => {
                const convId =
                  conv.LastMessage?.ConversationID ?? conv.OtherUserID;

                if (!conv.User) return null;

                return (
                  <ConversationItem
                    key={convId}
                    active={activeChat === convId}
                    onClick={() => {
                      setLiveMessages([]);
                      setActiveChat(convId);
                    }}
                    name={`${conv.User.first_name || ""} ${conv.User.last_name || ""}`.trim()}
                    msg={
                      conv.LastMessage?.Text ||
                      conv.LastMessage?.Caption ||
                      "No messages yet"
                    }
                    time={
                      conv.LastMessage?.CreatedAt
                        ? new Date(
                            conv.LastMessage.CreatedAt,
                          ).toLocaleTimeString([], {
                            hour: "2-digit",
                            minute: "2-digit",
                          })
                        : ""
                    }
                    avatar={
                      conv.User.profile_picture_url ||
                      "https://img.daisyui.com/images/stock/photo-1534528741775-53994a69daeb.webp"
                    }
                    unread={conv.UnseenCount}
                    lastsender={conv.LastMessage?.SenderID == userId}
                    isSeen={conv.LastMessage?.IsSeen}
                  />
                );
              })
            ) : (
              <div className="p-4 text-center text-outline text-sm">
                No conversations yet
              </div>
            )}
          </div>
        </aside>

        {/* Chat Window */}
        <section className="flex-1 flex flex-col bg-surface-container-lowest">
          {activeChat === -1 ? (
            // Empty state
            <div className="flex-1 flex flex-col items-center justify-center gap-6 text-center">
              <div className="w-24 h-24 rounded-full bg-primary/10 flex items-center justify-center">
                <span className="material-symbols-outlined text-5xl text-primary">
                  <MessageCircle size={48} />
                </span>
              </div>
              <div>
                <h2 className="text-2xl font-bold text-on-surface mb-2">
                  Select a conversation
                </h2>
                <p className="text-outline max-w-xs">
                  Choose a conversation from the list to start messaging
                </p>
              </div>
            </div>
          ) : (
            <>
              {/* Chat Header */}
              {conversations &&
                (() => {
                  const activeConversation = conversations.find(
                    (c: ConversationItemType) =>
                      (c.LastMessage?.ConversationID || c.OtherUserID) ===
                      activeChat,
                  );
                  if (!activeConversation || !activeConversation.User)
                    return null;

                  const userName =
                    `${activeConversation.User.first_name || ""} ${activeConversation.User.last_name || ""}`.trim();

                  return (
                    <header className="h-20 shrink-0 flex items-center justify-between px-8 border-b border-outline-variant/20 bg-white/50 backdrop-blur-md">
                      <div className="flex items-center gap-4">
                        <div className="relative">
                          <Image
                            src={
                              activeConversation.User.profile_picture_url ||
                              "https://img.daisyui.com/images/stock/photo-1534528741775-53994a69daeb.webp"
                            }
                            className="h-10 w-10 rounded-xl object-cover"
                            alt={userName}
                            width={40}
                            height={40}
                          />
                          {/* Commented: add online status if available */}
                          {/* <div className="absolute -bottom-1 -right-1 h-3.5 w-3.5 bg-green-500 border-2 border-white rounded-full"></div> */}
                        </div>
                        <div>
                          <h2 className="font-bold text-primary font-headline leading-none">
                            {userName}
                          </h2>
                          {activeConversation.User.headline && (
                            <span className="text-[10px] font-bold text-outline uppercase tracking-widest">
                              {activeConversation.User.headline}
                            </span>
                          )}
                        </div>
                      </div>
                      <div className="flex items-center gap-2">
                        <button className="hidden sm:flex items-center gap-2 px-4 py-2 bg-primary-fixed text-on-primary-fixed-variant rounded-xl font-bold text-xs hover:bg-primary-fixed-dim transition-all">
                          <span className="material-symbols-outlined text-sm">
                            description
                          </span>
                          View Contract
                        </button>
                        <div className="w-px h-6 bg-outline-variant mx-2 hidden sm:block"></div>
                        <button className="p-2 text-outline hover:bg-surface-container rounded-lg transition-all">
                          <span className="material-symbols-outlined">
                            call
                          </span>
                        </button>
                        <button className="p-2 text-outline hover:bg-surface-container rounded-lg transition-all">
                          <span className="material-symbols-outlined">
                            videocam
                          </span>
                        </button>
                      </div>
                    </header>
                  );
                })()}

              {/* Messages Area */}
              <div className="flex-1 overflow-y-auto p-6 md:p-10 space-y-8 custom-scrollbar overflow-scroll">
                {renderedMessages.length > 0 ? (
                  renderedMessages.map((msg: LastMessage, idx: number) =>
                    msg.SenderID === userId ? (
                      <SentMsg
                        key={idx}
                        texts={[msg.Text || ""]}
                        time={new Date(msg.CreatedAt).toLocaleTimeString([], {
                          hour: "2-digit",
                          minute: "2-digit",
                        })}
                        isSeen={msg.IsSeen}
                      />
                    ) : (
                      <ReceivedMsg
                        key={idx}
                        text={msg.Text || ""}
                        time={new Date(msg.CreatedAt).toLocaleTimeString([], {
                          hour: "2-digit",
                          minute: "2-digit",
                        })}
                      />
                    ),
                  )
                ) : (
                  <div className="text-center text-outline text-sm pt-8">
                    No messages yet. Start the conversation!
                  </div>
                )}
              </div>

              {/* Input Area */}
              <footer className="p-6 bg-white border-t border-outline-variant/20">
                <div className="max-w-4xl mx-auto">
                  <div className="bg-surface-container-low rounded-2xl p-3 border border-outline-variant/20 shadow-sm focus-within:shadow-md focus-within:border-primary/30 transition-all">
                    <textarea
                      value={messageText}
                      onChange={(event) => setMessageText(event.target.value)}
                      onKeyDown={(event) => {
                        if (event.key === "Enter" && !event.shiftKey) {
                          event.preventDefault();
                          handleSendMessage();
                        }
                      }}
                      className="w-full bg-transparent border-none focus:ring-0 text-sm text-on-surface placeholder:text-outline/60 resize-none font-medium h-12"
                      placeholder="Type your message..."
                      disabled={activeChat === -1 || isSending}
                    />
                    <div className="flex items-center justify-between mt-2 pt-2 border-t border-outline-variant/10">
                      <div className="flex items-center gap-1">
                        <button className="p-2 text-outline hover:text-primary hover:bg-primary/5 rounded-lg transition-all">
                          <Paperclip size={18} />
                        </button>
                        <button className="p-2 text-outline hover:text-primary hover:bg-primary/5 rounded-lg transition-all">
                          <Smile size={18} />
                        </button>
                        <button className="p-2 text-outline hover:text-primary hover:bg-primary/5 rounded-lg transition-all">
                          <ImageIcon size={18} />
                        </button>
                        <div className="w-px h-4 bg-outline-variant mx-1" />
                        <button className="p-2 text-outline hover:text-primary hover:bg-primary/5 rounded-lg transition-all">
                          Bold
                        </button>
                      </div>
                      <button
                        onClick={handleSendMessage}
                        disabled={
                          activeChat === -1 || !messageText.trim() || isSending
                        }
                        className="bg-primary text-white px-6 py-2 rounded-xl font-bold text-sm flex items-center gap-2 hover:shadow-lg active:scale-95 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        Send <Send size={16} />
                      </button>
                    </div>
                  </div>
                </div>
              </footer>
            </>
          )}
        </section>
      </div>
    </div>
  );
}

interface ConversationItemProps {
  active?: boolean;
  unread?: number;
  name: string;
  msg: string;
  time: string;
  online?: boolean;
  avatar?: string;
  initials?: string;
  lastsender?: boolean;
  isSeen?: boolean;
  onClick?: () => void;
}

const ConversationItem = ({
  active,
  unread,
  name,
  msg,
  time,
  online,
  avatar,
  initials,
  onClick,
  lastsender,
  isSeen,
}: ConversationItemProps) => (
  <div
    onClick={onClick}
    className={`p-4 rounded-2xl flex gap-4 cursor-pointer transition-all ${active ? "bg-primary/10" : "hover:bg-white"}`}
  >
    <div className="relative shrink-0">
      {avatar ? (
        <div className="h-12 w-12 rounded-2xl bg-primary/10 flex items-center justify-center text-primary font-bold font-headline text-sm">
          {initials || "U"}
        </div>
      ) : (
        <div className="h-12 w-12 rounded-2xl bg-surface-container-highest flex items-center justify-center text-primary font-bold font-headline">
          {initials || "U"}
        </div>
      )}
      {online && (
        <div className="absolute -bottom-0.5 -right-0.5 h-4 w-4 bg-green-500 border-4 border-surface-container-low rounded-full"></div>
      )}
    </div>
    <div className="flex-1 min-w-0">
      <div className="flex justify-between items-baseline mb-0.5">
        <span
          className={`font-bold font-headline truncate ${active ? "text-primary" : "text-on-surface"}`}
        >
          {name}
        </span>
        <span className="text-[10px] font-bold text-outline uppercase tracking-wider">
          {time}
        </span>
      </div>
      <div className="flex justify-between items-center">
        <p
          className={`text-sm truncate ${active ? "text-primary/80" : "text-outline"}`}
        >
          {msg}
        </p>
        {(unread ?? 0) > 0 ? (
          <span className="h-5 w-5 bg-tertiary-container text-on-tertiary-container text-[10px] font-bold flex items-center justify-center rounded-full">
            {unread}
          </span>
        ) : lastsender ? (
          isSeen ? (
            <span className="material-symbols-outlined text-sm text-primary">
              <CheckCheck size={16} />
            </span>
          ) : (
            <span className="material-symbols-outlined text-sm text-primary">
              <Check size={16} />
            </span>
          )
        ) : null}
      </div>
    </div>
  </div>
);

interface ReceivedMsgProps {
  text: string;
  time: string;
}

const ReceivedMsg = ({ text, time }: ReceivedMsgProps) => (
  <div className="flex gap-4 max-w-[80%] lg:max-w-[60%] animate-in slide-in-from-left-2 duration-300">
    <div className="h-8 w-8 rounded-lg bg-surface-container-highest mt-auto" />
    <div className="space-y-2">
      <div className="bg-white p-4 rounded-2xl rounded-bl-none shadow-sm text-sm border border-outline-variant/10 leading-relaxed">
        {text}
      </div>
      <div className="text-[10px] font-bold text-outline pl-1 uppercase">
        {time}
      </div>
    </div>
  </div>
);

interface SentMsgProps {
  texts: string[];
  time: string;
  isSeen?: boolean;
}

const SentMsg = ({ texts, time, isSeen }: SentMsgProps) => (
  <div className="flex flex-row-reverse gap-4 max-w-[80%] lg:max-w-[60%] ml-auto animate-in slide-in-from-right-2 duration-300">
    <div className="space-y-2 flex flex-col items-end">
      {texts.map((t: string, i: number) => (
        <div
          key={i}
          className="bg-primary text-white p-4 rounded-2xl rounded-br-none shadow-md text-sm leading-relaxed"
        >
          {t}
        </div>
      ))}
      <div className="flex items-center gap-1.5 text-[10px] font-bold text-outline pr-1 uppercase">
        {time}{" "}
        <span className="material-symbols-outlined text-sm text-primary">
          {isSeen ? <CheckCheck size={16} /> : <Check size={16} />}
        </span>
      </div>
    </div>
  </div>
);

interface IconButtonProps {
  icon: string;
}

const IconButton = ({ icon }: IconButtonProps) => (
  <button className="p-2 text-outline hover:text-primary hover:bg-primary/5 rounded-lg transition-all">
    <span className="material-symbols-outlined text-xl">{icon}</span>
  </button>
);

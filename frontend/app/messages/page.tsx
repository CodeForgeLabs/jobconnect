"use client";
import {
  Check,
  CheckCheck,
  Send,
  Paperclip,
  Image as ImageIcon,
  MessageCircle,
  Smile,
  Edit,
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
import {
  useGetMeQuery,
  useSearchUsersByNameQuery,
  User,
  useGetUserByIdQuery,
} from "@/api/userapi";
import { useChatSocket } from "@/hooks/useChatSocket";
import avatarPlaceholder from "@/assets/avatarsvg.png";

function isNewMessagePayload(
  value: unknown,
): value is { type: "new_message"; data: LastMessage } {
  if (typeof value !== "object" || value === null) return false;

  const candidate = value as { type?: unknown; data?: unknown };

  return candidate.type === "new_message" && typeof candidate.data === "object";
}

const DEFAULT_AVATAR_URL = avatarPlaceholder.src;

export default function ChatPage() {
  const { data: userData } = useGetMeQuery();
  const userId = userData?.id || 0;

  const [activeChat, setActiveChat] = useState<number>(-1);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [messageText, setMessageText] = useState("");
  const [searchName, setSearchName] = useState("");
  const [liveMessages, setLiveMessages] = useState<LastMessage[]>([]);
  const [optimisticMessages, setOptimisticMessages] = useState<LastMessage[]>(
    [],
  );
  const { data: temp, refetch: refetchConversations } =
    useGetConversationsQuery();
  const conversations = temp?.conversations || [];
  const [sendMessage, { isLoading: isSending }] = useSendMessageMutation();
  const [markAsSeen] = useMarkAsSeenMutation();

  const searchTerm = searchName.trim();
  const shouldSearchUsers = searchTerm.length > 1;
  const { data: searchedUsers, isFetching: isSearchingUsers } =
    useSearchUsersByNameQuery(searchTerm, {
      skip: !shouldSearchUsers,
    });

  const searchResults = (searchedUsers || []).filter(
    (user) => user.id !== userId,
  );

  const currentConversation =
    activeChat === -1
      ? null
      : conversations.find(
          (conversation: ConversationItemType) =>
            (conversation.LastMessage?.ConversationID ||
              conversation.OtherUserID) === activeChat,
        ) || null;

  const activeUser = currentConversation?.User || selectedUser;
  const activeUserId = activeUser?.id ?? 0;
  const { data: activeUserDetails } = useGetUserByIdQuery(activeUserId, {
    skip: !activeUserId,
  });
  const activeUserProfilePicture =
    activeUserDetails?.profile_picture_url ||
    activeUser?.profile_picture_url ||
    DEFAULT_AVATAR_URL;
  const activeUserDisplayName =
    `${activeUser?.first_name || ""} ${activeUser?.last_name || ""}`.trim() ||
    activeUser?.email ||
    "User";

  const buildOptimisticMessage = (
    text: string,
    receiverId: number,
  ): LastMessage => ({
    Caption: "",
    ConversationID: -1,
    CreatedAt: new Date().toISOString(),
    DeletedAt: null,
    EditedAt: null,
    Id: Date.now(),
    ImageUrl: null,
    IsDeleted: false,
    IsEdited: false,
    IsSeen: false,
    ReceiverID: receiverId,
    seenAt: null,
    SenderID: userId,
    Text: text,
    Type: "text",
    VideoUrl: null,
  });

  // Fetch messages only when activeChat is selected (activeChat stores conversationID)
  const shouldFetchMessages = activeChat !== -1;
  const { data: messagesData, refetch: refetchMessages } = useGetMessagesQuery(
    { conversationId: activeChat },
    { skip: !shouldFetchMessages },
  );

  const renderedMessages = shouldFetchMessages
    ? [...(messagesData?.messages ?? []), ...liveMessages]
    : selectedUser
      ? optimisticMessages
      : [];

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

    if (!text || !userData || !activeUser) return;

    const activeConversation = conversations.find(
      (conversation: ConversationItemType) =>
        conversation.User?.id === activeUser.id ||
        conversation.OtherUserID === activeUser.id,
    );

    setMessageText("");

    const receiverId = activeConversation?.User?.id ?? activeUser.id;

    const payload = {
      caption: "",
      image_url: "",
      receiver_id: receiverId,
      sender_id: userId,
      text: text,
      type: "text",
      video_url: "",
    };

    try {
      await sendMessage(payload).unwrap();

      if (activeChat === -1) {
        setOptimisticMessages((prev) => [
          ...prev,
          buildOptimisticMessage(text, receiverId),
        ]);
      }

      let nextConversation = conversations.find(
        (conversation: ConversationItemType) =>
          conversation.User?.id === receiverId ||
          conversation.OtherUserID === receiverId,
      );

      for (let attempt = 0; !nextConversation && attempt < 4; attempt += 1) {
        await new Promise((resolve) => window.setTimeout(resolve, 200));
        const refreshedConversations = await refetchConversations();

        nextConversation = refreshedConversations.data?.conversations.find(
          (conversation: ConversationItemType) =>
            conversation.User?.id === receiverId ||
            conversation.OtherUserID === receiverId,
        );
      }

      if (nextConversation) {
        const nextConversationId =
          nextConversation.LastMessage?.ConversationID ??
          nextConversation.OtherUserID;

        setActiveChat(nextConversationId);
        setSelectedUser(null);
        setLiveMessages([]);
        setOptimisticMessages([]);
        await refetchMessages();
      }
    } catch (error) {
      console.error("Failed to send message", error);
      setMessageText(text);
    }
  };

  const openSearchModal = () => {
    setSearchName("");
    (
      document.getElementById("search_user_modal") as HTMLDialogElement | null
    )?.showModal();
  };

  const closeSearchModal = () => {
    (
      document.getElementById("search_user_modal") as HTMLDialogElement | null
    )?.close();
  };

  const startChatWithUser = (user: User) => {
    const existingConversation = conversations.find(
      (conversation: ConversationItemType) =>
        conversation.User?.id === user.id ||
        conversation.OtherUserID === user.id,
    );

    setMessageText("");
    setLiveMessages([]);
    setOptimisticMessages([]);

    if (existingConversation) {
      const conversationId =
        existingConversation.LastMessage?.ConversationID ??
        existingConversation.OtherUserID;
      setActiveChat(conversationId);
      setSelectedUser(null);
    } else {
      setActiveChat(-1);
      setSelectedUser(user);
    }

    closeSearchModal();
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
            <button
              className="bg-primary text-white p-2.5 rounded-xl hover:opacity-90 transition-all flex items-center"
              onClick={openSearchModal}
              type="button"
            >
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
                    userId={conv.User.id}
                    onClick={() => {
                      setLiveMessages([]);
                      setOptimisticMessages([]);
                      setActiveChat(convId);
                      setSelectedUser(null);
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
                    unread={conv.UnseenCount}
                    lastsender={conv.LastMessage?.SenderID == userId}
                    isSeen={conv.LastMessage?.IsSeen}
                    fallbackAvatar={conv.User.profile_picture_url || undefined}
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
          {activeUser ? (
            <>
              {/* Chat Header */}
              <header className="h-20 shrink-0 flex items-center justify-between px-8 border-b border-outline-variant/20 bg-white/50 ">
                <div className="flex items-center gap-4">
                  <div className="relative">
                    <Image
                      src={activeUserProfilePicture}
                      className="h-10 w-10 rounded-xl object-cover"
                      alt={activeUserDisplayName}
                      width={40}
                      height={40}
                    />
                  </div>
                  <div>
                    <h2 className="font-bold text-primary font-headline leading-none">
                      {activeUserDisplayName}
                    </h2>
                    {activeUserDetails?.headline && (
                      <span className="text-[10px] font-bold text-outline uppercase tracking-widest">
                        {activeUserDetails.headline}
                      </span>
                    )}
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <div className="w-px h-6 bg-outline-variant mx-2 hidden sm:block"></div>
                  <button className="p-2 text-outline hover:bg-surface-container rounded-lg transition-all">
                    <span className="material-symbols-outlined">call</span>
                  </button>
                </div>
              </header>

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
                        avatarUrl={userData?.profile_picture_url || undefined}
                      />
                    ) : (
                      <ReceivedMsg
                        key={idx}
                        text={msg.Text || ""}
                        time={new Date(msg.CreatedAt).toLocaleTimeString([], {
                          hour: "2-digit",
                          minute: "2-digit",
                        })}
                        avatarUrl={activeUserProfilePicture}
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
                      disabled={!activeUser || isSending}
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
                          !activeUser || !messageText.trim() || isSending
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
          ) : (
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
          )}
        </section>
      </div>

      <dialog
        id="search_user_modal"
        className="rounded-3xl p-0 w-[min(92vw,42rem)] bg-surface text-on-surface shadow-2xl backdrop:bg-black/50"
      >
        <div className="p-6 md:p-8">
          <div className="flex items-start justify-between gap-4 mb-5">
            <div>
              <h3 className="text-xl font-extrabold font-headline text-primary">
                Search users
              </h3>
              <p className="text-sm text-outline mt-1">
                Find someone by name and start a conversation.
              </p>
            </div>
            <form method="dialog">
              <button className="p-2 rounded-lg hover:bg-surface-container transition-all text-outline">
                <span className="material-symbols-outlined">close</span>
              </button>
            </form>
          </div>

          <label className="block">
            <span className="sr-only">Search by name</span>
            <input
              value={searchName}
              onChange={(event) => setSearchName(event.target.value)}
              placeholder="Type a name..."
              className="w-full rounded-2xl border border-outline-variant/30 bg-surface-container-low px-4 py-3 text-sm outline-none focus:border-primary/40 focus:ring-2 focus:ring-primary/10"
            />
          </label>

          <div className="mt-5 max-h-[60vh] overflow-y-auto space-y-2 pr-1 custom-scrollbar">
            {!shouldSearchUsers ? (
              <div className="rounded-2xl border border-dashed border-outline-variant/40 px-4 py-6 text-sm text-outline text-center">
                Enter at least 2 characters to search.
              </div>
            ) : isSearchingUsers ? (
              <div className="rounded-2xl border border-outline-variant/20 px-4 py-6 text-sm text-outline text-center">
                Searching users...
              </div>
            ) : searchResults.length > 0 ? (
              searchResults.map((user) => {
                const fullName =
                  `${user.first_name || ""} ${user.last_name || ""}`.trim() ||
                  user.email;

                return (
                  <button
                    key={user.id}
                    type="button"
                    onClick={() => startChatWithUser(user)}
                    className="w-full flex items-center gap-4 rounded-2xl border border-outline-variant/20 bg-white px-4 py-3 text-left transition-all hover:border-primary/30 hover:bg-primary/5"
                  >
                    <div className="h-12 w-12 overflow-hidden rounded-2xl bg-primary/10 flex items-center justify-center shrink-0">
                      {user.profile_picture_url ? (
                        <Image
                          src={user.profile_picture_url}
                          alt={fullName}
                          width={48}
                          height={48}
                          className="h-full w-full object-cover"
                        />
                      ) : (
                        <span className="font-bold text-primary">
                          {fullName.charAt(0).toUpperCase() || "U"}
                        </span>
                      )}
                    </div>
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center justify-between gap-4">
                        <div className="min-w-0">
                          <p className="font-bold text-on-surface truncate">
                            {fullName}
                          </p>
                          <p className="text-xs text-outline truncate">
                            {user.headline || user.email}
                          </p>
                        </div>
                        <span className="shrink-0 rounded-full bg-primary text-white px-3 py-1 text-xs font-semibold">
                          Chat
                        </span>
                      </div>
                    </div>
                  </button>
                );
              })
            ) : (
              <div className="rounded-2xl border border-dashed border-outline-variant/40 px-4 py-6 text-sm text-outline text-center">
                No users found for “{searchTerm}”.
              </div>
            )}
          </div>

          <div className="mt-6 flex justify-end">
            <form method="dialog">
              <button className="rounded-xl border border-outline-variant/30 px-4 py-2 text-sm font-semibold text-on-surface hover:bg-surface-container transition-all">
                Close
              </button>
            </form>
          </div>
        </div>
      </dialog>
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
  lastsender?: boolean;
  isSeen?: boolean;
  onClick?: () => void;
  userId: number;
  fallbackAvatar?: string;
}

const ConversationItem = ({
  active,
  unread,
  name,
  msg,
  time,
  online,
  onClick,
  lastsender,
  isSeen,
  userId,
  fallbackAvatar,
}: ConversationItemProps) => (
  <div
    onClick={onClick}
    className={`p-4 rounded-2xl flex gap-4 cursor-pointer transition-all ${active ? "bg-primary/10" : "hover:bg-white"}`}
  >
    <div className="relative shrink-0">
      <ConversationAvatar userId={userId} fallbackAvatar={fallbackAvatar} />
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
  avatarUrl?: string;
}

const ReceivedMsg = ({ text, time, avatarUrl }: ReceivedMsgProps) => (
  <div className="flex gap-4 max-w-[80%] lg:max-w-[60%] animate-in slide-in-from-left-2 duration-300">
    <div className="h-8 w-8 rounded-lg bg-surface-container-highest mt-auto overflow-hidden shrink-0">
      <Image
        src={avatarUrl || DEFAULT_AVATAR_URL}
        alt="Conversation user"
        width={32}
        height={32}
        className="h-full w-full object-cover"
      />
    </div>
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
  avatarUrl?: string;
}

const SentMsg = ({ texts, time, isSeen, avatarUrl }: SentMsgProps) => (
  <div className="flex flex-row-reverse gap-4 max-w-[80%] lg:max-w-[60%] ml-auto animate-in slide-in-from-right-2 duration-300">
    <div className="h-8 w-8 rounded-lg bg-surface-container-highest mt-auto overflow-hidden shrink-0">
      <Image
        src={avatarUrl || DEFAULT_AVATAR_URL}
        alt="Your avatar"
        width={32}
        height={32}
        className="h-full w-full object-cover"
      />
    </div>
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

interface ConversationAvatarProps {
  userId: number;
  fallbackAvatar?: string;
}

const ConversationAvatar = ({
  userId,
  fallbackAvatar,
}: ConversationAvatarProps) => {
  const { data: user } = useGetUserByIdQuery(userId, { skip: !userId });
  const avatarUrl =
    user?.profile_picture_url || fallbackAvatar || DEFAULT_AVATAR_URL;

  return (
    <Image
      src={avatarUrl}
      alt={
        user
          ? `${user.first_name || ""} ${user.last_name || ""}`.trim() ||
            user.email
          : "Conversation user"
      }
      width={48}
      height={48}
      className="h-12 w-12 rounded-2xl object-cover"
    />
  );
};

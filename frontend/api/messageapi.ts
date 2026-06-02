import { baseApi } from "./baseapi";

export interface LastMessage {
  Caption: string;
  ConversationID: number;
  CreatedAt: string ;
  DeletedAt: string | null;
  EditedAt: string | null;
  Id: number;
  ImageUrl: string | null;
  IsDeleted: boolean;
  IsEdited: boolean;
  IsSeen: boolean;
  ReceiverID: number;
  seenAt: string | null;
  SenderID: number;
  Text: string | null;
  Type: string | null;
  VideoUrl: string | null;
}

export interface ConversationUser {
  availability: string;
  bio: string | null;
  created_at: string | null;
  email: string;
  first_name: string | null;
  headline: string | null;
  hourly_rate: number | null;
  id: number;
  last_name: string | null;
  location: string | null;
  phone_number: string | null;
  profile_picture_url: string | null;
  role: string;
  skills: string | null;
  updated_at: string | null;
}

export interface ConversationItem {
  LastMessage: LastMessage | null;
  OtherUserID: number;
  UnseenCount: number;
  User: ConversationUser;
}

export interface ConversationsResponse {
  conversations: ConversationItem[];
}

export const messageApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getMessages: builder.query <{ messages: LastMessage[] }, { conversationId: number }>({
      query: ({ conversationId }) =>
        `/messages?conversation_id=${conversationId}`,
    }),

    sendMessage: builder.mutation({
      query: (body) => ({
        url: "/messages",
        method: "POST",
        body,
      }),
    }),

    markAsSeen: builder.mutation({
      query: (conversationId) => ({
        url: `/messages/seen?conversation_id=${conversationId}`,
        method: "POST",
      }),
    }),

    getConversations: builder.query<ConversationsResponse, void>({
      query: () => "/messages/conversations",
    }),
  }),
});

export const {
  useGetMessagesQuery,
  useSendMessageMutation,
  useGetConversationsQuery,
  useMarkAsSeenMutation,
} = messageApi;
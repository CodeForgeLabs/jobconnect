"use client";

import { useEffect, useRef } from "react";

export interface ChatMessage {
  ID: number;
  SenderID: number;
  ReceiverID: number;
  ConversationID: number;
  Type: string;
  Text: string;
  ImageUrl: string;
  VideoUrl: string;
  Caption: string;
  IsSeen: boolean;
  SeenAt: string | null;
  IsEdited: boolean;
  EditedAt: string | null;
  IsDeleted: boolean;
  DeletedAt: string | null;
  CreatedAt: string;
}

export type IncomingWSMessage = {
  type: string;
  data: ChatMessage;
};

type MessageHandler = (msg: IncomingWSMessage | unknown) => void;

export function useChatSocket(userId: number, onMessage: MessageHandler) {
  const socketRef = useRef<WebSocket | null>(null);
  const reconnectTimer = useRef<number | null>(null);
  const shouldReconnect = useRef(true);
  const reconnectAttempts = useRef(0);

  useEffect(() => {
    if (!userId) return;

    shouldReconnect.current = true;

    const connect = () => {
      const ws = new WebSocket(`ws://localhost:8080/ws?user_id=${userId}`);
      socketRef.current = ws;

      ws.onopen = () => {
        reconnectAttempts.current = 0;
        console.log("WebSocket connected");
      };

      ws.onmessage = (event) => {
        try {
          const parsed = JSON.parse(event.data);
          if (parsed && typeof parsed === "object" && parsed.type === "new_message") {
            onMessage(parsed as IncomingWSMessage);
          } else {
            onMessage(parsed);
          }
        } catch (err) {
          console.error("WS parse error", err);
          onMessage(event.data);
        }
      };

      ws.onclose = (ev) => {
        socketRef.current = null;
        console.log("WebSocket disconnected", ev.code, ev.reason);
        if (shouldReconnect.current) {
          reconnectAttempts.current += 1;
          const timeout = Math.min(30000, 1000 * 2 ** reconnectAttempts.current);
          // schedule reconnection
          reconnectTimer.current = window.setTimeout(() => connect(), timeout);
          console.log(`Reconnecting in ${timeout}ms`);
        }
      };

      ws.onerror = (err) => {
        console.log("WebSocket error", err);
      };
    };

    connect();

    return () => {
      shouldReconnect.current = false;
      if (reconnectTimer.current) {
        clearTimeout(reconnectTimer.current);
        reconnectTimer.current = null;
      }
      if (socketRef.current) {
        try {
          socketRef.current.close();
        } catch (e) {
          /* ignore */
        }
        socketRef.current = null;
      }
    };
  }, [userId, onMessage]);

  return socketRef;
}
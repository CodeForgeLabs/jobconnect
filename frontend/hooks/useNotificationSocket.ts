"use client";

import { useEffect, useRef } from "react";

import type { Notification } from "@/api/notificationapi";

export type IncomingNotificationWSMessage = {
	type: string;
	data: Notification;
};

type NotificationHandler = (msg: IncomingNotificationWSMessage | unknown) => void;

export function useNotificationSocket(
	userId: number,
	onMessage: NotificationHandler,
) {
	const socketRef = useRef<WebSocket | null>(null);
	const reconnectTimer = useRef<number | null>(null);
	const shouldReconnect = useRef(true);
	const reconnectAttempts = useRef(0);
	const onMessageRef = useRef(onMessage);

	useEffect(() => {
		onMessageRef.current = onMessage;
	}, [onMessage]);

	useEffect(() => {
		if (!userId) return;

		shouldReconnect.current = true;

		const connect = () => {
			const ws = new WebSocket(`ws://localhost:8080/ws?user_id=${userId}`);
			socketRef.current = ws;

			ws.onopen = () => {
				reconnectAttempts.current = 0;
				console.log("Notification socket connected");
			};

			ws.onmessage = (event) => {
				try {
					const parsed = JSON.parse(event.data);
					if (
						parsed &&
						typeof parsed === "object" &&
						parsed.type === "new_notification"
					) {
						onMessageRef.current(parsed as IncomingNotificationWSMessage);
					} else {
						onMessageRef.current(parsed);
					}
				} catch (err) {
					console.error("Notification socket parse error", err);
					onMessageRef.current(event.data);
				}
			};

			ws.onclose = (ev) => {
				socketRef.current = null;
				console.log("Notification socket disconnected", ev.code, ev.reason);
				if (shouldReconnect.current) {
					reconnectAttempts.current += 1;
					const timeout = Math.min(30000, 1000 * 2 ** reconnectAttempts.current);
					reconnectTimer.current = window.setTimeout(() => connect(), timeout);
					console.log(`Reconnecting notification socket in ${timeout}ms`);
				}
			};

			ws.onerror = (err) => {
				console.log("Notification socket error", err);
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
				} catch {
					/* ignore */
				}
				socketRef.current = null;
			}
		};
	}, [userId]);

	return socketRef;
}
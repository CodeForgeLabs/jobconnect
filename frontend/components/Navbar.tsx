"use client";

import Image from "next/image";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState } from "react";
import { Bell, MessageCircle, Search, User, X } from "lucide-react";

import logo from "@/assets/Background.svg";
import { useGetMeQuery, useLogoutMutation } from "@/api/userapi";
import { useRouter } from "next/navigation";
import {
  useGetNotificationsQuery,
  useMarkNotificationAsReadMutation,
  type Notification,
} from "@/api/notificationapi";
import { useNotificationSocket } from "@/hooks/useNotificationSocket";
const Navbar = () => {
  const [searchQuery, setSearchQuery] = useState("");
  const [wsNotifications, setWsNotifications] = useState<Notification[]>([]);
  const [readNotificationIds, setReadNotificationIds] = useState<number[]>([]);
  const pathname = usePathname();
  const { data: userData } = useGetMeQuery();
  const [logout] = useLogoutMutation();
  const router = useRouter();
  const { data: notificationData = [] } = useGetNotificationsQuery(undefined, {
    skip: !userData,
  });
  const [markNotificationAsRead] = useMarkNotificationAsReadMutation();

  const notifications = [...wsNotifications, ...notificationData].reduce<
    Notification[]
  >((accumulator, notification) => {
    if (
      accumulator.some(
        (existingNotification) => existingNotification.id === notification.id,
      )
    ) {
      return accumulator;
    }

    const isRead =
      notification.is_read || readNotificationIds.includes(notification.id);

    return [
      ...accumulator,
      {
        ...notification,
        is_read: isRead,
      },
    ];
  }, []);

  const isLoggedIn = !!userData;
  const isFreelancer = userData?.role === "FREELANCER";
  const isClient = userData?.role === "CLIENT";
  const unreadCount = notifications.filter(
    (notification) => !notification.is_read,
  ).length;

  const handleClear = () => {
    setSearchQuery("");
  };

  useNotificationSocket(userData?.id ?? 0, (message) => {
    const payload =
      message && typeof message === "object" && "data" in message
        ? message.data
        : message;

    if (!payload || typeof payload !== "object" || !("id" in payload)) {
      return;
    }

    const nextNotification = payload as Notification;

    setWsNotifications((previousNotifications) => {
      console.log("Received WS notification:", previousNotifications);
      const withoutDuplicate = notifications.filter(
        (notification) => notification.id !== nextNotification.id,
      );

      return [nextNotification, ...withoutDuplicate];
    });
  });

  const handleMarkAsRead = async (notificationId: number) => {
    try {
      await markNotificationAsRead({ id: notificationId }).unwrap();
      setReadNotificationIds((current) =>
        current.includes(notificationId)
          ? current
          : [...current, notificationId],
      );
    } catch (error) {
      console.error("Failed to mark notification as read", error);
    }
  };

  return (
    <div className="navbar sticky top-0 z-50 min-h-7 h-12 py-0 px-6  w-full  bg-surface-container-low/70 glass-nav shadow-[0_8px_30px_rgb(13,28,46,0.04)]">
      <div className="flex flex-1 items-center gap-4">
        <Link href="/" className="flex items-center gap-2">
          <Image src={logo} alt="logo" className="w-8 h-8" />
          <span className="btn btn-ghost text-xl p-0">JobConnect</span>
        </Link>

        {isLoggedIn && (
          <div className="relative flex items-center w-full max-w-md">
            <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
              <Search className="w-5 h-5 text-gray-400" />
            </div>
            <input
              type="text"
              className="block w-full py-2 pl-10 pr-10 text-sm text-gray-900 border border-gray-300 rounded-lg bg-gray-50 focus:ring-blue-500 focus:border-blue-500 focus:outline-none transition-all duration-200"
              placeholder="Search for jobs, skills, or talent..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
            {searchQuery && (
              <button
                onClick={handleClear}
                className="absolute inset-y-0 right-0 flex items-center pr-3 text-gray-400 hover:text-gray-600"
              >
                <X className="w-4 h-4" />
              </button>
            )}
          </div>
        )}
      </div>

      <div className="flex items-center gap-1">
        {isLoggedIn ? (
          <>
            {isFreelancer && (
              <>
                <Link
                  href="/freelancer/dashboard"
                  className={`btn btn-sm bg-transparent border-none hover:text-black ${
                    pathname === "/freelancer/dashboard" ? "text-blue-600" : ""
                  }`}
                >
                  Dashboard
                </Link>
                <Link
                  href="/freelancer/jobsearch"
                  className={`btn btn-sm bg-transparent border-none hover:text-black ${
                    pathname === "/freelancer/jobsearch" ? "text-blue-600" : ""
                  }`}
                >
                  Find Work
                </Link>
                <Link
                  href="/freelancer/myproposals"
                  className={`btn btn-sm bg-transparent border-none hover:text-black ${
                    pathname === "/freelancer/myproposals"
                      ? "text-blue-600"
                      : ""
                  }`}
                >
                  Proposals
                </Link>

                <Link
                  href="/freelancer/mycontracts"
                  className={`btn btn-sm bg-transparent border-none hover:text-black ${
                    pathname === "/freelancer/mycontracts"
                      ? "text-blue-600"
                      : ""
                  }`}
                >
                  My Contracts
                </Link>
              </>
            )}

            {isClient && (
              <>
                <Link
                  href="/client/dashboard"
                  className={`btn btn-sm bg-transparent border-none hover:text-black ${
                    pathname === "/client/dashboard" ? "text-blue-600" : ""
                  }`}
                >
                  Dashboard
                </Link>
                <Link
                  href="/client/findtalent"
                  className={`btn btn-sm bg-transparent border-none hover:text-black ${
                    pathname === "/client/find-talent" ? "text-blue-600" : ""
                  }`}
                >
                  Find Talent
                </Link>

                <Link
                  href="/client/mypostings"
                  className={`btn btn-sm bg-transparent border-none hover:text-black ${
                    pathname === "/client/mypostings" ? "text-blue-600" : ""
                  }`}
                >
                  Postings
                </Link>
                <Link
                  href="/client/mycontracts"
                  className={`btn btn-sm bg-transparent border-none hover:text-black ${
                    pathname === "/client/mycontracts" ? "text-blue-600" : ""
                  }`}
                >
                  Contracts
                </Link>
              </>
            )}
            <Link
              href="/messages"
              className="btn btn-sm bg-transparent border-none hover:text-black flex items-center gap-1"
            >
              Messages
              <MessageCircle className="h-3 w-3" />
            </Link>

            <div className="dropdown dropdown-end   z-50">
              <div
                tabIndex={0}
                role="button"
                className="btn btn-ghost btn-circle relative border border-transparent hover:border-red-200 hover:bg-red-50"
                // onClick={() => handleMarkAsRead(0)}
              >
                <Bell className="h-4 w-4 text-slate-700" />
                {unreadCount > 0 ? (
                  <span className="absolute right-1 top-1 h-2.5 w-2.5 rounded-full bg-red-500 ring-2 ring-white animate-pulse" />
                ) : null}
              </div>

              <div className="dropdown-content  z-50 mt-3 w-84 rounded-2xl border border-slate-200 bg-white shadow-2xl">
                <div className="flex items-center justify-between border-b border-slate-100 px-4 py-3">
                  <div>
                    <p className="text-sm font-semibold text-slate-900">
                      Notifications
                    </p>
                    <p className="text-xs text-slate-500">
                      {unreadCount} unread
                    </p>
                  </div>
                </div>

                <div className="max-h-78 overflow-y-auto scroll-smooth p-2">
                  {notifications.length === 0 ? (
                    <div className="rounded-xl border border-dashed border-slate-200 px-4 py-6 text-center text-sm text-slate-500">
                      No notifications yet.
                    </div>
                  ) : (
                    <ul className="">
                      {notifications.map((notification) => (
                        <li key={notification.id}>
                          <button
                            type="button"
                            onClick={() => handleMarkAsRead(notification.id)}
                            className={`w-full border-y px-2 py-3 text-left transition hover:border-slate-300 hover:bg-slate-50 ${
                              notification.is_read
                                ? "border-slate-100 bg-white"
                                : "border-red-100 bg-red-50/70"
                            }`}
                          >
                            <div className="flex items-start justify-between gap-2">
                              <div className="min-w-0">
                                <p className="truncate text-sm font-semibold text-slate-900">
                                  {notification.title}
                                </p>
                                <p className="mt-1 text-sm text-slate-600 line-clamp-2">
                                  {notification.message}
                                </p>
                              </div>
                              <span className="text-green-500 text-xs">
                                {notification.is_read ? "" : "New"}
                              </span>
                            </div>
                          </button>
                        </li>
                      ))}
                    </ul>
                  )}
                </div>
              </div>
            </div>

            <div className="dropdown dropdown-end relative z-40">
              <div
                tabIndex={0}
                role="button"
                className="btn btn-ghost btn-circle avatar"
              >
                <div className="w-8 h-8 rounded-full overflow-hidden bg-gray-100 flex items-center justify-center">
                  {userData?.profile_picture_url ? (
                    <Image
                      src={userData.profile_picture_url}
                      alt="Profile picture"
                      width={32}
                      height={32}
                      className="h-full w-full object-cover"
                    />
                  ) : (
                    <User className="h-4 w-4 text-gray-500" />
                  )}
                </div>
              </div>

              <ul className="menu menu-sm dropdown-content bg-base-100 rounded-box z-40 mt-3 w-52 p-2 shadow">
                <li>
                  <Link href="/freelancer/profile">Profile</Link>
                </li>
                <li>
                  <Link href="/freelancer/wallet">Wallet</Link>
                </li>
                <li>
                  <button
                    onClick={async () => {
                      try {
                        const res = await logout(undefined);

                        if ("error" in res) {
                          console.error("Logout failed:", res.error);
                          return;
                        }

                        router.push("/login");
                      } catch (error) {
                        console.error(error);
                      }
                    }}
                    className="w-full text-left"
                  >
                    Logout
                  </button>
                </li>
              </ul>
            </div>
          </>
        ) : (
          <>
            <Link
              href="/login"
              className="btn btn-sm bg-transparent border-none hover:text-black"
            >
              Login
            </Link>
            <Link
              href="/signup"
              className="btn btn-sm bg-jobBlue text-white border-none hover:text-blue-600"
            >
              Sign Up
            </Link>
          </>
        )}
      </div>
    </div>
  );
};

export default Navbar;

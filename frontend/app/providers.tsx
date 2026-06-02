"use client";

import { usePathname } from "next/navigation";

import Footer from "@/components/Footer";
import Navbar from "@/components/Navbar";
import StoreProvider from "@/store/StoreProvider";

export default function AppShell({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const pathname = usePathname();
  const hideFooter = pathname.startsWith("/messages");

  return (
    <StoreProvider>
      <Navbar />
      {children}
      {!hideFooter && <Footer />}
    </StoreProvider>
  );
}

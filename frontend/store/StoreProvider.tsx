"use client";

import type { ReactNode } from "react";
import { Provider } from "react-redux";
import { store } from "@/store/store";

interface StoreProviderProps {
  children: ReactNode;
}

export default function StoreProvider({ children }: StoreProviderProps) {
  return <Provider store={store}>{children}</Provider>;
}

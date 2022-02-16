import React, { createContext, ReactNode, useContext, useMemo, useState } from "react";
import { useMounted } from "./Hooks";

const RenderContext = createContext<{
  renderer: Dispatcher<number>;
}>(undefined);

export const createRenderer = () => {
  // eslint-disable-next-line react-hooks/rules-of-hooks
  const [, renderer] = useState(0);
  // eslint-disable-next-line react-hooks/rules-of-hooks
  const mountedRef = useMounted();
  return () => {
    if (!mountedRef.current) return;
    renderer(x => x + 1);
  };
};

export const WithRenderer = ({ children }: { children: ReactNode }) => {
  const [render, renderer] = useState(0);
  const context = useMemo(() => ({ renderer }), [render]);
  return <RenderContext.Provider value={context}>{children}</RenderContext.Provider>;
};

export const useRenderer = () => {
  const { renderer } = useContext(RenderContext);
  const mountedRef = useMounted();
  return () => {
    if (!mountedRef.current) return;
    renderer(x => x + 1);
  };
};

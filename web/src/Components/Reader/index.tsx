import React, { createRef, useEffect, useMemo, useRef } from "react";
import { Helmet } from "react-helmet";
import { useMutableMemo } from "../Hooks";
import { createRenderer } from "../Renderer";
import Main from "./Main";
import ReaderContext from "./ReaderContext";
import Sidebar from "./Sidebar";

const Reader = ({ data, pref }: { data: ReaderData; pref: Preference }) => {
  const currentPageRef = useRef<PageState>();
  const pagesRef = useMutableMemo<PageState[]>(() => {
    const pages = data.chapter.pages.map<PageState>((fn, i) => ({
      ref: createRef<HTMLDivElement>(),
      fileName: fn,
      index: i
    }));

    let currentPage = parseInt(window.location.pathname.split("/")[3], 10) || 0;
    currentPage = Math.max(1, Math.min(currentPage, pages.length)) - 1;

    pages[currentPage].isViewing = true;
    currentPageRef.current = pages[currentPage];

    return pages;
  }, []);

  const isTransitionRef = useRef(false);
  const isLoadingRef = useRef(true);
  const render = createRenderer();

  const context = useMemo(
    () => ({
      pref,

      chapter: data.chapter,
      chapters: data.chapters,
      pagination: data.pagination,

      pagesRef,
      currentPageRef,

      isTransitionRef,
      isLoadingRef,
      render
    }),
    [render]
  );

  useEffect(() => {
    window.setTimeout(() => {
      isLoadingRef.current = false;
      render();
    }, 250);
  }, []);

  const title = useMemo(() => document.title, []);
  return (
    <ReaderContext.Provider value={context}>
      {!isLoadingRef.current && currentPageRef.current && (
        <Helmet>
          <title>
            Page {(currentPageRef.current.index + 1).toString()}: {title}
          </title>
        </Helmet>
      )}
      <Sidebar />
      <Main />
    </ReaderContext.Provider>
  );
};

export default Reader;

import React, { useCallback, useContext, useEffect, useMemo, useRef } from "react";
import { useHistory } from "react-router";
import { useMutableMemo, useNavigate } from "../Hooks";
import { WithIntersectionObserver } from "../IntersectionObserver";
import Spinner from "../Spinner";
import { PageDirection, PageScale } from "./constants";
import ReaderContext from "./ReaderContext";

const Page = ({ data }: { data: PageState }) => {
  const { pref, currentPageRef, isLoadingRef, render } = useContext(ReaderContext);
  const { prev, next } = useNavigate(render);
  const renderRef = useRef(render);

  const { ref } = data;
  const dataRef = useMutableMemo(() => data, [data]);
  const imageRef = useRef<HTMLImageElement>();

  const remainderRef = useRef(0);
  const lastPosRef = useRef(0);
  const posRef = useRef(0);
  const isDragRef = useRef<boolean>();

  /**
   * Navigate between pages by clicking the image,
   * abort if user is dragging the image.
   */
  const onClick = useCallback(
    (ev: React.MouseEvent) => {
      const target = (ev.target instanceof HTMLImageElement ? ev.target.parentElement : ev.target) as HTMLDivElement;
      let isPrev = ev.screenX - target.getBoundingClientRect().x <= target.clientWidth / 2;
      if (pref.direction === PageDirection.RightToLeft) {
        isPrev = !isPrev;
      }

      if (isPrev) prev();
      else next();
    },
    [pref, prev, next]
  );

  const scrollIntoView = useCallback(() => {
    const { top } = document.body.getBoundingClientRect();
    let offset: number;
    if (imageRef.current) {
      offset = imageRef.current.getBoundingClientRect().top;
    } else {
      offset = ref.current.getBoundingClientRect().top;
    }
    window.scrollTo({ top: offset - top });
  }, []);

  /**
   * Image dragging logic
   */

  const onDrag = useCallback((ev: MouseEvent) => {
    // Maybe unnecessary
    remainderRef.current = ref.current.clientWidth - imageRef.current.clientWidth;

    const nextPos = posRef.current + (ev.clientX - lastPosRef.current);
    if (nextPos >= remainderRef.current && nextPos <= 0) {
      posRef.current = nextPos;
      imageRef.current.style.left = `${posRef.current}px`;
    }

    lastPosRef.current = ev.clientX;
    isDragRef.current = true;
  }, []);

  const onRelease = useCallback((ev: Event) => {
    window.setTimeout(() => (isDragRef.current = false), 0);

    ref.current.removeEventListener("mouseup", onRelease);
    ref.current.removeEventListener("mousemove", onDrag);
    ref.current.removeEventListener("mouseleave", onRelease);
    ref.current.removeEventListener("pointerup", onRelease);
    ref.current.removeEventListener("pointermove", onDrag);
    ref.current.removeEventListener("pointerleave", onRelease);
  }, []);

  const onPress = useCallback((ev: MouseEvent) => {
    lastPosRef.current = ev.clientX;

    ref.current.addEventListener("mouseup", onRelease);
    ref.current.addEventListener("mousemove", onDrag);
    ref.current.addEventListener("mouseleave", onRelease);
    ref.current.addEventListener("pointerup", onRelease);
    ref.current.addEventListener("pointermove", onDrag);
    ref.current.addEventListener("pointerleave", onRelease);
  }, []);

  const calcRemainder = useCallback(() => {
    if (!ref.current || !imageRef.current) {
      return;
    }

    remainderRef.current = ref.current.clientWidth - imageRef.current.clientWidth;
    if (remainderRef.current < 0) {
      if (posRef.current < remainderRef.current) {
        posRef.current = remainderRef.current;
      }
      ref.current.classList.add("draggable");
      ref.current.addEventListener("mousedown", onPress);
      ref.current.addEventListener("pointerdown", onPress);
    } else {
      posRef.current = 0;
      ref.current.classList.remove("draggable");
      ref.current.removeEventListener("mousedown", onPress);
      ref.current.removeEventListener("pointerdown", onPress);
    }
    imageRef.current.style.left = `${posRef.current}px`;
  }, []);

  /**
   * Calculate and re-calculate the remainder when
   * the image has been downloaded/loaded.
   *
   * It will also fire when the image is resized,
   * or when the sidebar is toggled.
   */
  useEffect(() => {
    if (!ref.current || !imageRef.current) {
      return undefined;
    }

    const container = ref.current.parentElement.parentElement;

    calcRemainder();
    window.addEventListener("resize", calcRemainder);
    container.addEventListener("transitionend", calcRemainder);

    return () => {
      window.removeEventListener("resize", calcRemainder);
      container.removeEventListener("transitionend", calcRemainder);
    };
  }, [ref.current, imageRef.current, pref, data.isDownloaded]);

  /**
   * (Left-to-right or right-to-left mode)
   * Reset scroll position when navigating between pages.
   */
  useEffect(() => {
    if (!data.isViewing || pref.direction === PageDirection.TopToBottom) {
      return;
    }
    window.setTimeout(() => scrollIntoView(), 0);
  }, [data.isViewing]);

  /**
   * OnMount
   *
   * (Top-to-bottom mode)
   * Restore scroll position to the last viewed page.
   */
  useEffect(() => {
    if (pref.direction !== PageDirection.TopToBottom || !data.isViewing) {
      return;
    }
    const cancelRef = { current: false };
    const onScroll = () => {
      cancelRef.current = true;
      window.removeEventListener("scroll", onScroll);
    };

    window.setTimeout(() => {
      const interval = setInterval(() => {
        if (cancelRef.current || imageRef.current) {
          clearInterval(interval);
        }
        if (!imageRef.current) return;
        for (let i = 0; i < 5; i++) {
          if (cancelRef.current) break;
          window.setTimeout(() => {
            if (cancelRef.current) return;
            scrollIntoView();
          }, 100 * i);
        }
        window.addEventListener("scroll", onScroll);
      }, 100);
    }, 250 + Math.random());
  }, []);

  /**
   * (Top-to-bottom mode)
   * Attach IntersectionObserver to the image.
   */
  useEffect(() => {
    if (pref.direction !== PageDirection.TopToBottom) {
      return undefined;
    }

    const observer = new IntersectionObserver(
      entries =>
        entries.forEach(entry => {
          if (isLoadingRef.current) return;

          let doRender = false;
          if (entry.isIntersecting && !dataRef.current.isViewing) {
            dataRef.current.isViewing = true;
            currentPageRef.current = dataRef.current;
            doRender = true;
          } else if (!entry.isIntersecting && dataRef.current.isViewing) {
            dataRef.current.isViewing = false;
            doRender = true;
          }

          if (doRender) {
            window.setTimeout(() => {
              renderRef.current();
            }, 250 + Math.random());
          }
        }),
      { rootMargin: "0px 0px -100% 0px" }
    );
    observer.observe(dataRef.current.ref.current);

    return () => {
      if (observer) observer.disconnect();
    };
  }, []);

  if (pref.direction !== PageDirection.TopToBottom && !data.isViewing) {
    return null;
  }

  return (
    <div className="page" ref={ref}>
      {data.isDownloaded ? (
        <div className="wrapper" onClickCapture={pref.navigateOnClick ? onClick : undefined}>
          <img src={data.url} alt={`Page ${data.index + 1}`} ref={imageRef} />
        </div>
      ) : (
        // eslint-disable-next-line react/jsx-no-useless-fragment
        <>
          {data.isFailed ? (
            <div className="failed">
              <button
                type="button"
                onClick={() => {
                  dataRef.current.isFailed = false;
                  for (let i = 0; i < 3; i++) {
                    window.setTimeout(() => {
                      renderRef.current();
                    }, 100 * i);
                  }
                }}
              >
                <strong>Failed to load page</strong>
                <span>Click to retry</span>
              </button>
            </div>
          ) : (
            <WithIntersectionObserver className="loading">
              <Spinner width="120" height="120" strokeWidth="8" />
            </WithIntersectionObserver>
          )}
        </>
      )}
    </div>
  );
};

const Main = () => {
  const { pref, chapter, pagination, pagesRef, currentPageRef, isLoadingRef, render } = useContext(ReaderContext);
  const history = useHistory();
  const renderRef = useRef(render);

  const queuesRef = useRef<PageState[]>([]);
  const parallelSizeRef = useRef(0);

  const styles: any = useMemo(() => {
    const maxWidth = Number(pref.maxWidth);
    const maxHeight = Number(pref.maxHeight);
    return {
      "--gaps": `${Number(pref.gaps) / 10}rem`,
      "--zoom": pref.zoom,
      "--max-width": maxWidth ? `${maxWidth / 10}rem` : undefined,
      "--max-height": maxHeight ? `${maxHeight / 10}rem` : undefined
    };
  }, [pref.gaps, pref.maxHeight, pref.maxWidth, pref.zoom]);

  useEffect(() => {
    const timeout = window.setTimeout(() => {
      history.replace(`/chapters/${chapter.id}/${currentPageRef.current.index + 1}`);
    }, 250 + Math.random());
    return () => {
      window.clearTimeout(timeout);
    };
  }, [currentPageRef.current]);

  useEffect(() => {
    if (isLoadingRef.current || !pagesRef.current?.length || !currentPageRef.current) {
      return;
    }

    const logQueue = (page: PageState) => {
      console.info(
        "[Reader] Page %d has been added to queue | Queue size: %d/%d",
        page.index + 1,
        queuesRef.current.length,
        pref.maxPreloads
      );
    };

    const oldQueues = Array.from(queuesRef.current);
    queuesRef.current = queuesRef.current.filter(q => !q.isDownloaded && !q.isFailed);

    if (
      !currentPageRef.current.isDownloaded &&
      !currentPageRef.current.isFailed &&
      !queuesRef.current.includes(currentPageRef.current)
    ) {
      queuesRef.current.push(currentPageRef.current);
      logQueue(currentPageRef.current);
    }

    for (let i = 0; queuesRef.current.length < pref.maxPreloads && i < pref.maxPreloads; i++) {
      const prev = pagesRef.current[i];
      if (prev && !prev.isDownloaded && !prev.isFailed && !queuesRef.current.includes(prev)) {
        queuesRef.current.push(prev);
        logQueue(prev);
      }

      if (queuesRef.current.length < pref.maxPreloads) {
        const next = pagesRef.current[currentPageRef.current.index + i];
        if (next && !next.isDownloaded && !next.isFailed && !queuesRef.current.includes(next)) {
          queuesRef.current.push(next);
          logQueue(next);
        }
      }
    }

    if (
      queuesRef.current.length !== oldQueues.length ||
      !queuesRef.current.every((q, i) => q.index === oldQueues[i].index)
    ) {
      window.setTimeout(() => {
        renderRef.current();
      }, 250 + Math.random());
    }
  }, [
    isLoadingRef.current,
    currentPageRef.current,
    ...pagesRef.current.map(p => p.isDownloaded || p.isDownloading || p.isFailed)
  ]);

  useEffect(() => {
    if (
      isLoadingRef.current ||
      !pagesRef.current?.length ||
      !queuesRef.current.length ||
      parallelSizeRef.current >= pref.maxParallel
    ) {
      return;
    }

    for (let i = 0; i < queuesRef.current.length && parallelSizeRef.current < pref.maxParallel; i++) {
      const page = queuesRef.current[i];
      if (page.isDownloaded || page.isDownloading || page.isFailed) {
        continue;
      }

      page.isDownloading = true;
      parallelSizeRef.current++;

      console.info(
        "[Reader] Preloading page %d | Parallel size: %d/%d",
        page.index + 1,
        parallelSizeRef.current,
        pref.maxParallel
      );

      (async () => {
        await new Promise<void>((resolve, reject) => {
          let retryCount = 0;

          const load = () => {
            retryCount++;

            const xhr = new XMLHttpRequest();
            xhr.open("GET", `/pages/${chapter.id}/${page.fileName}`);
            xhr.responseType = "blob";

            const onLoaded = () => {
              page.url = URL.createObjectURL(xhr.response);
              page.isDownloaded = true;
              resolve();
            };

            const onError = () => {
              if (retryCount > 6) {
                page.isFailed = true;
                resolve();
              } else {
                setTimeout(load, 1000);
              }
            };

            xhr.addEventListener("load", onLoaded);
            xhr.addEventListener("error", onError);

            xhr.send();
          };
          load();
        });

        page.isDownloading = false;
        parallelSizeRef.current--;

        if (page.isDownloaded) {
          console.info(
            "[Reader] Page %d has been preloaded | Parallel size: %d/%d",
            page.index + 1,
            parallelSizeRef.current,
            pref.maxParallel
          );
        } else {
          console.info(
            "[Reader] Failed to preload page %d | Parallel size: %d/%d",
            page.index + 1,
            parallelSizeRef.current,
            pref.maxParallel
          );
        }
        window.setTimeout(() => {
          renderRef.current();
        }, 1000 * Math.random());
      })();
    }
  }, [isLoadingRef.current, queuesRef.current]);

  /**
   * Restore scroll position when changing page direction.
   */

  const ref = useRef<HTMLDivElement>();
  const offsetRef = useRef(0);

  useEffect(() => {
    if (!currentPageRef.current?.ref?.current) {
      return;
    }

    const currentPageElement = currentPageRef.current.ref.current;
    if (offsetRef.current > 0) {
      currentPageElement.scrollIntoView();
    } else {
      const { top } = document.body.getBoundingClientRect();
      const offset = currentPageElement.getBoundingClientRect().top;

      window.scrollTo({ top: offset - top - offsetRef.current });
      window.setTimeout(() => (offsetRef.current = offset), 0);
    }
  }, [pref.direction]);

  // Store scroll position on scroll.
  useEffect(() => {
    const onScroll = () => {
      if (!currentPageRef.current?.ref?.current) return;
      offsetRef.current = currentPageRef.current.ref.current.getBoundingClientRect().top;
    };
    window.addEventListener("scroll", onScroll);
  }, []);

  return (
    <main ref={ref}>
      <div className="top">
        <button
          className="toggle"
          type="button"
          onClick={() => {
            pref.showSidebar = !pref.showSidebar;
            localStorage.setItem("pref", JSON.stringify(pref));
            renderRef.current();
          }}
        />
      </div>
      <div
        className="pages"
        data-scale={PageScale[pref.scale]}
        data-single={pref.direction !== PageDirection.TopToBottom || undefined}
        style={styles}
      >
        {pagesRef.current.map(page => (
          <Page data={page} key={`page-${page.index}`} />
        ))}
      </div>
      {pagination.Next &&
        (pref.direction === PageDirection.TopToBottom ||
          currentPageRef.current?.index === pagesRef.current.length - 1) && (
          <footer>
            <a className="next" href={`/chapters/${pagination.Next.id}`}>
              Next Chapter
            </a>
          </footer>
        )}
    </main>
  );
};

export default Main;

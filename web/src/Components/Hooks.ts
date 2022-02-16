import { DependencyList, useContext, useEffect, useMemo, useRef, useState } from "react";
import { useHistory } from "react-router";
import { GetAuthors, GetTags } from "../api";
import { HasPerms } from "../utils/utils";
import ManageContext from "./Manage/ManageContext";
import ReaderContext from "./Reader/ReaderContext";
import { useToast } from "./Toast";

interface IntersectionObserverOptions extends IntersectionObserverInit {
  once?: boolean;
}

export const useIntersectionObserver = <T extends HTMLElement>(
  callbackfn: (entry: IntersectionObserverEntry) => void,
  options?: IntersectionObserverOptions
): Mutable<T> => {
  const ref = useRef<T>();
  useEffect(() => {
    const observer = new IntersectionObserver(entries => {
      entries.forEach(entry => {
        if (options?.once) {
          if (entry.isIntersecting) {
            observer.disconnect();
          } else {
            return;
          }
        }
        callbackfn(entry);
      });
    }, options);
    if (ref.current) {
      observer.observe(ref.current);
    }
    return () => {
      observer.disconnect();
    };
  }, [callbackfn, options]);
  return ref;
};

export const useModal = <T extends HTMLElement>(ref: Mutable<T>, stateRef: Mutable<boolean>, render: Renderer) => {
  useEffect(() => {
    const listener = (ev: MouseEvent) => {
      if (ref.current && !ref.current.contains(ev.target as HTMLElement)) {
        stateRef.current = false;
        render();
      }
    };
    if (ref.current) {
      document.addEventListener("click", listener);
    }
    return () => {
      document.removeEventListener("click", listener);
    };
  }, [ref.current, stateRef.current]);
};

export const useMounted = (): Mutable<boolean> => {
  const mountedRef = useRef<boolean>();
  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
    };
  }, []);
  return mountedRef;
};

export const useMutableMemo = <T>(callbackfn: (prev?: T) => T, deps: DependencyList): Mutable<T> => {
  const ref = useRef<T>();
  useMemo(() => {
    ref.current = callbackfn(ref.current);
  }, deps);
  return ref;
};

export const useKeydown = (callbackfn: (ev: KeyboardEvent) => void) => {
  useEffect(() => {
    window.addEventListener("keydown", callbackfn);
    return () => {
      window.removeEventListener("keydown", callbackfn);
    };
  }, [callbackfn]);
};

export const usePermissions = (permissions: string[], redirect = "/settings"): { user; history } => {
  const { user } = useContext(ManageContext);
  const history = useHistory();

  if (!useMemo(() => HasPerms(user, ...permissions), [])) {
    history.replace(redirect);
  }
  return { user, history };
};

const tagsCache: Entity[] = [];
export const useTags = (): [Entity[], boolean] => {
  const toast = useToast();
  const [tags, setTags] = useState<Tag[]>(tagsCache);
  const [isLoading, setIsLoading] = useState(!tags.length);

  useEffect(() => {
    if (tags.length) return;
    GetTags().then(({ response, error }) => {
      if (response) {
        setTags(response);
        tagsCache.push(...response);
      }
      if (error) toast.showError(error);
      setIsLoading(false);
    });
  }, []);

  return [tags, isLoading];
};

export const useAuthors = (): [Entity[], boolean] => {
  const toast = useToast();
  const [authors, setAuthors] = useState<Author[]>();
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    GetAuthors().then(({ response, error }) => {
      if (response) setAuthors(authors);
      if (error) toast.showError(error);
      setIsLoading(false);
    });
  }, []);

  return [authors, isLoading];
};

export const useNavigate = (render: Renderer) => {
  const { pagesRef, currentPageRef } = useContext(ReaderContext);

  const first = () => {
    if (!pagesRef.current.length || currentPageRef.current.index === 0) {
      return;
    }

    pagesRef.current.forEach(p => (p.isViewing = p.index === 0));
    const [firstPage] = pagesRef.current;
    firstPage.ref.current?.scrollIntoView();
    currentPageRef.current = firstPage;
    render();
  };

  const last = () => {
    if (!pagesRef.current.length || currentPageRef.current.index === pagesRef.current.length - 1) {
      return;
    }

    pagesRef.current.forEach(p => (p.isViewing = p.index === pagesRef.current.length - 1));
    const lastPage = pagesRef.current[pagesRef.current.length - 1];
    lastPage.ref.current?.scrollIntoView();
    currentPageRef.current = lastPage;
    render();
  };

  const prev = () => {
    if (!pagesRef.current.length || currentPageRef.current.index === 0) {
      return;
    }

    pagesRef.current.forEach(p => (p.isViewing = p.index === currentPageRef.current.index - 1));
    const prevPage = pagesRef.current[currentPageRef.current.index - 1];
    prevPage.ref.current?.scrollIntoView();
    currentPageRef.current = prevPage;
    render();
  };

  const next = () => {
    if (!pagesRef.current.length || currentPageRef.current.index === pagesRef.current.length - 1) {
      return;
    }

    pagesRef.current.forEach(p => (p.isViewing = p.index === currentPageRef.current.index + 1));
    const nextPage = pagesRef.current[currentPageRef.current.index + 1];
    nextPage.ref.current?.scrollIntoView();
    currentPageRef.current = nextPage;
    render();
  };

  const jump = (pageNum: number) => {
    if (
      !pagesRef.current.length ||
      pageNum < 1 ||
      pageNum > pagesRef.current.length ||
      pageNum === currentPageRef.current.index + 1
    ) {
      return;
    }

    pagesRef.current.forEach(p => (p.isViewing = p.index === pageNum - 1));
    const page = pagesRef.current[pageNum - 1];
    page.ref.current?.scrollIntoView();
    currentPageRef.current = page;
    render();
  };

  return { last, prev, next, first, jump };
};

export default {};

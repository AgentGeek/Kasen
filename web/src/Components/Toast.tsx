import React, {
  createContext,
  createRef,
  isValidElement,
  memo,
  ReactNode,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState
} from "react";
import { X } from "react-feather";

interface ToastNode {
  className?: string;
  type?: string;
  content: ReactNode;
}

interface ToastEntry {
  id: number;
  ref: Mutable<HTMLDivElement>;
}

const instance = createContext<{
  nodesRef: Mutable<ToastEntry[]>;

  show: (v: ToastNode | ReactNode) => Promise<number>;
  showError: (error: ApiError) => Promise<number>;
  hide: (id: number) => void;
}>(undefined);

const Toast = ({ node }: { node: ToastNode & ToastEntry }) => {
  const { hide } = useContext(instance);
  const [isInvisible, setIsInvisible] = useState(true);

  useEffect(() => {
    setTimeout(() => setIsInvisible(false), 0);
    const timeout = window.setTimeout(() => hide(node.id), 5000);
    return () => {
      clearTimeout(timeout);
    };
  }, []);

  return (
    <div
      className={`toast ${node.type || ""} ${node.className || ""}`}
      data-invisible={isInvisible || undefined}
      id={`t${node.id}`}
      ref={node.ref}
    >
      <div className="toastContentWrapper">{node.content}</div>
      <button type="button" onClick={() => hide(node.id)}>
        <X width="20" height="20" strokeWidth="3" />
      </button>
    </div>
  );
};

const MemoizedToast = memo(Toast, (prev, next) => prev.node.id === next.node.id);

export const useToast = () => {
  const { show, showError, hide } = useContext(instance);
  return { show, showError, hide };
};

export const WithToast = ({ children }: { children: ReactNode }) => {
  const nodesRef = useRef<(ToastNode & ToastEntry)[]>([]);
  const [, render] = useState(false);
  const mutex = useRef(false);

  const hide = useCallback((id: number) => {
    const node = nodesRef.current.find(m => m.id === id);
    if (node) {
      node.ref.current.setAttribute("data-invisible", "");
      window.setTimeout(() => {
        const idx = nodesRef.current.findIndex(m => m.id === id);
        if (idx >= 0) {
          nodesRef.current.splice(idx, 1);
          render(x => !x);
        }
      }, 250);
    }
  }, []);

  const show = useCallback(async (node: ToastNode | ReactNode) => {
    if (mutex.current) {
      await new Promise<void>(resolve => {
        const interval = window.setInterval(() => {
          if (mutex.current) return;
          clearInterval(interval);
          resolve();
        }, 250);
      });
    }
    mutex.current = true;

    const id = Math.floor(new Date().valueOf() + Math.random() * 1000);
    const ref = createRef<HTMLDivElement>();

    if (nodesRef.current.length >= 3) {
      hide(nodesRef.current[nodesRef.current.length - 1].id);
      await new Promise<void>(resolve => {
        setTimeout(resolve, 250);
      });
    }

    if (isValidElement(node)) {
      nodesRef.current.unshift({ id, ref, content: node });
    } else if (typeof node === "string") {
      nodesRef.current.unshift({ id, ref, content: <p>{node}</p> });
    } else {
      nodesRef.current.unshift({ id, ref, ...(node as ToastNode) });
    }

    mutex.current = false;
    render(x => !x);

    return id;
  }, []);

  const showError = useCallback(
    (error: ApiError) =>
      show({
        type: "error",
        content: (
          <p>
            {error.message}: {error.cause}
          </p>
        )
      }),
    []
  );

  const context = useMemo(
    () => ({
      nodesRef,

      show,
      showError,
      hide
    }),
    []
  );

  return (
    <instance.Provider value={context}>
      {children}
      {!!nodesRef.current.length && (
        <div id="toasts">
          <div className="toastsWrapper">
            {nodesRef.current.map(toast => (
              <MemoizedToast node={toast} key={`toast-${toast.id}`} />
            ))}
          </div>
        </div>
      )}
    </instance.Provider>
  );
};

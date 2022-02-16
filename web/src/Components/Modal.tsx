import React, {
  createContext,
  isValidElement,
  ReactNode,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState
} from "react";
import { useMounted } from "./Hooks";

interface ModalNode {
  className?: string;
  title?: string;
  content: ReactNode;
}

const instance = createContext<{
  ref: Mutable<HTMLDivElement>;
  node: ModalNode;
  setNode: Dispatcher<ModalNode>;

  show: (v: ModalNode | ReactNode) => void;
  hide: () => void;
}>(undefined);

const showScroll = () => {
  document.body.style.overflow = "";
  if (!document.body.style.length) {
    document.body.removeAttribute("style");
  }
};

const hideScroll = () => (document.body.style.overflow = "hidden");

const Modal = () => {
  const { ref, node, hide } = useContext(instance);
  const mountedRef = useMounted();

  useEffect(() => {
    hideScroll();
    window.setTimeout(() => {
      if (mountedRef.current) {
        ref.current.removeAttribute("data-invisible");
      }
    }, 0);
    return showScroll;
  }, []);

  return (
    <div className={node.className} id="modal" data-invisible ref={ref}>
      <div className="modalBg" onClick={hide} />
      <div className="modalContent">
        {node.title && (
          <div className="modalTitle">
            <b>{node.title}</b>
          </div>
        )}
        <div className="modalContentWrapper">{node.content}</div>
      </div>
    </div>
  );
};

export const useModal = () => {
  const { show, hide } = useContext(instance);
  return { show, hide };
};

export const WithModal = ({ children }: { children: ReactNode }) => {
  const [node, setNode] = useState<ModalNode>();
  const ref = useRef<HTMLDivElement>();
  const mountedRef = useMounted();

  const show = useCallback((v: ModalNode | ReactNode) => {
    if (isValidElement(v) || typeof v === "string") {
      setNode({ content: v });
    } else {
      setNode(v as ModalNode);
    }
  }, []);

  const hide = useCallback(() => {
    ref.current.setAttribute("data-invisible", "");
    window.setTimeout(() => {
      if (mountedRef.current) {
        setNode(undefined);
      }
    }, 250);
  }, []);

  useEffect(() => showScroll, []);

  const context = useMemo(
    () => ({
      node,
      setNode,
      ref,
      show,
      hide
    }),
    [node, setNode, ref, show, hide]
  );
  return (
    <instance.Provider value={context}>
      {children}
      {node && <Modal />}
    </instance.Provider>
  );
};

import React, { useEffect, useRef } from "react";
import { createRenderer } from "./Renderer";

interface WithIntersectionObserverProps extends Props<HTMLDivElement> {
  options?: IntersectionObserverInit;
  once?: boolean;
}

export const WithIntersectionObserver = ({
  className,
  children,
  options,
  once,
  ...props
}: WithIntersectionObserverProps) => {
  const ref = useRef<HTMLDivElement>();
  const isVisibleRef = useRef(false);
  const render = createRenderer();

  useEffect(() => {
    const observer = new IntersectionObserver(entries => {
      entries.forEach(entry => {
        if (entry.isIntersecting) {
          if (once) observer.disconnect();
          if (!isVisibleRef.current) {
            isVisibleRef.current = true;
            render();
          }
        } else if (isVisibleRef.current) {
          isVisibleRef.current = false;
          render();
        }
      });
    }, options);

    observer.observe(ref.current);
    return () => {
      observer.disconnect();
    };
  }, []);

  return (
    <div {...props} className={`wObs ${className || ""}`} data-hidden={!isVisibleRef.current || undefined} ref={ref}>
      {isVisibleRef.current && children}
    </div>
  );
};

export default {};

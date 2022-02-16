import React, { memo, MutableRefObject, ReactElement, SVGAttributes, useEffect, useState } from "react";
import { useMounted } from "./Hooks";

type SpinnerProps = SVGAttributes<SVGSVGElement>;

const Spinner = ({ className, ...props }: SpinnerProps) => (
  <div className={className}>
    <svg
      className="spinner"
      width="20"
      height="20"
      viewBox="0 0 150 150"
      fill="none"
      stroke="currentColor"
      strokeWidth="20"
      {...props}
    >
      <circle cx="75" cy="75" r="60" />
    </svg>
  </div>
);

interface WithSpinnerProps extends SpinnerProps {
  children?: ReactElement;
  dispatcherRef: MutableRefObject<Dispatcher<boolean>>;
}

export const WithSpinner = ({ dispatcherRef, children, ...props }: WithSpinnerProps) => {
  const mountedRef = useMounted();
  const [state, setState] = useState(false);

  useEffect(() => {
    let timeout = 0;
    dispatcherRef.current = (v: boolean) => {
      timeout = window.setTimeout(() => {
        if (mountedRef.current) {
          setState(v);
        }
      }, 250);
    };
    return () => {
      clearTimeout(timeout);
      dispatcherRef.current = (v: boolean) => {}; // noop
    };
  }, []);
  return state ? (
    <svg
      className="spinner"
      width="20"
      height="20"
      viewBox="0 0 150 150"
      fill="none"
      stroke="currentColor"
      strokeWidth="20"
      {...props}
    >
      <circle cx="75" cy="75" r="60" />
    </svg>
  ) : (
    children || null
  );
};

export default memo(Spinner, () => true);

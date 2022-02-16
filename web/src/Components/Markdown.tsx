import EasyMDE, { Options } from "easymde";
import "easymde/dist/easymde.min.css";
import React, { useEffect, useMemo, useRef, useState } from "react";
import { useMounted } from "./Hooks";

interface Markdown {
  markdown: Mutable<EasyMDE>;
  Markdown: (props: Props<HTMLTextAreaElement>) => JSX.Element;
}

const useMarkdown = (options: Options = {}): Markdown => {
  const markdown = useRef<EasyMDE>();

  const Markdown = (props: Props<HTMLTextAreaElement> = {}) => {
    const mountedRef = useMounted();
    const ref = useRef<HTMLTextAreaElement>();
    const [isInvisible, setIsInvisible] = useState(true);

    useEffect(() => {
      markdown.current = new EasyMDE({
        element: ref.current,
        status: false,
        spellChecker: false,
        toolbar: [
          "bold",
          "italic",
          "strikethrough",
          "|",
          "quote",
          "unordered-list",
          "ordered-list",
          "|",
          "link",
          "horizontal-rule",
          "|",
          "preview",
          "guide"
        ],
        ...(options || {})
      });
      window.setTimeout(() => {
        if (mountedRef.current) {
          setIsInvisible(false);
        }
      }, 0);
    }, []);

    return (
      <div className="markdownEditor" data-invisible={isInvisible || undefined}>
        <textarea {...props} ref={ref} />
      </div>
    );
  };

  return {
    markdown,
    Markdown: useMemo(() => Markdown, [])
  };
};

export default useMarkdown;

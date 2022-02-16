import React, { useContext, useRef } from "react";
import { UpdateMeta } from "../../../api";
import { createRenderer } from "../../Renderer";
import { WithSpinner } from "../../Spinner";
import { useToast } from "../../Toast";
import ManageContext from "../ManageContext";

const Meta = () => {
  const context = useContext(ManageContext);
  const { baseURL, title, description, language } = context;

  const toast = useToast();
  const mutex = useRef(false);
  const render = createRenderer();

  const isSubmittingRef = useRef<Dispatcher<boolean>>();
  const onSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    const target = e.target as HTMLFormElement;

    const meta = {
      baseURL: target.baseURL.value.trim(),
      title: target.t.value.trim(),
      description: target.description.value.trim(),
      language: target.language.value.trim()
    };

    if (
      mutex.current ||
      (meta.baseURL === baseURL &&
        meta.title === title &&
        meta.description === description &&
        meta.language === language)
    ) {
      return;
    }
    mutex.current = true;
    isSubmittingRef.current(true);

    await (async () => {
      const { error, status } = await UpdateMeta(meta);
      if (status === 204) {
        toast.show("Meta has been updated");

        context.baseURL = meta.baseURL;
        context.title = meta.title;
        context.description = meta.description;
        context.language = meta.language;

        render();
      }
      if (error) toast.showError(error);
    })();

    mutex.current = false;
    isSubmittingRef.current(false);
  };

  return (
    <section className="meta">
      <h3>Meta</h3>
      <form onSubmit={onSubmit}>
        <div className="inputContainer">
          <strong>Base URL</strong>
          <input className="input" type="text" name="baseURL" defaultValue={baseURL} placeholder="Required" required />
        </div>
        <div className="inputContainer">
          <strong>Title</strong>
          <input className="input" type="text" name="t" defaultValue={title} placeholder="Required" required />
        </div>
        <div className="inputContainer">
          <strong>Description</strong>
          <input
            className="input"
            type="text"
            name="description"
            defaultValue={description}
            placeholder="Required"
            required
          />
        </div>
        <div className="inputContainer">
          <strong>Language</strong>
          <input
            className="input"
            type="text"
            name="language"
            defaultValue={language}
            placeholder="Required"
            required
          />
        </div>
        <button className="button green" type="submit">
          <WithSpinner width="16" height="16" strokeWidth="8" dispatcherRef={isSubmittingRef} />
          <strong>Update Meta</strong>
        </button>
      </form>
    </section>
  );
};

export default Meta;

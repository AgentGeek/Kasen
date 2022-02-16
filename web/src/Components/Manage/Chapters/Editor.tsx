import deepEqual from "fast-deep-equal";
import React, { DragEvent, FormEvent, MouseEvent, useCallback, useEffect, useRef } from "react";
import { Plus, RefreshCw, Save, Trash, Upload, X } from "react-feather";
import { Helmet } from "react-helmet";
import { Link, useHistory } from "react-router-dom";
import { CreateChapter, DeletePage, GetChapterMd, GetPagesMd, UpdateChapter, UploadPage } from "../../../api";
import { FormatFileSize as FormatBytes, GetPageURL } from "../../../utils/utils";
import { useMounted } from "../../Hooks";
import { useRenderer } from "../../Renderer";
import Spinner, { WithSpinner } from "../../Spinner";
import { useToast } from "../../Toast";

interface QueueState {
  data?: File;
  url?: string;
  fileName: string;
  isUploading?: boolean;
  isUploaded?: boolean;
  isFailed?: boolean;
}

const Page = ({
  chapter,
  pagesRef,
  fileName,
  index
}: {
  chapter: Chapter;
  pagesRef: Mutable<string[]>;
  fileName: string;
  index: number;
}) => {
  const toast = useToast();

  const mutex = useRef(false);
  const mountedRef = useMounted();
  const render = useRenderer();

  const setIsDeletingRef = useRef<Dispatcher<boolean>>();
  const deleteBtnOnClick = useCallback(async (ev: MouseEvent) => {
    ev.preventDefault();
    ev.stopPropagation();

    if (mutex.current) return;
    setIsDeletingRef.current(true);
    mutex.current = true;

    await (async () => {
      const { response, error } = await DeletePage(chapter.id, fileName);
      if (!mountedRef.current) return;

      if (response) pagesRef.current = response;
      if (error) toast.showError(error);
      render();
    })();

    setIsDeletingRef.current(false);
    mutex.current = false;
  }, []);

  return (
    <div className="entry">
      <div className="thumbnail">
        <a href={GetPageURL(chapter, fileName)} target="_blank" rel="noreferrer">
          <img src={`${GetPageURL(chapter, fileName)}/64.jpg`} loading="lazy" />
        </a>
      </div>
      <strong className="order">#{index + 1}</strong>
      <div className="metadata">
        <h4 className="fileName">
          <a href={GetPageURL(chapter, fileName)} title={fileName} target="_blank" rel="noreferrer">
            {fileName}
          </a>
        </h4>
        <button className="delete" type="button" title="Delete page" onClick={deleteBtnOnClick}>
          <WithSpinner width="16" height="16" dispatcherRef={setIsDeletingRef}>
            <Trash width="16" height="16" strokeWidth="3" />
          </WithSpinner>
          <strong>Delete</strong>
        </button>
      </div>
    </div>
  );
};

const Queue = ({
  queue,
  queuesRef,
  chapter,
  pagesRef
}: {
  queue: QueueState;
  queuesRef: Mutable<QueueState[]>;
  chapter: Chapter;
  pagesRef: Mutable<string[]>;
}) => {
  const toast = useToast();

  const mountedRef = useMounted();
  const mutex = useRef(false);
  const render = useRenderer();

  const isCancelledRef = useRef(false);
  const progressBarRef = useRef<HTMLDivElement>();
  const speedRef = useRef<HTMLElement>();

  const deleteBtnOnClick = useCallback(() => {
    const idx = queuesRef.current.findIndex(f => f === queue);
    if (idx >= 0) {
      isCancelledRef.current = true;
      queuesRef.current.splice(idx, 1);
      render();
    }
  }, []);

  const retryBtnOnClick = useCallback(() => {
    queue.isFailed = false;
    queue.isUploading = true;
    render();
  }, []);

  useEffect(() => {
    if (!queue.isUploading || mutex.current) {
      return;
    }
    mutex.current = true;

    const request = UploadPage(chapter.id, { data: queue.data, url: queue.url });
    const start = new Date().getTime();

    request.upload.addEventListener("progress", ev => {
      const percentage = (ev.loaded / ev.total) * 100;
      if (!mountedRef.current && !(percentage % 100)) {
        if (!isCancelledRef.current && !queue.isUploaded) {
          toast.showError({
            message: "Unable to upload page",
            cause: "Operation was interrupted"
          });
        }
        request.cancel();
        return;
      }

      if (progressBarRef.current) {
        progressBarRef.current.style.width = `${percentage}%`;
      }

      if (speedRef.current) {
        const elapsed = (new Date().getTime() - start) / 1000;
        speedRef.current.textContent = `${FormatBytes(ev.loaded / elapsed)}/s`;
      }
    });

    request.then(({ response, error }) => {
      if (!mountedRef.current) return;

      if (response) {
        pagesRef.current = response;
        queue.isUploaded = true;
      } else if (error) queue.isFailed = true;

      if (error) {
        if (error.cause.toLowerCase().includes("already exists")) {
          queue.isUploaded = true;
          queue.isFailed = false;
        } else toast.showError(error);
      }

      mutex.current = false;
      queue.isUploading = false;
      render();
    });
  }, [queue.isUploading]);

  return (
    <div className="entry">
      {queue.isUploading && <div className="progress" ref={progressBarRef} />}
      <div className="wrapper">
        <div className="metadata">
          {queue.isFailed && <strong className="failed">[Failed]</strong>}
          <strong className="fileName" title={queue.fileName}>
            {queue.fileName}
          </strong>
          {queue.data && (
            <>
              {queue.isUploading && <span className="speed" ref={speedRef} />}
              <span className="fileSize">{FormatBytes(queue.data.size)}</span>
            </>
          )}
        </div>
        <button className="delete" type="button" title="Cancel" onClick={deleteBtnOnClick}>
          <X width="16" height="16" strokeWidth="3" />
        </button>
        {queue.isFailed && (
          <button className="retry button red" type="button" onClick={retryBtnOnClick}>
            <RefreshCw width="16" height="16" strokeWidth="3" />
          </button>
        )}
      </div>
    </div>
  );
};

interface EditorProps extends Props<HTMLDivElement> {
  chapter: Chapter;
  pagesRef: Mutable<string[]>;
  isNew?: boolean;
}

const createDraft = (chapter: Chapter): ChapterDraft => ({
  chapter: chapter.chapter || undefined,
  volume: chapter.volume || undefined,
  title: chapter.title || undefined,
  scanlationGroups: chapter.scanlationGroups?.map(e => e.name) || []
});

const mimeTypes = "image/png image/jpeg";

const Editor = ({ chapter, pagesRef, isNew, ...props }: EditorProps) => {
  const history = useHistory<{ queues: QueueState[] }>();
  const toast = useToast();

  const mountedRef = useMounted();
  const mutex = useRef(false);
  const render = useRenderer();

  const { project } = chapter;
  const draftRef = useRef(createDraft(chapter));
  const chapterInputRef = useRef<HTMLInputElement>();
  const volumeInputRef = useRef<HTMLInputElement>();
  const titleInputRef = useRef<HTMLInputElement>();

  const queuesRef = useRef<QueueState[]>(history.location.state?.queues || []);

  const setIsSavingRef = useRef<Dispatcher<boolean>>();
  const saveBtnOnClick = useCallback(async () => {
    if (mutex.current) return;
    setIsSavingRef.current(true);
    mutex.current = true;

    await (async () => {
      if (!draftRef.current.chapter) {
        toast.showError({ message: "Unable to save chapter", cause: "Chapter number is required" });
        return;
      }

      if (deepEqual(draftRef.current, createDraft(chapter))) {
        return;
      }

      let result: ApiResult<Chapter>;
      if (isNew) result = await CreateChapter(chapter.project.id, draftRef.current);
      else result = await UpdateChapter(chapter.id, draftRef.current);
      if (!mountedRef.current) return;

      const { response, error } = result;

      if (response) {
        Object.assign(chapter, response);

        if (isNew) {
          toast.show("Chapter has been saved as a draft.");
          window.setTimeout(() => {
            history.push(`/chapters/${response.id}`, { queues: queuesRef.current });
          }, 0);
        } else {
          toast.show("Your changes have been saved.");
        }
      }
      if (error) toast.showError(error);
    })();

    if (!mountedRef.current) return;
    setIsSavingRef.current(false);
    mutex.current = false;
  }, []);

  const addGroup = useCallback((ev: FormEvent) => {
    ev.preventDefault();

    const name: string = ev.target[0].value;
    if (!name || mutex.current) {
      return;
    }

    if (!draftRef.current.scanlationGroups.find(e => e.toLowerCase() === name.toLowerCase())) {
      draftRef.current.scanlationGroups.push(name);
      render();
    }
    ev.target[0].value = "";
  }, []);

  const removeGroup = useCallback((name: string) => {
    if (mutex.current) return;
    const idx = draftRef.current.scanlationGroups.findIndex(e => e.toLowerCase() === name.toLowerCase());
    if (idx >= 0) {
      draftRef.current.scanlationGroups.splice(idx, 1);
      render();
    }
  }, []);

  const addFile = useCallback((file: File) => {
    if (
      !mimeTypes.includes(file.type) ||
      queuesRef.current.some(f => f.fileName === file.name && f.data.size === file.size)
    ) {
      return;
    }
    queuesRef.current.push({ data: file, fileName: file.name });
  }, []);

  const pageInputRef = useRef<HTMLInputElement>();
  const pageInputOnChange = useCallback(() => {
    if (!pageInputRef.current) return;

    for (let i = 0; i < pageInputRef.current.files.length; i++) {
      addFile(pageInputRef.current.files[i]);
    }
    queuesRef.current.sort((a, b) => a.fileName.localeCompare(b.fileName, undefined, { numeric: true }));

    pageInputRef.current.value = "";
    render();
  }, []);

  const uploadBtnOnClick = useCallback(() => {
    if (!pageInputRef.current) return;
    pageInputRef.current.click();
  }, []);

  const uploadBtnOnDrop = useCallback((ev: DragEvent) => {
    ev.preventDefault();

    if (ev.dataTransfer.items) {
      for (let i = 0; i < ev.dataTransfer.items.length; i++) {
        const item = ev.dataTransfer.items[i];
        if (item.kind === "file") {
          addFile(item.getAsFile());
        }
      }
    } else {
      for (let i = 0; i < ev.dataTransfer.files.length; i++) {
        addFile(ev.dataTransfer.files[i]);
      }
    }
    queuesRef.current.sort((a, b) => a.fileName.localeCompare(b.fileName, undefined, { numeric: true }));
    render();
  }, []);

  const setIsImportingRef = useRef<Dispatcher<boolean>>();
  const importOnSubmit = useCallback(async (ev: FormEvent) => {
    ev.preventDefault();
    if (mutex.current) return;

    const input = (ev.target as HTMLFormElement).hash as HTMLInputElement;
    const hash = input.value;

    if (!hash.length) return;
    setIsImportingRef.current(true);
    mutex.current = true;
    input.value = "";

    const { response: metadata, error: e } = await GetChapterMd(hash);
    if (!mountedRef.current) return;

    if (metadata) {
      if (metadata.chapter) {
        draftRef.current.chapter = metadata.chapter;
        chapterInputRef.current.value = metadata.chapter;
      }

      if (metadata.volume) {
        draftRef.current.volume = metadata.volume;
        volumeInputRef.current.value = metadata.volume;
      }

      if (metadata.title) {
        draftRef.current.title = metadata.title;
        titleInputRef.current.value = metadata.title;
      }

      if (metadata.scanlationGroups) {
        draftRef.current.scanlationGroups.push(...metadata.scanlationGroups);
      }

      const { response, error } = await GetPagesMd(hash);
      if (!mountedRef.current) return;

      if (response) {
        queuesRef.current.push(
          ...response.pages
            .filter(fn => {
              const p = fn.split("-")[1].split(".").slice(0, -1).join(".");
              return !(
                queuesRef.current.some(q => q.fileName.includes(p)) || pagesRef.current.some(q => q.includes(p))
              );
            })
            .map(p => ({
              url: `${response.baseUrl}/data/${response.hash}/${p}`,
              fileName: p
            }))
        );
      }
      if (error) toast.showError(error);
    }
    if (e) toast.show(e);

    setIsImportingRef.current(false);
    mutex.current = false;
    render();
  }, []);

  useEffect(() => {
    if (history.location.state?.queues) {
      history.replace({ ...history.location, state: undefined });
    }
  }, []);

  useEffect(() => {
    if (isNew || !queuesRef.current?.length) {
      return;
    }

    let ok = false;
    for (let i = 0; i < queuesRef.current.length; i++) {
      const file = queuesRef.current[i];
      if (!file || file.isUploading) break;

      if (file.isUploaded) {
        queuesRef.current.splice(i, 1);
        i--;
        ok = true;
      } else if (!file.isFailed) {
        file.isUploading = true;
        ok = true;
        break;
      }
    }
    if (ok) render();
  }, [queuesRef.current?.map(e => e.isUploading || e.isFailed)]);

  return (
    <>
      {!!queuesRef.current?.length && queuesRef.current.some(q => q.isUploading) && (
        <Helmet>
          <title>Uploading Pages...</title>
        </Helmet>
      )}
      <div className="editor" id="chapter" {...props}>
        <header>
          <h2 className="title">{project.title}</h2>
          <div className="actions">
            <Link className="cancel button" to="/chapters">
              <X width="16" height="16" strokeWidth="3" />
              <strong>Cancel</strong>
            </Link>
            <button className="save button green" type="button" onClick={saveBtnOnClick}>
              <WithSpinner width="16" height="16" dispatcherRef={setIsSavingRef}>
                <Save width="16" height="16" strokeWidth="3" />
              </WithSpinner>
              <strong>Save</strong>
            </button>
          </div>
        </header>
        <section className="metadata">
          <div className="import formContainer">
            <strong>Import</strong>
            <form className="form" onSubmit={importOnSubmit}>
              <label>
                <span>mangadex.org/chapter/</span>
                <input
                  type="text"
                  name="hash"
                  placeholder="4f945a70-a8cc-44b6-92fc-985a5b3458e6"
                  autoComplete="off"
                  required
                />
              </label>
              <button type="submit">
                <WithSpinner width="16" height="16" dispatcherRef={setIsImportingRef}>
                  <Upload width="16" height="16" strokeWidth="3" />
                </WithSpinner>
              </button>
            </form>
          </div>
          <div className="chapter inputContainer">
            <strong>Chapter</strong>
            <input
              className="input"
              type="text"
              placeholder="Required"
              defaultValue={chapter.chapter}
              onChange={ev => (draftRef.current.chapter = ev.target.value)}
              ref={chapterInputRef}
            />
          </div>
          <div className="volume inputContainer">
            <strong>Volume</strong>
            <input
              className="input"
              type="text"
              placeholder="Optional"
              defaultValue={chapter.volume}
              onChange={ev => (draftRef.current.volume = ev.target.value)}
              ref={volumeInputRef}
            />
          </div>
          <div className="title inputContainer">
            <strong>Title</strong>
            <input
              className="input"
              type="text"
              placeholder="Optional"
              defaultValue={chapter.title}
              onChange={ev => (draftRef.current.title = ev.target.value)}
              ref={titleInputRef}
            />
          </div>
          <div className="groups">
            <form className="form" onSubmit={addGroup}>
              <input type="text" placeholder="Scanlation Groups" required />
              <button type="submit">
                <Plus width="16" height="16" strokeWidth="3" />
              </button>
            </form>
            {!!draftRef.current.scanlationGroups.length && (
              <div className="buttonGroups">
                {draftRef.current.scanlationGroups.map(name => (
                  <button
                    className="button"
                    type="button"
                    data-active
                    onClick={() => removeGroup(name)}
                    key={`artist-${name}`}
                  >
                    <strong>{name}</strong>
                    <X width="16" height="16" strokeWidth="3" />
                  </button>
                ))}
              </div>
            )}
          </div>
        </section>
        <section className="queues">
          <h3 className="title">Queues ({queuesRef.current.length})</h3>
          <input
            hidden
            multiple
            type="file"
            accept="image/png, image/jpeg"
            onChange={pageInputOnChange}
            ref={pageInputRef}
          />
          <button
            className="upload"
            type="button"
            onClick={uploadBtnOnClick}
            onDrop={uploadBtnOnDrop}
            onDragOver={ev => ev.preventDefault()}
          >
            <span>Drag files here or choose your files</span>
            <small>(Acceptable mime types: image/png, image/jpeg)</small>
            {!!queuesRef.current?.length && queuesRef.current.some(q => q.isUploading) && (
              <Spinner className="uploading" width="24" height="24" />
            )}
          </button>
          {!!queuesRef.current.length && (
            <div className="entries">
              {queuesRef.current.map(queue => (
                <Queue
                  queue={queue}
                  queuesRef={queuesRef}
                  chapter={chapter}
                  pagesRef={pagesRef}
                  key={`file-${queue.fileName}`}
                />
              ))}
            </div>
          )}
        </section>
        <section className="pages">
          <h3 className="title">Pages ({pagesRef.current?.length})</h3>
          <div className="entries">
            {pagesRef.current.length ? (
              pagesRef.current.map((fileName, i) => (
                <Page chapter={chapter} pagesRef={pagesRef} fileName={fileName} index={i} key={`page-${fileName}`} />
              ))
            ) : (
              <div className="empty">
                <strong>This chapter has no pages.</strong>
              </div>
            )}
          </div>
        </section>
      </div>
    </>
  );
};

export default Editor;

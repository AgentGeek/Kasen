import "easymde/dist/easymde.min.css";
import deepEqual from "fast-deep-equal";
import React, { FormEvent, useCallback, useContext, useEffect, useMemo, useRef, useState } from "react";
import { Image, Plus, RefreshCw, Save, Star, Trash, Upload, X } from "react-feather";
import { Link, useHistory } from "react-router-dom";
import {
  CheckProjectExists,
  CreateProject,
  DeleteCover,
  GetCovers,
  GetProjectMd,
  SetCover,
  UpdateProject,
  UploadCover
} from "../../../api";
import {
  Demographic,
  Permission,
  ProjectCols as ProjectCol,
  ProjectStatus,
  Rating,
  SeriesStatus
} from "../../../constants";
import { GetCoverURL, HasPerms } from "../../../utils/utils";
import { useAuthors, useMounted, useTags } from "../../Hooks";
import useMarkdown from "../../Markdown";
import { useModal } from "../../Modal";
import { createRenderer, useRenderer } from "../../Renderer";
import Spinner, { WithSpinner } from "../../Spinner";
import { useToast } from "../../Toast";
import ManageContext from "../ManageContext";

interface CoverData {
  data?: File;
  url?: string;
}

const createDraft = (data: Project): ProjectDraft => ({
  title: data.title || undefined,
  description: data.description || undefined,
  projectStatus: data.projectStatus || ProjectStatus.Ongoing,
  seriesStatus: data.seriesStatus || SeriesStatus.Ongoing,
  demographic: data.demographic || Demographic.None,
  rating: data.rating || Rating.None,
  artists: data.artists?.map(e => e.name) || [],
  authors: data.authors?.map(e => e.name) || [],
  tags: data.tags?.map(e => e.name) || []
});

const Cover = ({ data, cover, render }: { data: Project; cover: Cover; render: Renderer }) => {
  const { user } = useContext(ManageContext);
  const canDeleteCover = useMemo(() => HasPerms(user, Permission.DeleteCover), []);
  const canSetCover = useMemo(() => HasPerms(user, Permission.SetCover), []);

  const editorRender = useRenderer();
  const toast = useToast();
  const mountedRef = useMounted();
  const mutex = useRef(false);

  const setIsDeletingRef = useRef<Dispatcher<boolean>>();
  const setIsSettingCoverRef = useRef<Dispatcher<boolean>>();

  const deleteBtnOnClick = useCallback(async () => {
    if (mutex.current) return;
    setIsDeletingRef.current(true);
    mutex.current = true;

    await (async () => {
      const { error, status } = await DeleteCover(data.id, cover.id);
      if (!mountedRef.current) return;

      if (status === 204) {
        const idx = data.covers.findIndex(c => c.id === cover.id);
        if (idx >= 0) {
          data.covers.splice(idx, 1);

          render();
          toast.show("Cover has been deleted");
        }
      }
      if (error) {
        setIsDeletingRef.current(false);
        toast.showError(error);
      }
    })();

    if (!mountedRef.current) return;
    mutex.current = false;
  }, []);

  const setBtnOnClick = useCallback(async () => {
    if (mutex.current) return;
    setIsSettingCoverRef.current(true);
    mutex.current = true;

    await (async () => {
      const { error, status } = await SetCover(data.id, cover.id);
      if (!mountedRef.current) return;

      if (status === 204) {
        data.cover = cover;

        render();
        editorRender();
        toast.show("Main cover has been set");
      }
      if (error) {
        setIsSettingCoverRef.current(false);
        toast.showError(error);
      }
    })();

    if (!mountedRef.current) return;
    mutex.current = false;
  }, []);

  return (
    <div className="entry">
      <figure className="thumbnail">
        <img src={`${GetCoverURL(data, cover)}/320.jpg`} loading="lazy" />
      </figure>
      {cover.id !== data.cover?.id && (canDeleteCover || canSetCover) && (
        <div className="actions">
          {canDeleteCover && (
            <button className="delete" type="button" onClick={deleteBtnOnClick}>
              <WithSpinner width="16" height="16" dispatcherRef={setIsDeletingRef}>
                <Trash width="16" height="16" strokeWidth="3" />
              </WithSpinner>
            </button>
          )}
          {canSetCover && (
            <button className="set" type="button" onClick={setBtnOnClick}>
              <WithSpinner width="16" height="16" dispatcherRef={setIsSettingCoverRef}>
                <Star width="16" height="16" strokeWidth="3" />
              </WithSpinner>
              <strong>Set as main cover</strong>
            </button>
          )}
        </div>
      )}
    </div>
  );
};

const circumference = 376.991118;
const CoverUpload = ({ data, render }: { data: Project; render: Renderer }) => {
  const toast = useToast();
  const mountedRef = useMounted();

  const progressBarRef = useRef<SVGCircleElement>();
  const coverDataInputRef = useRef<HTMLInputElement>();
  const coverBlob = useMemo(
    () => !!coverDataInputRef.current?.files?.length && URL.createObjectURL(coverDataInputRef.current.files[0]),
    [coverDataInputRef.current?.files?.length]
  );

  const coverDataInputOnChange = useCallback(async () => {
    if (!coverDataInputRef.current.files?.length) return;
    render();

    const request = UploadCover(data.id, { data: coverDataInputRef.current.files?.[0] });
    request.upload.addEventListener("progress", ev => {
      const percentage = (ev.loaded / ev.total) * 100;
      if (!mountedRef.current && !(percentage % 100)) {
        toast.showError({ message: "Unable to upload cover", cause: "Operation was interrupted" });
        request.cancel();
        return;
      }
      const offset = ((100 - percentage) / 100) * circumference;
      progressBarRef.current?.setAttribute("stroke-dashoffset", offset.toString());
    });

    const { response, error } = await request;
    if (!mountedRef.current) return;

    if (response) {
      if (data.covers.some(c => c.id === response.id)) {
        toast.showError({ message: "Unable to upload cover", cause: "Cover already exists" });
      } else data.covers.push(response);
    }
    if (error) toast.showError(error);

    coverDataInputRef.current.value = "";
    render();
  }, [coverDataInputRef.current]);

  const uploadCoverBtnOnClick = useCallback(() => {
    if (!coverDataInputRef.current) return;
    coverDataInputRef.current.click();
  }, []);

  return (
    <div className="entry">
      {coverBlob && (
        <figure className="thumbnail">
          <img src={coverBlob} />
        </figure>
      )}
      <input
        type="file"
        accept="image/png, image/jpeg"
        onChange={coverDataInputOnChange}
        ref={coverDataInputRef}
        hidden
      />
      {coverBlob ? (
        <div className="progress">
          <svg strokeWidth="8" viewBox="0 0 150 150">
            <circle className="background" fill="transparent" cx="75" cy="75" r="60" />
            <circle
              fill="transparent"
              stroke="currentColor"
              strokeDasharray={circumference}
              strokeDashoffset={circumference}
              cx="75"
              cy="75"
              r="60"
              ref={progressBarRef}
            />
          </svg>
        </div>
      ) : (
        <button className="upload" type="button" onClick={uploadCoverBtnOnClick}>
          <Upload width="48" height="48" strokeWidth="1.5" />
        </button>
      )}
    </div>
  );
};

const Covers = ({ data }: { data: Project }) => {
  const { user } = useContext(ManageContext);
  const canUploadCover = useMemo(() => HasPerms(user, Permission.UploadCover), []);

  const toast = useToast();
  const mountedRef = useMounted();

  const [isLoading, setIsLoading] = useState(!(data.covers ??= []).length);
  const render = createRenderer();

  useEffect(() => {
    if (data.covers.length) return;
    GetCovers(data.id).then(({ response, error }) => {
      if (!mountedRef.current) return;

      if (response) data.covers = response || [];
      if (error) toast.showError(error);
      setIsLoading(false);
    });
  }, []);

  return isLoading ? (
    <Spinner className="loading" width="60" height="60" strokeWidth="8" />
  ) : (
    <div className="entries">
      {canUploadCover && <CoverUpload data={data} render={render} />}
      {Array.from(data.covers)
        .reverse()
        .map(cover => (
          <Cover data={data} cover={cover} render={render} key={`cover-${cover.id}`} />
        ))}
    </div>
  );
};

interface EditorProps extends Props<HTMLDivElement> {
  initialData: Project;
  isNew?: boolean;
}

const Editor = ({ initialData: project, isNew, ...props }: EditorProps) => {
  const history = useHistory();
  const modal = useModal();
  const toast = useToast();

  const { id, slug } = project;
  const draftRef = useRef<ProjectDraft>(createDraft(project));

  const titleRef = useRef<HTMLHeadingElement>();
  const { description } = draftRef.current;

  const { projectStatus, seriesStatus, demographic } = draftRef.current;
  const { rating, artists, authors, tags } = draftRef.current;

  const { markdown, Markdown } = useMarkdown({
    placeholder: "Series description",
    initialValue: description,
    toolbar: false
  });

  const [resetKey, reset] = useState(0);
  const render = useRenderer();

  const mountedRef = useMounted();
  const mutex = useRef(false);

  const [tagList, isLoadingTags] = useTags();
  const [authorList, isLoadingAuthors] = useAuthors();

  const coverDataRef = useRef<CoverData>();
  const coverDataInputRef = useRef<HTMLInputElement>();
  const coverBlob = useMemo(
    () => coverDataRef.current?.data && URL.createObjectURL(coverDataRef.current.data),
    [coverDataRef.current?.data]
  );

  const resetBtnOnClick = useCallback(() => {
    if (mutex.current) return;
    mutex.current = true;

    draftRef.current.title = titleRef.current.textContent.trim() || undefined;
    draftRef.current.description = markdown.current.value() || undefined;

    (() => {
      const newDraft = createDraft(project);
      if (deepEqual(draftRef.current, newDraft) && !coverDataRef.current) {
        return;
      }

      draftRef.current = newDraft;
      coverDataRef.current = undefined;

      reset(i => ++i);

      window.requestAnimationFrame(() => {
        titleRef.current.textContent = project.title;
        markdown.current.value(project.description || "");
      });

      toast.show("Changes have been discarded.");
    })();

    mutex.current = false;
  }, []);

  const setIsSavingRef = useRef<Dispatcher<boolean>>();
  const saveBtnOnClick = useCallback(async () => {
    if (mutex.current) return;
    setIsSavingRef.current(true);
    mutex.current = true;

    await (async () => {
      draftRef.current.title = titleRef.current.textContent;
      draftRef.current.description = markdown.current.value();

      if (!draftRef.current.title) {
        toast.showError({ message: "Unable to save project", cause: "Project title is required" });
        return;
      }

      if (isNew && !coverDataRef.current) {
        toast.showError({ message: "Unable to save project", cause: "Cover is required" });
        return;
      }

      if (deepEqual(draftRef.current, createDraft(project))) {
        return;
      }

      if (draftRef.current.title !== project.title) {
        const { response, error } = await CheckProjectExists({ title: draftRef.current.title });
        if (!mountedRef.current) return;
        if (response?.id) {
          toast.show({
            type: "error",
            content: (
              <p>
                Unable to save project: Title is already in use -{" "}
                <b>
                  <Link to={`/projects/${response.id}/edit`}>
                    {draftRef.current.title} (#{response.id})
                  </Link>
                </b>
              </p>
            )
          });
        }
        if (error) toast.showError(error);
        if (response?.id || error) return;
      }

      let result: ApiResult<Project>;
      if (isNew) result = await CreateProject(draftRef.current);
      else result = await UpdateProject(id, draftRef.current);
      if (!mountedRef.current) return;

      const { response, error: e } = result;

      if (response) {
        Object.assign(project, {
          ...response,
          artists: response.artists,
          authors: response.authors,
          tags: response.tags
        });
        if (isNew) {
          const request = UploadCover(response.id, { ...coverDataRef.current, isInitialCover: true });
          request.upload.addEventListener("progress", ev => {
            if (mountedRef.current) return;
            request.cancel();

            toast.show({
              type: "error",
              content: <p>Unable to upload cover: Operation was interrupted.</p>
            });
          });

          const { error: er } = await request;
          toast.show(
            <p>
              Project &apos;<strong>{response.title}</strong>&apos; has been saved as a draft.
            </p>
          );
          if (er) toast.showError(er);
          if (!mountedRef.current) return;

          window.setTimeout(() => history.push(`/projects/${response.id}/edit`), 0);
        } else {
          toast.show("Your changes have been saved");
        }
        resetBtnOnClick();
      }
      if (e) toast.showError(e);
    })();

    if (!mountedRef.current) return;
    setIsSavingRef.current(false);
    mutex.current = false;
  }, []);

  const coverDataInputOnChange = useCallback(() => {
    if (!coverDataInputRef.current.files?.length || mutex.current) {
      return;
    }
    coverDataRef.current = { data: coverDataInputRef.current.files[0] };
    render();
  }, []);

  const uploadCoverBtnOnClick = useCallback(() => {
    if (!coverDataInputRef.current || mutex.current) {
      return;
    }
    coverDataInputRef.current.click();
  }, []);

  const showCoversBtnOnClick = useCallback(() => {
    if (mutex.current) return;
    modal.show({
      className: "covers",
      title: "Covers",
      content: <Covers data={project} />
    });
  }, []);

  const setColumn = (k: ProjectCol, v: string) => {
    if (mutex.current) return;
    if (draftRef.current[k] !== v) {
      draftRef.current[k] = v;
      render();
    }
  };

  const addArtist = useCallback((ev: FormEvent) => {
    ev.preventDefault();

    const name: string = ev.target[0].value;
    if (!name || mutex.current) {
      return;
    }

    if (!draftRef.current.artists.find(e => e.toLowerCase() === name.toLowerCase())) {
      draftRef.current.artists.push(name);
      render();
    }
    ev.target[0].value = "";
  }, []);

  const removeArtist = useCallback((name: string) => {
    if (mutex.current) return;
    const idx = draftRef.current.artists.findIndex(e => e.toLowerCase() === name.toLowerCase());
    if (idx >= 0) {
      draftRef.current.artists.splice(idx, 1);
      render();
    }
  }, []);

  const addAuthor = useCallback((ev: FormEvent) => {
    ev.preventDefault();

    const name: string = ev.target[0].value;
    if (!name || mutex.current) {
      return;
    }

    if (!draftRef.current.authors.some(e => e.toLowerCase() === name.toLowerCase())) {
      draftRef.current.authors.push(name);
      render();
    }
    ev.target[0].value = "";
  }, []);

  const removeAuthor = useCallback((name: string) => {
    if (mutex.current) return;
    const idx = draftRef.current.authors.findIndex(e => e.toLowerCase() === name.toLowerCase());
    if (idx >= 0) {
      draftRef.current.authors.splice(idx, 1);
      render();
    }
  }, []);

  const selectTag = useCallback((name: string) => {
    if (mutex.current) return;
    const idx = draftRef.current.tags.findIndex(e => e.toLowerCase() === name.toLowerCase());
    if (idx >= 0) draftRef.current.tags.splice(idx, 1);
    else draftRef.current.tags.push(name);
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

    const { response: metadata, error: e } = await GetProjectMd(hash);
    if (!mountedRef.current) return;

    if (metadata) {
      Object.assign(draftRef.current, metadata);

      if (metadata.title) {
        titleRef.current.textContent = metadata.title;
      }
      markdown.current.value(metadata.description);
      coverDataRef.current = { url: metadata.coverUrl };
    }
    if (e) toast.show(e);

    setIsImportingRef.current(false);
    mutex.current = false;
    render();
  }, []);

  useEffect(() => {
    titleRef.current.addEventListener("keydown", ev => ev.key === "Enter" && ev.preventDefault());
    titleRef.current.addEventListener("paste", ev => {
      ev.preventDefault();
      ev.stopPropagation();

      titleRef.current.textContent = ev.clipboardData.getData("Text").trim();
    });
    titleRef.current.addEventListener("drop", ev => ev.preventDefault());
    titleRef.current.textContent = draftRef.current.title;
  }, []);

  return (
    <div className="editor" id="project" {...props}>
      <div className="main">
        <div className="import formContainer">
          <strong>Import</strong>
          <form className="form" onSubmit={importOnSubmit}>
            <label>
              <span>mangadex.org/manga/</span>
              <input
                type="text"
                name="hash"
                placeholder="18ef1e92-f574-4c8c-99af-555e44ba418e"
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
        <h1
          className="title"
          placeholder="Series title"
          spellCheck="false"
          contentEditable
          ref={titleRef}
          key={resetKey}
        />
        <Markdown />
        <section className="metadata">
          <div className="artists">
            <form className="form" onSubmit={addArtist}>
              <input type="text" placeholder="Artists" required key={resetKey} />
              <button type="submit">
                <Plus width="16" height="16" strokeWidth="3" />
              </button>
            </form>
            {!!artists.length && (
              <div className="buttonGroups">
                {artists.map(name => (
                  <button
                    className="button"
                    type="button"
                    data-active
                    onClick={() => removeArtist(name)}
                    key={`artist-${name}`}
                  >
                    <strong>{name}</strong>
                    <X width="16" height="16" strokeWidth="3" />
                  </button>
                ))}
              </div>
            )}
          </div>
          <div className="authors">
            <form className="form" onSubmit={addAuthor}>
              <input type="text" placeholder="Authors" required key={resetKey} />
              <button type="submit">
                <Plus width="16" height="16" strokeWidth="3" />
              </button>
            </form>
            {!!authors.length && (
              <div className="buttonGroups">
                {authors.map(name => (
                  <button
                    className="button"
                    type="button"
                    data-active
                    onClick={() => removeAuthor(name)}
                    key={`author-${name}`}
                  >
                    <strong>{name}</strong>
                    <X width="16" height="16" strokeWidth="3" />
                  </button>
                ))}
              </div>
            )}
          </div>
          {isLoadingTags ? (
            <Spinner className="loading" width="60" height="60" strokeWidth="8" />
          ) : (
            !!tagList.length && (
              <div className="tags">
                <b>Tags</b>
                <div className="buttonGroups">
                  {tagList.map(tag => (
                    <button
                      className="button"
                      type="button"
                      data-active={tags.includes(tag.name) || undefined}
                      onClick={() => selectTag(tag.name)}
                      key={`tag-${tag.slug}`}
                    >
                      <strong>{tag.name}</strong>
                    </button>
                  ))}
                </div>
              </div>
            )
          )}
        </section>
      </div>
      <div className="sidebar">
        <div className="wrapper">
          <div className="actions">
            <div>
              <Link className="cancel button" to="/projects">
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
            <button className="reset button red" type="button" onClick={resetBtnOnClick}>
              <RefreshCw width="16" height="16" strokeWidth="3" />
              <strong>Reset</strong>
            </button>
          </div>
          <div id="cover">
            <figure className="cover">
              {coverDataRef.current ? (
                <img src={coverDataRef.current.url || coverBlob} />
              ) : (
                project.cover && <img src={GetCoverURL(project, project.cover)} />
              )}
            </figure>
            <div className="actions">
              {isNew ? (
                <>
                  <input
                    type="file"
                    accept="image/png, image/jpeg"
                    onChange={coverDataInputOnChange}
                    ref={coverDataInputRef}
                    hidden
                  />
                  <button className="uploadCover button blue" type="button" onClick={uploadCoverBtnOnClick}>
                    <Upload width="16" height="16" strokeWidth="3" />
                    <strong>Upload Cover</strong>
                  </button>
                </>
              ) : (
                <button className="showCovers button blue" type="button" onClick={showCoversBtnOnClick}>
                  <Image width="16" height="16" strokeWidth="3" />
                  <strong>Covers</strong>
                </button>
              )}
            </div>
          </div>
        </div>
        <section className="metadata">
          <div className="projectStatus">
            <b>Project Status</b>
            <div className="buttonGroups">
              {Object.keys(ProjectStatus).map(k => (
                <button
                  className="button"
                  type="button"
                  data-active={projectStatus === ProjectStatus[k] || undefined}
                  onClick={() => setColumn(ProjectCol.ProjectStatus, ProjectStatus[k])}
                  key={`projectStatus-${k}`}
                >
                  <strong>{k}</strong>
                </button>
              ))}
            </div>
          </div>
          <div className="seriesStatus">
            <b>Series Status</b>
            <div className="buttonGroups">
              {Object.keys(SeriesStatus).map(k => (
                <button
                  className="button"
                  type="button"
                  data-active={seriesStatus === SeriesStatus[k] || undefined}
                  onClick={() => setColumn(ProjectCol.SeriesStatus, SeriesStatus[k])}
                  key={`seriesStatus-${k}`}
                >
                  <strong>{k}</strong>
                </button>
              ))}
            </div>
          </div>
          <div className="demographic">
            <b>Demographic</b>
            <div className="buttonGroups">
              {Object.keys(Demographic).map(k => (
                <button
                  className="button"
                  type="button"
                  data-active={demographic === Demographic[k] || undefined}
                  onClick={() => setColumn(ProjectCol.Demographic, Demographic[k])}
                  key={`demographic-${k}`}
                >
                  <strong>{k}</strong>
                </button>
              ))}
            </div>
          </div>
          <div className="contentRating">
            <b>Content Rating</b>
            <div className="buttonGroups">
              {Object.keys(Rating).map(k => (
                <button
                  className="button"
                  type="button"
                  data-active={rating === Rating[k] || undefined}
                  onClick={() => setColumn(ProjectCol.Rating, Rating[k])}
                  key={`rating-${k}`}
                >
                  <strong>{k}</strong>
                </button>
              ))}
            </div>
          </div>
        </section>
      </div>
    </div>
  );
};

export default Editor;

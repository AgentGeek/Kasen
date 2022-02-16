import React, { memo, useContext, useRef } from "react";
import { Check, Edit, ExternalLink, Eye, EyeOff, Lock, Trash, Unlock, Upload, X } from "react-feather";
import { Link, useHistory } from "react-router-dom";
import { DeleteProject, LockProject, PublishProject, UnlockProject, UnpublishProject } from "../../../api";
import { Permission } from "../../../constants";
import { FormatUnix, GetCoverURL, HasPerms } from "../../../utils/utils";
import { useMounted } from "../../Hooks";
import { WithIntersectionObserver } from "../../IntersectionObserver";
import { useModal } from "../../Modal";
import { WithSpinner } from "../../Spinner";
import { useToast } from "../../Toast";
import ManageContext from "../ManageContext";

interface ModalProps {
  // eslint-disable-next-line react/no-unused-prop-types
  entriesRef: Mutable<Project[]>;
  isDeletedRef: Mutable<boolean>;
  project: Project;

  // eslint-disable-next-line react/no-unused-prop-types
  render: Renderer;
}

const PublishStateModal = ({ entriesRef, isDeletedRef, project, render }: ModalProps) => {
  const modal = useModal();
  const toast = useToast();

  const mountedRef = useMounted();
  const mutex = useRef(false);

  const setIsChangingPublishStateRef = useRef<Dispatcher<boolean>>();
  const changePublishStateCallback = async () => {
    if (!mountedRef.current || mutex.current || project.locked || isDeletedRef.current) {
      return;
    }
    setIsChangingPublishStateRef.current(true);
    mutex.current = true;

    let result: ApiResult<Project>;
    if (project.publishedAt) result = await UnpublishProject(project.id);
    else result = await PublishProject(project.id);
    if (!mountedRef.current) return;

    const { response, error } = result;
    if (response) {
      const idx = entriesRef.current.findIndex(e => e.id === project.id);
      if (idx >= 0) {
        entriesRef.current[idx] = { ...entriesRef.current[idx], publishedAt: response.publishedAt };
        Object.assign(entriesRef.current[idx], response);
        render();
      }

      toast.show(
        <>
          Project &apos;<b>{project.title}</b>&apos; has been {project.publishedAt ? "unpublished" : "published"}.
        </>
      );
    }

    if (error) toast.showError(error);
    setIsChangingPublishStateRef.current(false);
    mutex.current = false;
    modal.hide();
  };

  return (
    <div className="prompt">
      <p>
        Are you sure you want to {project.publishedAt ? "unpublish" : "publish"} &apos;<b>{project.title}</b>&apos;?
      </p>
      {project.publishedAt && (
        <small>Chapters will still be visible to the public. This action can be undone anytime.</small>
      )}
      <div className="actions">
        <button className="cancel" type="button" onClick={modal.hide}>
          <X width="16" height="16" strokeWidth="3" />
          <strong>Cancel</strong>
        </button>
        <button className="confirm" type="button" onClick={changePublishStateCallback}>
          <WithSpinner width="16" height="16" dispatcherRef={setIsChangingPublishStateRef}>
            <Check width="16" height="16" strokeWidth="3" />
          </WithSpinner>
          <strong>Confirm</strong>
        </button>
      </div>
    </div>
  );
};

const LockStateModal = ({ entriesRef, isDeletedRef, project, render }: ModalProps) => {
  const modal = useModal();
  const toast = useToast();

  const mountedRef = useMounted();
  const mutex = useRef(false);

  const setIsChangingLockStateRef = useRef<Dispatcher<boolean>>();
  const changeLockStateCallback = async () => {
    if (!mountedRef.current || mutex.current || isDeletedRef.current) {
      return;
    }
    setIsChangingLockStateRef.current(true);
    mutex.current = true;

    let result: ApiResult<Project>;
    if (project.locked) result = await UnlockProject(project.id);
    else result = await LockProject(project.id);
    if (!mountedRef.current) return;

    const { response, error } = result;
    if (response) {
      const idx = entriesRef.current.findIndex(e => e.id === project.id);
      if (idx >= 0) {
        entriesRef.current[idx] = { ...entriesRef.current[idx], locked: response.locked };
        Object.assign(entriesRef.current[idx], response);
        render();
      }

      toast.show(
        <>
          Project &apos;<b>{project.title}</b>&apos; has been {project.locked ? "unlocked" : "locked"}.
        </>
      );
    }
    if (error) toast.showError(error);

    setIsChangingLockStateRef.current(false);
    mutex.current = false;
    modal.hide();
  };

  return (
    <div className="prompt">
      <p>
        Are you sure you want to {project.locked ? "unlock" : "lock"} &apos;<b>{project.title}</b>&apos;?
      </p>

      {!project.locked && (
        <small>
          This project and its chapters will still be visible to the public, but will no longer be editable and
          deletable. This action can be undone anytime.
        </small>
      )}

      <div className="actions">
        <button className="cancel" type="button" onClick={modal.hide}>
          <X width="16" height="16" strokeWidth="3" />
          <strong>Cancel</strong>
        </button>

        <button className="confirm" type="button" onClick={changeLockStateCallback}>
          <WithSpinner width="16" height="16" dispatcherRef={setIsChangingLockStateRef}>
            <Check width="16" height="16" strokeWidth="3" />
          </WithSpinner>
          <strong>Confirm</strong>
        </button>
      </div>
    </div>
  );
};

const DeleteModal = ({ isDeletedRef, project, render }: ModalProps) => {
  const history = useHistory();
  const modal = useModal();
  const toast = useToast();

  const mountedRef = useMounted();
  const mutex = useRef(false);

  const setIsDeletingRef = useRef<Dispatcher<boolean>>();
  const deleteProjectCallback = async () => {
    if (!mountedRef.current || mutex.current || project.locked || isDeletedRef.current) {
      return;
    }
    setIsDeletingRef.current(true);
    mutex.current = true;

    const { error, status } = await DeleteProject(project.id);
    if (!mountedRef.current) return;

    if (error) {
      toast.showError(error);
    } else if (status === 204) {
      toast.show(
        <>
          Project &apos;<b>{project.title}</b>&apos; has been deleted.
        </>
      );
      isDeletedRef.current = true;
      history.replace({ ...history.location, state: { deleted: new Date() } });
    }

    setIsDeletingRef.current(false);
    mutex.current = false;
    modal.hide();
    render();
  };

  return (
    <div className="prompt">
      <p>
        Are you sure you want to delete &apos;<b>{project.title}</b>&apos;?
      </p>
      <small>
        This project will be removed permanently, including its covers and chapters. This action cannot be undone.
      </small>
      <div className="actions">
        <button className="cancel" type="button" onClick={modal.hide}>
          <X width="16" height="16" strokeWidth="3" />
          <strong>Cancel</strong>
        </button>
        <button className="confirm" type="button" onClick={deleteProjectCallback}>
          <WithSpinner width="16" height="16" dispatcherRef={setIsDeletingRef}>
            <Check width="16" height="16" strokeWidth="3" />
          </WithSpinner>
          <strong>Confirm</strong>
        </button>
      </div>
    </div>
  );
};

export const Entry = ({
  project,
  entriesRef,
  render
}: {
  project: Project;
  entriesRef: Mutable<Project[]>;
  render: Renderer;
}) => {
  const modal = useModal();

  const { user } = useContext(ManageContext);
  const { id, slug, title, locked, cover, artists, authors, createdAt, updatedAt, publishedAt } = project;

  const isDeletedRef = useRef(false);
  const modalProps = { entriesRef, isDeletedRef, project, render };

  const changePublishStateModal = () => {
    if (locked || isDeletedRef.current) {
      return;
    }
    modal.show({
      title: `${publishedAt ? "Unpublish" : "Publish"} Project`,
      content: <PublishStateModal {...modalProps} />
    });
  };

  const changeLockStateModal = () => {
    if (isDeletedRef.current) return;
    modal.show({
      title: `${locked ? "Unlock" : "Lock"} Project`,
      content: <LockStateModal {...modalProps} />
    });
  };

  const deleteProjectModal = () => {
    if (locked || isDeletedRef.current) {
      return;
    }
    modal.show({
      title: "Delete Project",
      content: <DeleteModal {...modalProps} />
    });
  };

  return (
    <WithIntersectionObserver
      className="entry"
      data-published={publishedAt > 0 || undefined}
      data-locked={locked || undefined}
      data-deleted={isDeletedRef.current || undefined}
      once
    >
      <div>
        {cover && (
          <figure className="cover">
            <Link to={`/chapters?project=${id}`}>
              <img src={`${GetCoverURL(project, cover)}/320.jpg`} loading="lazy" />
            </Link>
          </figure>
        )}
      </div>
      <div className="metadata">
        <h3 className="title">
          {!publishedAt && <span className="status">Draft</span>}
          {locked && <span className="status">Locked</span>}
          <Link to={`/chapters?project=${id}`}>{title}</Link>
        </h3>
        <div className="meta-line-1">
          <span className="createdAt">
            <strong>{"Created: "}</strong>
            <span>{FormatUnix(createdAt)}</span>
          </span>
          <span className="separator">&#xB7;</span>
          <span className="updatedAt">
            <strong>{"Updated: "}</strong>
            <span>{FormatUnix(updatedAt || createdAt)}</span>
          </span>
          {publishedAt && (
            <>
              <span className="separator">&#xB7;</span>
              <span className="publishedAt">
                <strong>{"Published: "}</strong>
                <span>{FormatUnix(publishedAt)}</span>
              </span>
            </>
          )}
        </div>
        <div className="meta-line-2">
          <span className="artists">
            <strong>{"Artists: "}</strong>
            {artists?.length ? (
              artists.map((x, i) => (
                <span key={`artist-${x.name}`}>
                  {!!i && ", "}
                  <Link to={`/projects?artist=${x.name}`}>{x.name}</Link>
                </span>
              ))
            ) : (
              <span>N/A</span>
            )}
          </span>
          <span className="separator">&#xB7;</span>
          <span className="authors">
            <strong>{"Authors: "}</strong>
            {authors?.length ? (
              authors.map((x, i) => (
                <span key={`author-${x.name}`}>
                  {!!i && ", "}
                  <Link to={`/projects?author=${x.name}`}>{x.name}</Link>
                </span>
              ))
            ) : (
              <span>N/A</span>
            )}
          </span>
        </div>
        <div className="actions">
          <a
            className="view"
            href={`/projects/${id}/${slug}`}
            target="_blank"
            rel="noreferrer"
            tabIndex={!publishedAt || isDeletedRef.current ? -1 : undefined}
          >
            <ExternalLink width="16" height="16" strokeWidth="3" />
            <strong>View</strong>
          </a>
          {HasPerms(
            user,
            Permission.EditProject,
            Permission.UploadCover,
            Permission.SetCover,
            Permission.DeleteCover
          ) && (
            <Link className="edit" to={`/projects/${id}/edit`} tabIndex={isDeletedRef.current ? -1 : undefined}>
              <Edit width="16" height="16" strokeWidth="3" />
              <strong>Edit</strong>
            </Link>
          )}
          {HasPerms(user, Permission.CreateChapter) && (
            <Link className="newChapter" to={`/chapters/new?project=${id}`}>
              <Upload width="16" height="16" strokeWidth="3" />
              <strong>New Chapter</strong>
            </Link>
          )}
          {publishedAt
            ? HasPerms(user, Permission.UnpublishProject) && (
                <button
                  className="publishState"
                  type="button"
                  onClick={changePublishStateModal}
                  tabIndex={locked || isDeletedRef.current ? -1 : undefined}
                >
                  <EyeOff width="16" height="16" strokeWidth="3" />
                  <strong>Unpublish</strong>
                </button>
              )
            : HasPerms(user, Permission.PublishProject) && (
                <button
                  className="publishState"
                  type="button"
                  onClick={changePublishStateModal}
                  tabIndex={locked || isDeletedRef.current ? -1 : undefined}
                >
                  <Eye width="16" height="16" strokeWidth="3" />
                  <strong>Publish</strong>
                </button>
              )}
          {locked
            ? HasPerms(user, Permission.UnlockProject) && (
                <button
                  className="lockState"
                  type="button"
                  onClick={changeLockStateModal}
                  tabIndex={isDeletedRef.current ? -1 : undefined}
                >
                  <Unlock width="16" height="16" strokeWidth="3" />
                  <strong>Unlock</strong>
                </button>
              )
            : HasPerms(user, Permission.LockProject) && (
                <button
                  className="lockState"
                  type="button"
                  onClick={changeLockStateModal}
                  tabIndex={isDeletedRef.current ? -1 : undefined}
                >
                  <Lock width="16" height="16" strokeWidth="3" />
                  <strong>Lock</strong>
                </button>
              )}
          {HasPerms(user, Permission.DeleteProject) && (
            <button
              className="delete"
              type="button"
              onClick={deleteProjectModal}
              tabIndex={locked || isDeletedRef.current ? -1 : undefined}
            >
              <Trash width="16" height="16" strokeWidth="3" />
              <strong>Delete</strong>
            </button>
          )}
        </div>
      </div>
    </WithIntersectionObserver>
  );
};

export const MemoizedEntry = memo(
  Entry,
  (prev, next) =>
    prev.project.publishedAt === next.project.publishedAt &&
    prev.project.updatedAt === next.project.updatedAt &&
    prev.project.locked === next.project.locked
);

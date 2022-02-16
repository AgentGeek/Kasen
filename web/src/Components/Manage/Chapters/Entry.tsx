import React, { memo, useContext, useRef } from "react";
import { Book, Check, Edit, ExternalLink, Eye, EyeOff, Lock, Trash, Unlock, X } from "react-feather";
import { Link, useHistory } from "react-router-dom";
import { DeleteChapter, LockChapter, PublishChapter, UnlockChapter, UnpublishChapter } from "../../../api";
import { Permission } from "../../../constants";
import { FormatChapter, FormatUnix, HasPerms } from "../../../utils/utils";
import { useMounted } from "../../Hooks";
import { WithIntersectionObserver } from "../../IntersectionObserver";
import { useModal } from "../../Modal";
import { WithSpinner } from "../../Spinner";
import { useToast } from "../../Toast";
import ManageContext from "../ManageContext";

interface ModalProps {
  // eslint-disable-next-line react/no-unused-prop-types
  entriesRef: Mutable<Chapter[]>;
  isDeletedRef: Mutable<boolean>;
  chapter: Chapter;

  // eslint-disable-next-line react/no-unused-prop-types
  render: Renderer;
}

const PublishStateModal = ({ entriesRef, isDeletedRef, chapter, render }: ModalProps) => {
  const modal = useModal();
  const toast = useToast();

  const mountedRef = useMounted();
  const mutex = useRef(false);

  const setIsChangingPublishStateRef = useRef<Dispatcher<boolean>>();
  const changePublishStateCallback = async () => {
    if (!mountedRef.current || mutex.current || chapter.locked || isDeletedRef.current) {
      return;
    }
    setIsChangingPublishStateRef.current(true);
    mutex.current = true;

    let result: ApiResult<Chapter>;
    if (chapter.publishedAt) result = await UnpublishChapter(chapter.id);
    else result = await PublishChapter(chapter.id);
    if (!mountedRef.current) return;

    const { response, error } = result;
    if (response) {
      const idx = entriesRef.current.findIndex(e => e.id === chapter.id);
      if (idx >= 0) {
        entriesRef.current[idx] = { ...entriesRef.current[idx], publishedAt: response.publishedAt };
        Object.assign(entriesRef.current[idx], response);
        render();
      }

      toast.show(
        <>
          &apos;<b>{FormatChapter(chapter)}</b>&apos; of project &apos;<b>{chapter.project.title}</b>&apos; has been{" "}
          {chapter.publishedAt ? "unpublished" : "published"}.
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
        Are you sure you want to {chapter.publishedAt ? "unpublish" : "publish"} &apos;<b>{FormatChapter(chapter)}</b>
        &apos; of project &apos;<b>{chapter.project.title}</b>&apos;?
      </p>
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

const LockStateModal = ({ entriesRef, isDeletedRef, chapter, render }: ModalProps) => {
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

    let result: ApiResult<Chapter>;
    if (chapter.locked) result = await UnlockChapter(chapter.id);
    else result = await LockChapter(chapter.id);
    if (!mountedRef.current) return;

    const { response, error } = result;
    if (response) {
      const idx = entriesRef.current.findIndex(e => e.id === chapter.id);
      if (idx >= 0) {
        entriesRef.current[idx] = { ...entriesRef.current[idx], locked: response.locked };
        Object.assign(entriesRef.current[idx], response);
        render();
      }

      toast.show(
        <>
          &apos;<b>{FormatChapter(chapter)}</b>&apos; of project &apos;<b>{chapter.project.title}</b>&apos; has been{" "}
          {chapter.locked ? "unlocked" : "locked"}.
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
        Are you sure you want to {chapter.locked ? "unlock" : "lock"} &apos;<b>{FormatChapter(chapter)}</b>
        &apos; of project &apos;<b>{chapter.project.title}</b>&apos;?
      </p>

      {!chapter.locked && (
        <small>
          This chapter will still be visible to the public, but will no longer be editable and deletable. This action
          can be undone anytime.
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

const DeleteModal = ({ isDeletedRef, chapter }: ModalProps) => {
  const history = useHistory();
  const modal = useModal();
  const toast = useToast();

  const mountedRef = useMounted();
  const mutex = useRef(false);

  const setIsDeletingRef = useRef<Dispatcher<boolean>>();
  const deleteChapterCallback = async () => {
    if (!mountedRef.current || mutex.current || chapter.locked || isDeletedRef.current) {
      return;
    }
    setIsDeletingRef.current(true);
    mutex.current = true;

    const { error, status } = await DeleteChapter(chapter.id);
    if (!mountedRef.current) return;

    if (error) {
      toast.showError(error);
    } else if (status === 204) {
      toast.show(
        <>
          Chapter &apos;<b>{FormatChapter(chapter)}</b>
          &apos; of project &apos;<b>{chapter.project.title}</b>&apos; has been deleted.
        </>
      );
      isDeletedRef.current = true;
      history.replace({ ...history.location, state: { deleted: new Date() } });
    }

    setIsDeletingRef.current(false);
    mutex.current = false;
    modal.hide();
  };

  return (
    <div className="prompt">
      <p>
        Are you sure you want to delete &apos;<b>{FormatChapter(chapter)}</b>
        &apos; of project &apos;<b>{chapter.project.title}</b>&apos;?
      </p>
      <small>This chapter will be removed permanently. This action cannot be undone.</small>
      <div className="actions">
        <button className="cancel" type="button" onClick={modal.hide}>
          <X width="16" height="16" strokeWidth="3" />
          <strong>Cancel</strong>
        </button>
        <button className="confirm" type="button" onClick={deleteChapterCallback}>
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
  chapter,
  entriesRef,
  render
}: {
  chapter: Chapter;
  entriesRef: Mutable<Chapter[]>;
  render: Renderer;
}) => {
  const modal = useModal();

  const { user } = useContext(ManageContext);
  const { project, uploader, scanlationGroups } = chapter;
  const { id, locked, createdAt, updatedAt, publishedAt } = chapter;

  const isDeletedRef = useRef(false);
  const modalProps = { entriesRef, isDeletedRef, chapter, render };

  const changePublishStateModal = () => {
    if (locked || isDeletedRef.current) {
      return;
    }
    modal.show({
      title: `${publishedAt ? "Unpublish" : "Publish"} Chapter`,
      content: <PublishStateModal {...modalProps} />
    });
  };

  const changeLockStateModal = () => {
    if (isDeletedRef.current) return;
    modal.show({
      title: `${locked ? "Unlock" : "Lock"} Chapter`,
      content: <LockStateModal {...modalProps} />
    });
  };

  const deleteChapterModal = () => {
    if (locked || isDeletedRef.current) {
      return;
    }
    modal.show({
      title: "Delete Chapter",
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
      <div className="metadata">
        <div className="projectTitle">
          <Link to={`/chapters?project=${project.id}`}>
            <Book width="14" height="14" strokeWidth="3" />
            <span>{project.title}</span>
          </Link>
        </div>
        <h3 className="title">
          {!publishedAt && <span className="status">Draft</span>}
          {locked && <span className="status">Locked</span>}
          {FormatChapter(chapter)}
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
          <span className="uploader">
            <strong>{"Uploader: "}</strong>
            <span>
              <Link to={`/chapters?uploader=${uploader.name}`}>{uploader.name}</Link>
            </span>
          </span>
          {scanlationGroups && (
            <>
              <span className="separator">&#xB7;</span>
              <span className="scanlationGroups">
                <strong>{"Groups: "}</strong>
                {scanlationGroups.map((group, i) => (
                  <span key={`group-${group.id}`}>
                    {!!i && ", "}
                    <Link to={`/chapters?scanlation_group=${group.slug}`}>{group.name}</Link>
                  </span>
                ))}
              </span>
            </>
          )}
        </div>
        <div className="actions">
          <a
            className="view"
            href={`/chapters/${id}`}
            target="_blank"
            rel="noreferrer"
            tabIndex={!publishedAt || isDeletedRef.current ? -1 : undefined}
          >
            <ExternalLink width="16" height="16" strokeWidth="3" />
            <strong>View</strong>
          </a>
          {(HasPerms(user, Permission.EditChapters) ||
            (uploader?.id === user.id && HasPerms(user, Permission.EditChapter))) && (
            <Link className="edit" to={`/chapters/${id}`} tabIndex={isDeletedRef.current ? -1 : undefined}>
              <Edit width="16" height="16" strokeWidth="3" />
              <strong>Edit</strong>
            </Link>
          )}
          {publishedAt
            ? (HasPerms(user, Permission.UnpublishChapters) ||
                (uploader?.id === user.id && HasPerms(user, Permission.UnpublishChapter))) && (
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
            : (HasPerms(user, Permission.PublishChapters) ||
                (uploader?.id === user.id && HasPerms(user, Permission.PublishChapter))) && (
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
            ? (HasPerms(user, Permission.UnlockChapters) ||
                (uploader?.id === user.id && HasPerms(user, Permission.UnlockChapter))) && (
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
            : (HasPerms(user, Permission.LockChapters) ||
                (uploader?.id === user.id && HasPerms(user, Permission.LockChapter))) && (
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
          {(HasPerms(user, Permission.DeleteChapters) ||
            (uploader?.id === user.id && HasPerms(user, Permission.DeleteChapter))) && (
            <button
              className="delete"
              type="button"
              onClick={deleteChapterModal}
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
    prev.chapter.publishedAt === next.chapter.publishedAt &&
    prev.chapter.updatedAt === next.chapter.updatedAt &&
    prev.chapter.locked === next.chapter.locked
);

import React, { useContext, useEffect, useRef } from "react";
import { Helmet } from "react-helmet";
import { useHistory, useParams } from "react-router";
import { ChapterPreloads, GetChapter, GetPages } from "../../../api";
import { Permission } from "../../../constants";
import { FormatChapter, HasPerms } from "../../../utils/utils";
import { useMounted, usePermissions } from "../../Hooks";
import { createRenderer, WithRenderer } from "../../Renderer";
import Spinner from "../../Spinner";
import { useToast } from "../../Toast";
import ManageContext from "../ManageContext";
import Editor from "./Editor";

const Edit = () => {
  usePermissions([Permission.EditChapter, Permission.EditChapters]);

  const { title, user } = useContext(ManageContext);
  const { id } = useParams<{ id: string }>();
  const history = useHistory();
  const toast = useToast();

  const mountedRef = useMounted();
  const isLoadingRef = useRef(true);
  const render = createRenderer();

  const chapterRef = useRef({} as Chapter);
  const pagesRef = useRef<string[]>([]);

  useEffect(() => {
    const chapterId = Number.parseInt(id, 10);
    if (Number.isNaN(chapterId) || chapterId <= 0) {
      toast.showError({ message: "Failed to get chapter", cause: "Chapter id is invalid" });
      history.replace("/chapters");
      return;
    }

    GetChapter(chapterId, {
      preloads: [ChapterPreloads.Project, ChapterPreloads.Uploader, ChapterPreloads.ScanlationGroups],
      includesDrafts: true
    }).then(async ({ response, error }) => {
      if (!mountedRef.current) return;

      if (response) chapterRef.current = response;
      if (error) {
        toast.showError(error);
        if (error.cause.toLowerCase().includes("does not exists")) {
          history.replace("/chapters");
          return;
        }
      }

      const result = await GetPages(chapterId);
      if (!mountedRef.current) return;

      if (result.response) {
        pagesRef.current = result.response;
        if (!HasPerms(user, Permission.EditChapters)) {
          if (response.uploader.id !== user.id || !HasPerms(user, Permission.EditChapter)) {
            toast.showError({
              message: "Failed to get chapter",
              cause: "Not enough privileges"
            });
            history.replace("/chapters");
            return;
          }
        }
      }
      if (result.error) toast.showError(result.error);

      isLoadingRef.current = false;
      render();
    });
  }, []);

  return (
    <>
      <Helmet>
        {chapterRef.current?.project ? (
          <title>
            Edit: {FormatChapter(chapterRef.current)} | {chapterRef.current.project.title} - {title}
          </title>
        ) : (
          <title>Edit Chapter - {title}</title>
        )}
      </Helmet>
      <main>
        {isLoadingRef.current ? (
          <Spinner className="loading" width="120" height="120" strokeWidth="8" />
        ) : (
          <WithRenderer>
            <Editor chapter={chapterRef.current} pagesRef={pagesRef} />
          </WithRenderer>
        )}
      </main>
    </>
  );
};

export default Edit;

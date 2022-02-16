import React, { useContext, useEffect, useRef } from "react";
import { Helmet } from "react-helmet";
import { useHistory } from "react-router";
import { GetProject } from "../../../api";
import { Permission } from "../../../constants";
import { useMounted, usePermissions } from "../../Hooks";
import { createRenderer, WithRenderer } from "../../Renderer";
import Spinner from "../../Spinner";
import { useToast } from "../../Toast";
import ManageContext from "../ManageContext";
import Editor from "./Editor";

const New = () => {
  usePermissions([Permission.CreateChapter]);

  const { title } = useContext(ManageContext);
  const history = useHistory();
  const toast = useToast();

  const mountedRef = useMounted();
  const isLoadingRef = useRef(true);
  const render = createRenderer();

  const chapterRef = useRef({} as Chapter);
  const pagesRef = useRef<string[]>([]);

  useEffect(() => {
    const projectId = Number.parseInt(new URLSearchParams(history.location.search).get("project"), 10);
    if (Number.isNaN(projectId) || projectId <= 0) {
      toast.showError({ message: "Failed to get project", cause: "Project id is invalid" });
      history.replace("/chapters");
      return;
    }

    GetProject(projectId, { includesDrafts: true }).then(({ response, error }) => {
      if (!mountedRef.current) return;

      if (response) chapterRef.current.project = response;
      if (error) {
        toast.showError(error);
        if (error.cause.toLowerCase().includes("does not exists")) {
          history.replace("/chapters");
          return;
        }
      }

      isLoadingRef.current = false;
      render();
    });
  }, []);

  return (
    <>
      <Helmet>
        <title>New Chapter - {title}</title>
      </Helmet>
      <main>
        {isLoadingRef.current ? (
          <Spinner className="loading" width="120" height="120" strokeWidth="8" />
        ) : (
          <WithRenderer>
            <Editor chapter={chapterRef.current} pagesRef={pagesRef} isNew />
          </WithRenderer>
        )}
      </main>
    </>
  );
};

export default New;

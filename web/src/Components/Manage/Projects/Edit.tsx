import React, { useContext, useEffect, useState } from "react";
import { Helmet } from "react-helmet";
import { useHistory, useParams } from "react-router";
import { GetProject, ProjectPreloads } from "../../../api";
import { Permission } from "../../../constants";
import { usePermissions } from "../../Hooks";
import { WithModal } from "../../Modal";
import { WithRenderer } from "../../Renderer";
import Spinner from "../../Spinner";
import { useToast } from "../../Toast";
import ManageContext from "../ManageContext";
import Editor from "./Editor";

const Edit = () => {
  usePermissions([Permission.EditProject, Permission.UploadCover, Permission.SetCover, Permission.DeleteCover]);

  const params = useParams<{ id: string }>();
  const history = useHistory();

  const { title } = useContext(ManageContext);
  const toast = useToast();

  const [data, setData] = useState({} as Project);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const projectId = Number.parseInt(params?.id, 10);
    if (Number.isNaN(projectId) || projectId <= 0) {
      toast.showError({
        message: "Unable to get project",
        cause: "Project id is invalid"
      });
      history.replace("/projects");
      return;
    }

    GetProject(projectId, {
      preloads: [ProjectPreloads.Artists, ProjectPreloads.Authors, ProjectPreloads.Cover, ProjectPreloads.Tags],
      includesDrafts: true
    }).then(({ response, error }) => {
      if (response) setData(response);
      if (error) {
        toast.showError(error);
        if (error.cause.toLowerCase().includes("does not exists")) {
          history.replace("/projects");
          return;
        }
      }
      setIsLoading(false);
    });
  }, []);

  return (
    <>
      <Helmet>
        {data?.title ? (
          <title>
            Edit: {data.title} - {title}
          </title>
        ) : (
          <title>Edit - {title}</title>
        )}
      </Helmet>
      <main>
        {isLoading ? (
          <Spinner className="loading" width="120" height="120" strokeWidth="8" />
        ) : (
          <WithRenderer>
            <WithModal>
              <Editor initialData={data} />
            </WithModal>
          </WithRenderer>
        )}
      </main>
    </>
  );
};

export default Edit;

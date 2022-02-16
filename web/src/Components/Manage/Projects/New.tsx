import React, { useContext, useRef } from "react";
import { Helmet } from "react-helmet";
import { Permission } from "../../../constants";
import { usePermissions } from "../../Hooks";
import { WithModal } from "../../Modal";
import { WithRenderer } from "../../Renderer";
import ManageContext from "../ManageContext";
import Editor from "./Editor";

const New = () => {
  usePermissions([Permission.CreateProject]);
  const { title } = useContext(ManageContext);

  const projectRef = useRef({} as Project);

  return (
    <>
      <Helmet>
        <title>New Project - {title}</title>
      </Helmet>
      <main>
        <WithRenderer>
          <WithModal>
            <Editor initialData={projectRef.current} isNew />
          </WithModal>
        </WithRenderer>
      </main>
    </>
  );
};

export default New;

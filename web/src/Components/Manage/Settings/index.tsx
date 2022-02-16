import React, { useContext } from "react";
import { Helmet } from "react-helmet";
import { Permission } from "../../../constants";
import { HasPerms } from "../../../utils/utils";
import ManageContext from "../ManageContext";
import Meta from "./Meta";
import Service from "./Service";
import User from "./User";

const Settings = () => {
  const { title, user } = useContext(ManageContext);

  return (
    <>
      <Helmet>
        <title>Settings - {title}</title>
      </Helmet>
      <main className="settings">
        {HasPerms(user, Permission.Manage) && (
          <>
            <Meta />
            <Service />
          </>
        )}
        {HasPerms(user, Permission.EditUser, Permission.DeleteUser) && <User />}
      </main>
    </>
  );
};

export default Settings;

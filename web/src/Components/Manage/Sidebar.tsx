import React, { useContext, useMemo } from "react";
import { NavLink } from "react-router-dom";
import routes from "../../routes";
import { HasPerms } from "../../utils/utils";
import ManageContext from "./ManageContext";

const Sidebar = () => {
  const { user } = useContext(ManageContext);
  const links = useMemo(
    () =>
      Object.values(routes).filter(({ permissions }) => (permissions?.length ? HasPerms(user, ...permissions) : true)),
    []
  );
  return (
    <aside id="sidebar">
      <nav aria-label="Primary">
        <ul>
          {links.map(({ path, name }) => (
            <li key={`link-${name}`}>
              <NavLink isActive={(_, { pathname, search }) => `${pathname}${search}`.startsWith(path)} to={path}>
                {name}
              </NavLink>
            </li>
          ))}
        </ul>
      </nav>
      <div className="actions">
        <a href="/">Home</a>
        <a className="logout" href="/logout">
          Logout
        </a>
      </div>
    </aside>
  );
};

export default Sidebar;

import React, { useEffect, useMemo } from "react";
import { BrowserRouter as Router, Redirect, Route, Switch, useHistory } from "react-router-dom";
import { WithToast } from "../Toast";
import Chapters from "./Chapters";
import EditChapter from "./Chapters/Edit";
import NewChapter from "./Chapters/New";
import ManageContext from "./ManageContext";
import Projects from "./Projects";
import EditProject from "./Projects/Edit";
import NewProject from "./Projects/New";
import Settings from "./Settings";
import Sidebar from "./Sidebar";

const ResetScrollPosition = () => {
  const history = useHistory();
  useEffect(() => {
    history.listen(() => {
      window.scrollTo(0, 0);
    });
  }, []);

  return null;
};

const Manage = ({ data }: { data: ViewData }) => {
  const context = useMemo<ViewData>(() => data, []);
  return (
    <ManageContext.Provider value={context}>
      <Router basename="/manage">
        <ResetScrollPosition />
        <WithToast>
          <Sidebar />
          <Switch>
            <Route exact path="/projects" component={Projects} />
            <Route exact path="/projects/new" component={NewProject} />
            <Route exact path="/projects/:id/edit" component={EditProject} />

            <Route exact path="/chapters" component={Chapters} />
            <Route exact path="/chapters/new" component={NewChapter} />
            <Route exact path="/chapters/:id" component={EditChapter} />

            <Route exact path="/settings" component={Settings} />

            <Redirect from="/project*" to="/projects" />
            <Redirect from="/chapter*" to="/chapters" />
            <Redirect to="/projects" />
          </Switch>
        </WithToast>
      </Router>
    </ManageContext.Provider>
  );
};

export default Manage;

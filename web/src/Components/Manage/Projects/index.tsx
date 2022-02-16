import React, { FormEvent, useCallback, useContext, useEffect, useMemo, useRef } from "react";
import { ChevronsLeft, ChevronsRight, Plus, Search } from "react-feather";
import Helmet from "react-helmet";
import { Link, useHistory } from "react-router-dom";
import { GetProjects, Order, ProjectPreloads, ProjectSort, ProjectSortKeys } from "../../../api";
import { Limit, Permission } from "../../../constants";
import routes from "../../../routes";
import { CreatePagination, HasPerms } from "../../../utils/utils";
import { useMutableMemo, usePermissions } from "../../Hooks";
import { WithModal } from "../../Modal";
import NavLinkWrapper from "../../NavLinkWrapper";
import { createRenderer, WithRenderer } from "../../Renderer";
import Spinner from "../../Spinner";
import { useToast } from "../../Toast";
import ManageContext from "../ManageContext";
import { MemoizedEntry } from "./Entry";

const Projects = () => {
  const { user } = usePermissions(routes.projects.permissions);
  const { title } = useContext(ManageContext);
  const history = useHistory<{ deleted: Date }>();

  const render = createRenderer();
  const toast = useToast();

  const isLoadingRef = useRef(true);
  const entriesRef = useRef<Project[]>([]);
  const totalEntriesRef = useRef(0);
  const totalPagesRef = useRef(0);

  const queryRef = useMutableMemo(() => {
    const searchParams = new URLSearchParams(history.location.search);
    return {
      title: searchParams.get("title") || undefined,
      artists: searchParams.getAll("artist") || undefined,
      authors: searchParams.getAll("author") || undefined,
      currentPage: Number.parseInt(searchParams.get("page") || "1", 10),
      sort: (searchParams.get("sort") || ProjectSort.UpdatedAt) as ProjectSort,
      order: (searchParams.get("order") || Order.DESC) as Order
    };
  }, [history.location.search]);

  const pagination = useMemo(
    () => CreatePagination(queryRef.current.currentPage, totalPagesRef.current),
    [queryRef.current.currentPage, totalPagesRef.current]
  );

  const toSearchParams = useCallback((v: number) => {
    const searchParams = new URLSearchParams(history.location.search);
    if (v <= 1) searchParams.delete("page");
    else searchParams.set("page", v.toString());

    const str = searchParams.toString();
    if (str) return `${history.location.pathname}?${str}`;
    return history.location.pathname;
  }, []);

  const changeTitle = useCallback((ev: FormEvent) => {
    ev.preventDefault();
    ev.stopPropagation();

    const titleInput = (ev.target as any).title as HTMLInputElement;
    if (!titleInput || titleInput.value === queryRef.current.title || (!titleInput.value && !queryRef.current.title))
      return;

    const searchParams = new URLSearchParams(history.location.search);
    if (titleInput.value) searchParams.set("title", titleInput.value);
    else searchParams.delete("title");

    searchParams.delete("page");
    isLoadingRef.current = true;
    history.push({ search: searchParams.toString() });
  }, []);

  const changeSort = useCallback((v: ProjectSort) => {
    const searchParams = new URLSearchParams(history.location.search);
    if (v.toLowerCase() === ProjectSort.UpdatedAt) searchParams.delete("sort");
    else searchParams.set("sort", v);

    searchParams.delete("page");
    isLoadingRef.current = true;
    history.push({ search: searchParams.toString() });
  }, []);

  const changeOrder = useCallback(() => {
    const searchParams = new URLSearchParams(history.location.search);
    if (queryRef.current.order.toLowerCase() === Order.ASC) searchParams.delete("order");
    else searchParams.set("order", Order.ASC);

    searchParams.delete("page");
    isLoadingRef.current = true;
    history.push({ search: searchParams.toString() });
  }, []);

  useEffect(() => {
    if (!isLoadingRef.current) {
      isLoadingRef.current = true;
      render();
    }
    GetProjects({
      title: queryRef.current.title,
      artists: queryRef.current.artists,
      authors: queryRef.current.authors,
      limit: Limit,
      offset: Limit * (queryRef.current.currentPage - 1),
      preloads: [ProjectPreloads.Cover, ProjectPreloads.Artists, ProjectPreloads.Authors],
      sort: queryRef.current.sort,
      order: queryRef.current.order,
      includesDrafts: true
    }).then(({ response, error }) => {
      if (response) {
        entriesRef.current = response.data;
        totalEntriesRef.current = response.total;
        totalPagesRef.current = Math.ceil(response.total / Limit);
      } else if (error) {
        toast.showError(error);
      }

      isLoadingRef.current = false;
      render();
    });
  }, [history.location.search, history.location.state?.deleted]);

  return (
    <>
      <Helmet>
        <title>Manage: Project - {title}</title>
      </Helmet>
      <main className="feed" id="projects">
        {isLoadingRef.current ? (
          <Spinner className="loading" width="120" height="120" strokeWidth="8" />
        ) : (
          <>
            <header>
              <h2 className="title">Projects ({totalEntriesRef.current})</h2>
              {HasPerms(user, Permission.CreateProject) && (
                <Link className="new button blue" to="/projects/new">
                  <Plus width="16" height="16" strokeWidth="3" />
                  <strong>New Project</strong>
                </Link>
              )}
            </header>

            <nav className="filters" aria-label="Filters">
              <form className="flex" onSubmit={changeTitle}>
                <div className="sort">
                  <select
                    defaultValue={queryRef.current.sort}
                    onChange={ev => changeSort(ev.target.value as ProjectSort)}
                  >
                    {ProjectSortKeys.map(v => (
                      <option value={ProjectSort[v]} key={`sort-${ProjectSort[v]}`}>
                        {v}
                      </option>
                    ))}
                  </select>
                </div>

                <button className="order" type="button" onClick={changeOrder}>
                  {queryRef.current.order.toUpperCase()}{" "}
                </button>

                <div className="search">
                  <input name="title" type="text" placeholder="Search" defaultValue={queryRef.current.title} />
                  <button type="submit">
                    <Search width="16" height="16" strokeWidth="3" />
                  </button>
                </div>
              </form>
            </nav>

            <div className="entries" data-empty={!entriesRef.current?.length || undefined}>
              {entriesRef.current?.length ? (
                <WithRenderer>
                  <WithModal>
                    {entriesRef.current.map(p => (
                      <MemoizedEntry project={p} entriesRef={entriesRef} render={render} key={`project-${p.id}`} />
                    ))}
                  </WithModal>
                </WithRenderer>
              ) : (
                <div className="empty">
                  {queryRef.current.title ? (
                    <>
                      <h3>No results found</h3>
                      <p>There are no results that match your search.</p>
                    </>
                  ) : (
                    <>
                      <h3>Empty in projects</h3>
                      <p>Create a new project and it will show up here.</p>
                    </>
                  )}
                </div>
              )}
            </div>

            {!!totalEntriesRef.current && totalPagesRef.current > 1 && (
              <nav className="pagination" arial-label="Page">
                <ul>
                  <li className="first">
                    <NavLinkWrapper exact to={toSearchParams(1)} title="First page">
                      <ChevronsLeft width="16" height="16" strokeWidth="3" />
                    </NavLinkWrapper>
                  </li>

                  {pagination.map(n => (
                    <li key={`page-${n}`}>
                      <NavLinkWrapper exact to={toSearchParams(n)}>
                        <strong>{n}</strong>
                      </NavLinkWrapper>
                    </li>
                  ))}

                  <li className="last">
                    <NavLinkWrapper exact to={toSearchParams(totalPagesRef.current)} title="Last page">
                      <ChevronsRight width="16" height="16" strokeWidth="3" />
                    </NavLinkWrapper>
                  </li>
                </ul>
              </nav>
            )}
          </>
        )}
      </main>
    </>
  );
};

export default Projects;

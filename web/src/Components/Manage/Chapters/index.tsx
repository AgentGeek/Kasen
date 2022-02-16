import React, { useCallback, useContext, useEffect, useMemo, useRef } from "react";
import { ChevronsLeft, ChevronsRight, Plus } from "react-feather";
import { Helmet } from "react-helmet";
import { useHistory } from "react-router";
import { Link } from "react-router-dom";
import { ChapterPreloads, ChapterSort, ChapterSortKeys, GetChapters, GetChaptersByProject, Order } from "../../../api";
import { Limit, Permission } from "../../../constants";
import routes from "../../../routes";
import { CreatePagination, HasPerms } from "../../../utils/utils";
import { useMutableMemo, usePermissions } from "../../Hooks";
import { WithModal } from "../../Modal";
import NavLinkWrapper from "../../NavLinkWrapper";
import { createRenderer } from "../../Renderer";
import Spinner from "../../Spinner";
import { useToast } from "../../Toast";
import ManageContext from "../ManageContext";
import { MemoizedEntry } from "./Entry";

const Chapters = () => {
  const { user } = usePermissions(routes.chapters.permissions);
  const history = useHistory<{ deleted: Date }>();

  const { title } = useContext(ManageContext);
  const toast = useToast();
  const render = createRenderer();

  const isLoadingRef = useRef(true);

  const entriesRef = useRef<Chapter[]>([]);
  const totalEntriesRef = useRef(0);
  const totalPagesRef = useRef(0);

  const queryRef = useMutableMemo(() => {
    const searchParams = new URLSearchParams(history.location.search);
    return {
      projectId: Number.parseInt(searchParams.get("project") || "0", 10),
      uploader: searchParams.get("uploader") || undefined,
      scanlationGroups: searchParams.getAll("scanlation_group") || undefined,
      currentPage: Number.parseInt(searchParams.get("page") || "1", 10),
      sort: (searchParams.get("sort") || ChapterSort.UpdatedAt) as ChapterSort,
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

  const changeSort = useCallback((v: ChapterSort) => {
    const searchParams = new URLSearchParams(history.location.search);
    if (v.toLowerCase() === ChapterSort.UpdatedAt) searchParams.delete("sort");
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

    const o = {
      uploader: queryRef.current.uploader,
      scanlationGroups: queryRef.current.scanlationGroups,
      limit: Limit,
      offset: Limit * (queryRef.current.currentPage - 1),
      preloads: [ChapterPreloads.Project, ChapterPreloads.Uploader, ChapterPreloads.ScanlationGroups],
      sort: queryRef.current.sort,
      order: queryRef.current.order,
      includesDrafts: true
    };

    let result: Api<{ data?: Chapter[]; total?: number }>;
    if (queryRef.current.projectId > 0) {
      result = GetChaptersByProject(queryRef.current.projectId, o);
    } else result = GetChapters(o);

    result.then(({ response, error }) => {
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
    <WithModal>
      <Helmet>
        <title>Manage: Chapters - {title}</title>
      </Helmet>
      <main className="feed" id="chapters">
        {isLoadingRef.current ? (
          <Spinner className="loading" width="120" height="120" strokeWidth="8" />
        ) : (
          <>
            <header>
              <h2 className="title">Chapters ({totalEntriesRef.current})</h2>
              {queryRef.current.projectId > 0 && HasPerms(user, Permission.CreateChapter) && (
                <Link className="new button blue" to={`/chapters/new?project=${queryRef.current.projectId}`}>
                  <Plus width="16" height="16" strokeWidth="3" />
                  <strong>New Chapter</strong>
                </Link>
              )}
            </header>

            <nav className="filters" aria-label="Filters">
              <div className="flex">
                <div className="sort">
                  <select
                    defaultValue={queryRef.current.sort}
                    onChange={ev => changeSort(ev.target.value as ChapterSort)}
                  >
                    {ChapterSortKeys.map(v => (
                      <option value={ChapterSort[v]} key={`sort-${ChapterSort[v]}`}>
                        {v}
                      </option>
                    ))}
                  </select>
                </div>
                <button className="order" type="button" onClick={changeOrder}>
                  {queryRef.current.order.toUpperCase()}{" "}
                </button>
              </div>
            </nav>

            <section className="entries" data-empty={!entriesRef.current?.length || undefined}>
              {entriesRef.current?.length ? (
                entriesRef.current.map(c => (
                  <MemoizedEntry chapter={c} entriesRef={entriesRef} render={render} key={`chapter-${c.id}`} />
                ))
              ) : (
                <div className="empty">
                  <h3>{queryRef.current.projectId > 0 ? "This project has no chapters" : "Empty in chapters"}</h3>
                  <p>Create a new chapter and it will show up here.</p>
                </div>
              )}
            </section>

            {!!totalEntriesRef.current && totalPagesRef.current > 1 && (
              <nav className="pagination" aria-label="Page">
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
    </WithModal>
  );
};

export default Chapters;

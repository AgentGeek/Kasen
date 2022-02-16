import React, { useCallback, useContext, useEffect, useRef } from "react";
import { ChevronDown, ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, Settings, X } from "react-feather";
import { useKeydown, useModal, useNavigate } from "../Hooks";
import { PageDirection, PageScale, SidebarPosition, SidebarPositionKeys } from "./constants";
import ReaderContext from "./ReaderContext";

const directions: [string, PageDirection][] = [
  ["Top-to-bottom", PageDirection.TopToBottom],
  ["Right-to-left", PageDirection.RightToLeft],
  ["Left-to-right", PageDirection.LeftToRight]
];

const scales: [string, string, PageScale][] = [
  ["Default", "Keep original sizes, fit to container width if larger", PageScale.Default],
  ["Original", "Keep original sizes", PageScale.Original],
  ["Width", "Fit to max-width if larger", PageScale.Width],
  ["Height", "Fit to max-height if larger", PageScale.Height],
  ["Stretch", "Stretch to max-width", PageScale.Stretch],
  ["Fit Width", "Fit to screen width if larger", PageScale.FitWidth],
  ["Fit Height", "Fit to screen height if larger", PageScale.FitHeight],
  ["Stretch Width", "Stretch to screen width", PageScale.StretchWidth],
  ["Stretch Height", "Stretch to screen height", PageScale.StretchHeight]
];

const noop = () => {};

const formatChapter = (c: Chapter, short?: boolean) => {
  let str = "";
  if (c.volume) {
    str = `Vol. ${c.volume} `;
  }

  str += `Ch. ${c.chapter}`;
  if (c.title && !short) {
    str += ` - ${c.title}`;
  }
  return str;
};

const Sidebar = () => {
  const { pref, chapter, chapters, pagination, pagesRef, currentPageRef, render } = useContext(ReaderContext);
  const { project } = chapter;

  const { sidebarPosition, showSidebar } = pref;
  const isReverse = pref.direction === PageDirection.RightToLeft;
  const { first, prev, next, last, jump } = useNavigate(render);

  const popRef = useRef<HTMLDivElement>();
  const isPopRef = useRef(false);
  const timeoutRef = useRef(0);

  const activeItemRef = useRef<HTMLLIElement>();

  const changePreference = useCallback(
    (ev: React.FormEvent, isNumber?: boolean) => {
      const target = ev.target as HTMLInputElement;
      if (!target.value) return;
      clearTimeout(timeoutRef.current);
      timeoutRef.current = window.setTimeout(() => {
        const { type, name, checked } = target;
        if (target.type === "number") {
          const n = Number(target.value);
          if (target.min && n < Number(target.min)) {
            target.value = target.min;
          } else if (target.max && n > Number(target.max)) {
            target.value = target.max;
          }
        }

        if (type === "checkbox") {
          pref[name] = checked;
        } else {
          pref[name] = isNumber ? Number(target.value) : target.value;
        }

        localStorage.setItem("pref", JSON.stringify(pref));
        render();
      }, 250);
    },
    [pref]
  );

  const changeKeybind = useCallback(
    (ev: React.KeyboardEvent) => {
      clearTimeout(timeoutRef.current);

      timeoutRef.current = window.setTimeout(() => {
        const { name } = ev.target as HTMLInputElement;
        pref.keybinds[name] = ev.code;

        localStorage.setItem("pref", JSON.stringify(pref));
        render();
      }, 250);
    },
    [pref]
  );

  useKeydown(ev => {
    if (isPopRef.current) {
      return;
    }
    switch (ev.code) {
      case pref.keybinds.previousPage:
        prev();
        break;
      case pref.keybinds.nextPage:
        next();
        break;
      case pref.keybinds.previousChapter:
        if (!pagination.Previous) break;
        window.location.pathname = `/chapters/${pagination.Previous.id}`;
        break;
      case pref.keybinds.nextChapter:
        if (!pagination.Next) break;
        window.location.pathname = `/chapters/${pagination.Next.id}`;
        break;
      default:
        break;
    }
  });

  useModal(popRef, isPopRef, render);

  useEffect(() => {
    if (activeItemRef.current) {
      activeItemRef.current.scrollIntoView({
        block: "center"
      });
    }
  }, []);

  return (
    <aside data-position={SidebarPosition[sidebarPosition]} data-hidden={!showSidebar || undefined}>
      <header>
        <h1 className="title">
          <a href={`/projects/${project.id}/${project.slug}`} title={project.title}>
            <span>{project.title}</span>
          </a>
        </h1>
        <nav arial-label="Chapter">
          <ul>
            <li className={pagination.Previous ? undefined : "disabled"}>
              {pagination.Previous ? (
                <a href={`/chapters/${pagination.Previous.id}`}>
                  <span>Previous</span>
                </a>
              ) : (
                <span>Previous</span>
              )}
            </li>
            <li>
              <button
                type="button"
                onClick={() => {
                  isPopRef.current = !isPopRef.current;
                  render();
                }}
              >
                <Settings width="16" height="16" strokeWidth="2" />
              </button>
            </li>
            <li className={pagination.Next ? undefined : "disabled"}>
              {pagination.Next ? (
                <a href={`/chapters/${pagination.Next.id}`}>
                  <span>Next</span>
                </a>
              ) : (
                <span>Next</span>
              )}
            </li>
          </ul>
        </nav>
      </header>
      <div className="body">
        <nav className="chapters" aria-label="Chapters">
          <ul>
            {chapters.map(c => {
              const isActive = c.id === chapter.id;
              return (
                <li className={isActive ? "active" : undefined} ref={isActive ? activeItemRef : undefined} key={c.id}>
                  <a href={`/chapters/${c.id}`}>
                    <span>{formatChapter(c)}</span>
                  </a>
                </li>
              );
            })}
          </ul>
        </nav>
        {isPopRef.current && (
          <div className="settings">
            <div className="wrapper" ref={popRef}>
              <header>
                <h2>Reader Settings</h2>
                <button
                  className="close"
                  type="button"
                  title="Close"
                  onClick={() => {
                    isPopRef.current = false;
                    render();
                  }}
                >
                  <X width="16" height="16" strokeWidth="3" />
                </button>
              </header>
              <div>
                <h3>Sidebar</h3>
                <div className="select">
                  <strong>Position</strong>
                  <div>
                    <select
                      name="sidebarPosition"
                      onChange={e => changePreference(e, true)}
                      defaultValue={pref.sidebarPosition}
                    >
                      {SidebarPositionKeys.filter(x => !Number.isInteger(Number(x))).map(k => (
                        <option key={k} value={SidebarPosition[k]}>
                          {k}
                        </option>
                      ))}
                    </select>
                    <ChevronDown width="14" height="14" strokeWidth="3" />
                  </div>
                </div>

                <h3>Page</h3>
                <div className="checkbox">
                  <strong>Navigate on click</strong>
                  <div>
                    <input
                      type="checkbox"
                      name="navigateOnClick"
                      defaultChecked={pref.navigateOnClick}
                      onChange={changePreference}
                    />
                  </div>
                </div>
                <div className="select">
                  <strong>Direction</strong>
                  <div>
                    <select name="direction" onChange={e => changePreference(e, true)} defaultValue={pref.direction}>
                      {directions.map(([text, v]) => (
                        <option key={PageDirection[v]} value={v}>
                          {text}
                        </option>
                      ))}
                    </select>
                    <ChevronDown width="14" height="14" strokeWidth="3" />
                  </div>
                </div>
                <div className="select">
                  <strong>Scale</strong>
                  <div>
                    <select name="scale" onChange={e => changePreference(e, true)} defaultValue={pref.scale}>
                      {scales.map(([text, desc, v]) => (
                        <option key={PageScale[v]} value={v}>
                          {text}
                          {desc && ` (${desc})`}
                        </option>
                      ))}
                    </select>
                    <ChevronDown width="14" height="14" strokeWidth="3" />
                  </div>
                </div>
                <div className="group">
                  <div className="input">
                    <strong>Max. width</strong>
                    <div>
                      <input
                        type="number"
                        name="maxWidth"
                        defaultValue={pref.maxWidth}
                        min={0}
                        onChange={changePreference}
                      />
                      <span className="unit">px</span>
                    </div>
                  </div>
                  <div className="input">
                    <strong>Max. height</strong>
                    <div>
                      <input
                        type="number"
                        name="maxHeight"
                        defaultValue={pref.maxHeight}
                        min={0}
                        onChange={changePreference}
                      />
                      <span className="unit">px</span>
                    </div>
                  </div>
                </div>
                <div className="group">
                  <div className="input">
                    <strong>Gaps</strong>
                    <div>
                      <input type="number" name="gaps" defaultValue={pref.gaps} min={10} onChange={changePreference} />
                      <span className="unit">px</span>
                    </div>
                  </div>
                  <div className="input">
                    <strong>Zoom</strong>
                    <div>
                      <input
                        type="number"
                        name="zoom"
                        defaultValue={pref.zoom}
                        step={0.1}
                        min={0.1}
                        max={2.0}
                        onChange={changePreference}
                      />
                    </div>
                  </div>
                </div>

                <h3>Download</h3>
                <div className="group">
                  <div className="input">
                    <strong>Max. preloads</strong>
                    <div>
                      <input
                        type="number"
                        name="maxPreloads"
                        min={1}
                        defaultValue={pref.maxPreloads}
                        onChange={e => changePreference(e, true)}
                      />
                    </div>
                  </div>
                  <div className="input">
                    <strong>Max. parallels</strong>
                    <div>
                      <input
                        type="number"
                        name="maxParallel"
                        defaultValue={pref.maxParallel}
                        min={1}
                        onChange={e => changePreference(e, true)}
                      />
                    </div>
                  </div>
                </div>

                <h3>Keyboard shortcuts</h3>
                <div className="group">
                  <div className="input">
                    <strong>Previous page</strong>
                    <div>
                      <input
                        name="previousPage"
                        value={pref.keybinds.previousPage}
                        onChange={noop}
                        onKeyDown={changeKeybind}
                      />
                    </div>
                  </div>
                  <div className="input">
                    <strong>Next page</strong>
                    <div>
                      <input name="nextPage" value={pref.keybinds.nextPage} onChange={noop} onKeyDown={changeKeybind} />
                    </div>
                  </div>
                </div>
                <div className="group">
                  <div className="input">
                    <strong>Previous chapter</strong>
                    <div>
                      <input
                        name="previousChapter"
                        value={pref.keybinds.previousChapter}
                        onChange={noop}
                        onKeyDown={changeKeybind}
                      />
                    </div>
                  </div>
                  <div className="input">
                    <strong>Next chapter</strong>
                    <div>
                      <input
                        name="nextChapter"
                        value={pref.keybinds.nextChapter}
                        onChange={noop}
                        onKeyDown={changeKeybind}
                      />
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
      <footer>
        <nav aria-label="Page">
          <button type="button" title={isReverse ? "Last Page" : "First Page"} onClick={isReverse ? last : first}>
            <ChevronsLeft width="16" height="16" strokeWidth="3" />
          </button>
          <button type="button" title={isReverse ? "Next Page" : "Previous Page"} onClick={isReverse ? next : prev}>
            <ChevronLeft width="16" height="16" strokeWidth="3" />
          </button>
          <div className="count">
            {currentPageRef.current && pagesRef.current?.length && (
              <select defaultValue={currentPageRef.current.index + 1} onChange={e => jump(Number(e.target.value))}>
                {pagesRef.current.map(p => (
                  <option value={p.index + 1} key={p.index}>
                    {p.index + 1}
                  </option>
                ))}
              </select>
            )}
            {Number(currentPageRef.current?.index) + 1 || "?"}/{pagesRef.current.length || "?"}
          </div>
          <button type="button" title={isReverse ? "Previous Page" : "Next Page"} onClick={isReverse ? prev : next}>
            <ChevronRight width="16" height="16" strokeWidth="3" />
          </button>
          <button type="button" title={isReverse ? "First Page" : "Last Page"} onClick={isReverse ? first : last}>
            <ChevronsRight width="16" height="16" strokeWidth="3" />
          </button>
        </nav>
      </footer>
    </aside>
  );
};

export default Sidebar;

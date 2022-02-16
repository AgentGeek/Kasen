CREATE TABLE IF NOT EXISTS menu (
  id       BIGSERIAL PRIMARY KEY
);

ALTER TABLE menu
  ADD IF NOT EXISTS name VARCHAR(32) NOT NULL DEFAULT NULL,
  ADD IF NOT EXISTS url VARCHAR(255) NOT NULL DEFAULT NULL,
  ADD IF NOT EXISTS priority SMALLINT NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS author (
  id BIGSERIAL PRIMARY KEY
);

ALTER TABLE author
  ADD IF NOT EXISTS slug VARCHAR(255) NOT NULL DEFAULT NULL,
  ADD IF NOT EXISTS name VARCHAR(128) NOT NULL DEFAULT NULL,
  ALTER COLUMN name TYPE VARCHAR(128);

CREATE UNIQUE INDEX IF NOT EXISTS author_slug_uindex ON author(slug);
CREATE UNIQUE INDEX IF NOT EXISTS author_name_uindex ON author(name);

CREATE TABLE IF NOT EXISTS scanlation_group (
  id BIGSERIAL PRIMARY KEY
);

ALTER TABLE scanlation_group
  ADD IF NOT EXISTS slug VARCHAR(255) NOT NULL DEFAULT NULL,
  ADD IF NOT EXISTS name VARCHAR(128) NOT NULL DEFAULT NULL,
  ALTER COLUMN name TYPE VARCHAR(128);

CREATE UNIQUE INDEX IF NOT EXISTS scanlation_group_slug_uindex ON scanlation_group(slug);
CREATE UNIQUE INDEX IF NOT EXISTS scanlation_group_name_uindex ON scanlation_group(name);

CREATE TABLE IF NOT EXISTS tag (
  id   BIGSERIAL PRIMARY KEY
);

ALTER TABLE tag
  ADD IF NOT EXISTS slug VARCHAR(255) NOT NULL DEFAULT NULL,
  ADD IF NOT EXISTS name VARCHAR(32) NOT NULL DEFAULT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS tag_slug_uindex ON tag(slug);
CREATE UNIQUE INDEX IF NOT EXISTS tag_name_uindex ON tag(name);

CREATE TABLE IF NOT EXISTS user_account (
  id BIGSERIAL PRIMARY KEY
);

ALTER TABLE user_account
  ADD IF NOT EXISTS created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
  ADD IF NOT EXISTS updated_at  TIMESTAMP NOT NULL DEFAULT NOW(),
  ADD IF NOT EXISTS deleted_at  TIMESTAMP,
  ADD IF NOT EXISTS name        VARCHAR(32) NOT NULL DEFAULT NULL,
  ADD IF NOT EXISTS email       VARCHAR(255) NOT NULL DEFAULT NULL,
  ADD IF NOT EXISTS password    TEXT NOT NULL DEFAULT NULL,
  ADD IF NOT EXISTS permissions VARCHAR(32)[] NOT NULL DEFAULT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS user_account_email_uindex ON user_account(email);
CREATE INDEX IF NOT EXISTS user_account_created_at_index ON user_account(created_at);
CREATE INDEX IF NOT EXISTS user_account_updated_at_index ON user_account(updated_at);
CREATE INDEX IF NOT EXISTS user_account_deleted_at_index ON user_account(deleted_at);

CREATE TABLE IF NOT EXISTS project (
  id BIGSERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS cover (
  id BIGSERIAL PRIMARY KEY
);

ALTER TABLE cover
  ADD IF NOT EXISTS created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  ADD IF NOT EXISTS updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  ADD IF NOT EXISTS project_id BIGINT NOT NULL DEFAULT NULL REFERENCES project(id) ON DELETE CASCADE,
  ADD IF NOT EXISTS file_name  VARCHAR(255) NOT NULL DEFAULT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS cover_pid_fn_uindex ON cover(project_id, file_name);
CREATE INDEX IF NOT EXISTS cover_created_at_index ON cover(created_at);
CREATE INDEX IF NOT EXISTS cover_updated_at_index ON cover(updated_at);
CREATE INDEX IF NOT EXISTS cover_project_id_index ON cover(project_id);

ALTER TABLE project 
  ADD IF NOT EXISTS slug              VARCHAR(255) NOT NULL DEFAULT NULL,
  ADD IF NOT EXISTS locked            BOOLEAN DEFAULT FALSE,
  ADD IF NOT EXISTS created_at        TIMESTAMP NOT NULL DEFAULT NOW(),
  ADD IF NOT EXISTS updated_at        TIMESTAMP NOT NULL DEFAULT NOW(),
  ADD IF NOT EXISTS published_at      TIMESTAMP,
  ADD IF NOT EXISTS title             VARCHAR(128) NOT NULL DEFAULT NULL,
  ADD IF NOT EXISTS description       VARCHAR(4096) DEFAULT NULL,
  ADD IF NOT EXISTS cover_id          BIGINT DEFAULT NULL REFERENCES cover(id) ON DELETE SET NULL,
  ADD IF NOT EXISTS project_status    VARCHAR(32) NOT NULL DEFAULT NULL,
  ADD IF NOT EXISTS series_status     VARCHAR(32) NOT NULL DEFAULT NULL,
  ADD IF NOT EXISTS demographic       VARCHAR(32) DEFAULT NULL,
  ADD IF NOT EXISTS rating            VARCHAR(32) DEFAULT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS project_slug_uindex ON project(slug);
CREATE UNIQUE INDEX IF NOT EXISTS project_title_uindex ON project(title);
CREATE UNIQUE INDEX IF NOT EXISTS project_slug_title_uindex ON project(slug, title);
CREATE INDEX IF NOT EXISTS project_locked_index ON project(locked);
CREATE INDEX IF NOT EXISTS project_created_at_index ON project(created_at);
CREATE INDEX IF NOT EXISTS project_updated_at_index ON project(updated_at);
CREATE INDEX IF NOT EXISTS project_published_at_index ON project(published_at);
CREATE INDEX IF NOT EXISTS project_title_index ON project(title);
CREATE INDEX IF NOT EXISTS project_cover_id_index ON project(cover_id);
CREATE INDEX IF NOT EXISTS project_project_status_index ON project(project_status);
CREATE INDEX IF NOT EXISTS project_series_status_index ON project(series_status);
CREATE INDEX IF NOT EXISTS project_demographic_index ON project(demographic);
CREATE INDEX IF NOT EXISTS project_rating_index ON project(rating);

CREATE TABLE IF NOT EXISTS project_authors (
  project_id  BIGINT NOT NULL DEFAULT NULL REFERENCES project(id) ON DELETE CASCADE,
  author_id   BIGINT NOT NULL DEFAULT NULL REFERENCES author(id) ON DELETE CASCADE,
  PRIMARY KEY(project_id, author_id)
);

CREATE INDEX IF NOT EXISTS project_authors_project_id_index ON project_authors(project_id);
CREATE INDEX IF NOT EXISTS project_authors_author_id_index ON project_authors(author_id);

CREATE TABLE IF NOT EXISTS project_artists (
  project_id  BIGINT NOT NULL DEFAULT NULL REFERENCES project(id) ON DELETE CASCADE,
  artist_id   BIGINT NOT NULL DEFAULT NULL REFERENCES author(id) ON DELETE CASCADE,
  PRIMARY KEY(project_id, artist_id)
);

CREATE INDEX IF NOT EXISTS project_artists_project_id_index ON project_artists(project_id);
CREATE INDEX IF NOT EXISTS project_artists_artist_id_index ON project_artists(artist_id);

CREATE TABLE IF NOT EXISTS project_tags (
  project_id  BIGINT NOT NULL DEFAULT NULL REFERENCES project(id) ON DELETE CASCADE,
  tag_id      BIGINT NOT NULL DEFAULT NULL REFERENCES tag(id) ON DELETE CASCADE,
  PRIMARY KEY(project_id, tag_id)
);

CREATE INDEX IF NOT EXISTS project_tags_project_id_index ON project_tags(project_id);
CREATE INDEX IF NOT EXISTS project_tags_tag_id_index ON project_tags(tag_id);

CREATE TABLE IF NOT EXISTS chapter (
  id BIGSERIAL PRIMARY KEY
);

ALTER TABLE chapter
  ADD IF NOT EXISTS locked        BOOLEAN DEFAULT FALSE,
  ADD IF NOT EXISTS created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
  ADD IF NOT EXISTS updated_at    TIMESTAMP NOT NULL DEFAULT NOW(),
  ADD IF NOT EXISTS published_at  TIMESTAMP,
  ADD IF NOT EXISTS project_id    BIGINT NOT NULL DEFAULT NULL REFERENCES project(id) ON DELETE CASCADE,
  ADD IF NOT EXISTS uploader_id   BIGINT DEFAULT NULL REFERENCES user_account(id) ON DELETE SET NULL,
  ADD IF NOT EXISTS chapter       VARCHAR(8) NOT NULL DEFAULT NULL,
  ADD IF NOT EXISTS volume        VARCHAR(8) DEFAULT NULL,
  ADD IF NOT EXISTS title         VARCHAR(128) DEFAULT NULL,
  ADD IF NOT EXISTS pages         VARCHAR(255)[] DEFAULT NULL;

CREATE INDEX IF NOT EXISTS chapter_locked_index ON chapter(locked);
CREATE INDEX IF NOT EXISTS chapter_created_at_index ON chapter(created_at);
CREATE INDEX IF NOT EXISTS chapter_updated_at_index ON chapter(updated_at);
CREATE INDEX IF NOT EXISTS chapter_published_at_index ON chapter(published_at);
CREATE INDEX IF NOT EXISTS chapter_project_id_index ON chapter(project_id);
CREATE INDEX IF NOT EXISTS chapter_uploader_id_index ON chapter(uploader_id);

CREATE TABLE IF NOT EXISTS chapter_scanlation_groups (
  chapter_id BIGINT NOT NULL DEFAULT NULL REFERENCES chapter(id) ON DELETE CASCADE,
  scanlation_group_id   BIGINT NOT NULL DEFAULT NULL REFERENCES scanlation_group(id) ON DELETE CASCADE,
  PRIMARY KEY(chapter_id, scanlation_group_id)
);

CREATE INDEX IF NOT EXISTS chapter_scanlation_groups_chapter_id_index ON chapter_scanlation_groups(chapter_id);
CREATE INDEX IF NOT EXISTS chapter_scanlation_groups_scanlation_group_id_index ON chapter_scanlation_groups(scanlation_group_id);

CREATE TABLE IF NOT EXISTS statistics (
  id BIGSERIAL PRIMARY KEY
);

ALTER TABLE statistics  
  ADD IF NOT EXISTS             project_id          BIGINT DEFAULT NULL REFERENCES project(id) ON DELETE CASCADE,
  ADD IF NOT EXISTS             chapter_id          BIGINT DEFAULT NULL REFERENCES chapter(id) ON DELETE CASCADE,
  ADD IF NOT EXISTS             view_count          BIGINT NOT NULL DEFAULT 0,
  ADD IF NOT EXISTS             unique_view_count   BIGINT NOT NULL DEFAULT 0,
  DROP CONSTRAINT IF EXISTS     statistics_check,
  ADD CONSTRAINT                statistics_check    CHECK(project_id > 0 OR chapter_id > 0);

CREATE UNIQUE INDEX IF NOT EXISTS statistics_project_id_uindex ON statistics(project_id);
CREATE UNIQUE INDEX IF NOT EXISTS statistics_chapter_id_uindex ON statistics(chapter_id);
# Kasen

Oh no, a yet-another open-source CMS for scanlators. Anyways...

The back-end is written in Go, so it's memory-efficient, but make sure that your server has at least 512 MB of RAM.

The front-end is written in Go template, so you don't need to waste your server resources on server-side rendering a JavaScript front-end. It's lightweight and mobile-friendly, it's not the best front-end but it's pretty straightforward and not bloated. No analytics, tracking or third-party scripts.

The admin front-end is written in TypeScript/React (Client-side rendering) because it requires interactivity, and it's pain in the ass to write in pure HTML/Go template.

The reader is written in TypeScript/React (Client-side rendering), reusing the reader component from [nonbiri](https://github.com/rs1703/nonbiri), and the legacy reader is written in Go template (Server-side rendering) to support NoScript users.

The back-end uses in-memory LRU cache to cache database queries and templates, meanwhile, the front-end uses service worker to cache pages, assets and files.

## Table of Content

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Setup](#setup)
  - [Install the prerequisites](#install-the-prerequisites)
  - [Initialize database cluster](#initialize-database-cluster)
  - [Enable and start PostgreSQL and Redis](#enable-and-start-postgresql-and-redis)
  - [Create a new database and user/role](#create-a-new-database-and-user-role)
  - [Create a new system user](#create-a-new-system-user)
  - [Create data and configuration directories](#create-data-and-configuration-directories)
  - [Build the back-end](#build-the-back-end)
  - [Build or download the front-end](#build-or-download-the-front-end)
  - [Create a Systemd unit file](#create-a-systemd-unit-file)
  - [Start the back-end](#start-the-back-end)
- [NGINX Setup](#nginx-setup)
- [Models](#models)
  - [Project](#project)
  - [Chapter](#chapter)
  - [Permissions](#permissions)
- [Directory structure](#directory-structure)
- [Notes](#notes)
- [License](#license)

## Features

- [x] Manga, covers and chapters management
- [x] Markdown-based editor (Project description)
- [x] Permission-based access control
- [x] View and unique view count
- [x] RSS and Atom feeds
- [x] Search engine
- [x] Import manga metadata and auto-download main cover from MangaD\*x
- [x] Import chapter metadata and auto-download pages from MangaD\*x
- [x] Resizable covers and pages
- [x] Download chapter as a zipped file

Chapter pages are ordered by digits in file names. Left zero padding, prefixes and suffixes will not affect page ordering.

## Prerequisites

Prerequisites for building and running the back-end

- Git
- Go 1.17+
- libvips 8.3+ (8.8+ recommended)
- C compatible compiler such as gcc 4.6+ or clang 3.0+
- PostgreSQL
- Redis

## Setup

### Install the prerequisites

```sh
# Arch-based distributions
sudo pacman -Syu
sudo pacman -S git go libvips postgresql redis

# Debian-based distributions
sudo apt-get install -y software-properties-common
sudo add-apt-repository -y ppa:strukturag/libde265
sudo add-apt-repository -y ppa:strukturag/libheif
sudo add-apt-repository -y ppa:tonimelisma/ppa
sudo add-apt-repository -y ppa:longsleep/golang-backports

sudo apt-get update -y
sudo apt-get install -y build-essential git golang-go libvips-dev postgresql redis-server
```

### Initialize database cluster

**Only for Arch-based distributions** - Before PostgreSQL can function correctly, the database cluster must be initialized - [wiki.archlinux.org](https://wiki.archlinux.org/title/PostgreSQL#Installation).

```sh
echo initdb -D /var/lib/postgres/data | sudo su - postgres
```

### Enable and start PostgreSQL and Redis

```sh
# Arch-based distributions
systemctl --now enable postgresql redis

# Debian-based distributions
systemctl --now enable postgresql redis-server
```

### Create a new database and user/role

```sh
sudo -u postgres psql --command "CREATE USER kasen LOGIN SUPERUSER PASSWORD 'kasen';"
# expected output: CREATE ROLE

sudo -u postgres psql --command "CREATE DATABASE kasen OWNER kasen;"
# expected output: CREATE DATABASE
```

### Create a new system user

```sh
# Arch-based distributions
sudo useradd \
  --system \
  --shell /sbin/nologin \
  --user-group \
  --create-home \
  --home-dir /home/kasen \
  kasen

# Debian-based distributions
sudo adduser \
  --system \
  --shell /bin/bash \
  --group \
  --disabled-password \
  --home /home/kasen \
  kasen
```

### Create data and configuration directories

```sh
sudo mkdir -p /var/lib/kasen/{assets,data,templates}
sudo chown -R kasen:kasen /var/lib/kasen
sudo chmod -R 755 /var/lib/kasen

sudo mkdir /etc/kasen
sudo chown kasen:kasen /etc/kasen
sudo chmod 755 /etc/kasen
```

### Build the back-end

```sh
git clone https://github.com/rs1703/Kasen
cd Kasen
sudo go build -ldflags="-s -w" -o /usr/local/bin/kasen
sudo go build -ldflags="-s -w" -o /usr/local/bin/kasen-image ./internal/cmd/image/image.go
sudo chmod +x /usr/local/bin/kasen /usr/local/bin/kasen-image
```

### Build or download the front-end

You need **Node.js** and **Yarn** to be able to build the front-end.

```sh
# Build
cd web
yarn && yarn prod
sudo mv ../bin/assets ../bin/templates -t /var/lib/kasen
sudo chown -R kasen:kasen /var/lib/kasen
sudo chmod -R 755 /var/lib/kasen/{assets,templates}

# Download
wget -O front-end.tar.xz https://github.com/rs1703/Kasen/releases/download/v0.1.1/front-end.tar.xz
sudo tar -xf front-end.tar.xz -C /var/lib/kasen
sudo chown -R kasen:kasen /var/lib/kasen
sudo chmod -R 755 /var/lib/kasen/{assets,templates}
```

### Create a Systemd unit file

```sh
sudo wget https://raw.githubusercontent.com/rs1703/Kasen/master/kasen.service -P /etc/systemd/system/
```

### Start the back-end

The back-end will prompt you name, email and password when you start it for the first time, so you have to run it manually, and not through the systemd.

```sh
sudo -u kasen kasen -config=/etc/kasen/config.ini
```

Once the setup is completed, exit the back-end by pressing <kbd>CTRL+C</kbd> and then start the back-end through the systemd.

```sh
sudo systemctl daemon-reload
sudo systemctl enable --now kasen
```

The back-end is running on port 42072 by default. The front-end is accessible by going to `http://localhost:42072`, and the admin front-end is accessible on the `/admin` path. Replace localhost with your server's IP address or domain.

A configuration will be generated in the `/etc/kasen` directory once you run the back-end. Variables such as server port, site base url, title, description and language are stored inside it, but you can change the site meta and service configurations from the front-end. You have to modify the base url if you are using your own domain.

If you want the front-end to be accessible without using port, then change the port which used by the back-end to 80 or use a [reverse-proxy](#nginx-setup) (recommended).

## NGINX Setup

It's recommended to deploy the back-end behind a reverse-proxy. The followings are the basic steps for installing and configuring NGINX on your server.

```sh
# Arch-based distributions
sudo pacman -S nginx-mainline nano

# Debian-based distributions
sudo apt-get install nginx nano
```

Create the required directories and configuration file.

```sh
sudo mkdir /etc/nginx/{sites-available,sites-enabled}
sudo touch /etc/nginx/sites-available/kasen
sudo ln -s /etc/nginx/sites-available/kasen /etc/nginx/sites-enabled/
sudo nano /etc/nginx-sites-available/kasen
```

Then copy and paste the following

```sh
server {
  listen 80;
  listen [::]:80;

  server_name yourdomain.com www.yourdomain.com;

  location / {
    proxy_pass http://localhost:42072;
  }
}
```

Press <kbd>CTRL+SHIFT+V</kbd> or <kbd>SHIFT+INSERT</kbd> to paste, and press <kbd>CTRL+X</kbd> to save the file.

---

You might as well use Cloudflare or DDoS-Guard to hide your server's IP address, then use certbot for a free SSL/TLS. You should also disable access by IP address, install a firewall and only allow frequently used ports such as 22, 80 and 443. You could find the tutorials on search engines, as they are not within the scope of this project.

## Models

### Project

- Title - required, unique
- Description
- Artists
- Authors
- Project Status - required
- Series status - required
- Demographic
- Content Rating

### Chapter

- Chapter number - required
- Volume number
- Title
- Scanlation groups

### Permissions

**Permission name - grants the ability to**

- create_project - create a new project
- edit_project - edit projects
- publish_project - publish projects
- unpublish_project - unpublish projecs
- lock_project - lock projects
- unlock_project - unlock projects
- delete_project - delete projects
- upload_cover - upload a new cover
- set_cover - set main cover of a project
- delete_cover - delete covers
- create_chapter - create a new chapter
- edit_chapter - edit chapters created by the user
- edit_chapters - edit chapters created by all users
- publish_chapter - publish chapters created by the user
- publish_chapters - publish chapters created by all users
- unpublish_chapter - unpublish chapters created by the user
- unpublish_chapters - unpublish chapters created by all users
- lock_chapter - lock chapters created by the user
- lock_chapters - lock chapters created by all users
- unlock_chapter - unlock chapters created by the user
- unlock_chapters - unlock chapters created by all users
- delete_chapter - delete chapters created by the users
- delete_chapters - delete chapters created by all users
- edit_user - edit name and password of the user
- edit_users - edit name and password of all users
- delete_user - delete user's account
- delete_users - delete all user accounts
- manage - update meta and service configuration

## Directory structure

```sh
├── data
│   ├── tmp # temporary files
│   ├── touhou-ibara-kasen-wild-and-horned-hermit # human-readable directory
│   │   ├── chapters
│   │   │   ├── vol-10-ch-47-schadenfreude-utopia
│   │   │   ├── vol-10-ch-48-not-stopping-to-ask-for-direction-in-the-land-of-darkness
│   │   ├── covers
│   ├── symlinks # symbolic links, points to the real directory
│   │   ├── 18 -> data/touhou-ibara-kasen-wild-and-horned-hermit
│   │   ├── chapters # symbolic links
│   │   │   ├── 427 -> data/touhou-ibara-kasen-wild-and-horned-hermit/vol-10-ch-47-schadenfreude-utopia
│   │   │   ├── 428 -> data/touhou-ibara-kasen-wild-and-horned-hermit/vol-10-ch-48-not-stopping-to-ask-for-direction-in-the-land-of-darkness
```

## Notes

**Navigation menu** - Unfortunately, you can not change the navigation menu from the front-end, so you have to edit the `header.html` manually. It's better than bloating the front-end with a draggable component. Don't forget to restart the back-end once you have modified the templates (`sudo systemctl restart kasen`).

**Registration and new users** - Registration is disabled by default, and new users will have the following permissions by default: `create_project, upload_cover, set_cover, edit_user, delete_user, create_chapter, edit_chapter, lock_chapter, publish_chapter, unlock_chapter, unpublish_chapter`.

Same with navigation menu, you can not change other user's permissions from the front-end. There's an api for it, but I'm too lazy to write a custom front-end for it, so you have to use PgAdmin or DBeaver to access your database and edit the permissions manually.

## License

**Kasen** is licensed under the [GNU Affero General Public License v3.0](https://www.gnu.org/licenses/agpl-3.0.en.html).

# Apple Store Refurbished Machine Inventory Monitor Bot

https://www.apple.com.cn/shop/refurbished/mac/2021-macbook-pro

## Feature

- Lark notification
- Docker deployment
- No deps and minimal

## Deploy

Docker Compose to deploy.

Firstly, clone the code base:

```sh
git clone https://github.com/qwqcode/apple-inventory-bot.git
```

Use the Docker compose to build container image:

```sh
docker compose build
```

Copy the `.env.example` file to `.env`, and edit it by your preference.

```sh
cp .env.example .env

vim .env
```

Please notice the meaning of the env variables:

|Key|Description|Details|
|-|-|-|
|COOKIES|The COOKIE when launching a request|key-value pairs format. |
|LARK|The lark notification webhook URL| URL start with `https://open.feishu.cn/open-apis/bot/v2/hook/`. |
|KEYWORDS|The keywords of MODEL TITLE for each inventory available check|use `,` to give the multiple keywords, leave empty to omit.|
|MEM|The memory size for each inventory available check|e.g. `"32gb"`. (leave empty to omit) |
|DISK|The disk size for each inventory available check|e.g. `"512gb"`. (leave empty to omit) |
|LIST_URL|The request url of the list page in Apple Store|default is `https://www.apple.com.cn/shop/refurbished/mac/2021-macbook-pro`|

Launch docker container and let it run in background:

```sh
docker compose up -d
```

Some commands may help:

```sh
docker compose ps
docker compose logs
```

## App lifecycle

An inventory check will be performed every minute.

Notice that the app will be EXIT automatically while the specific product model is fetched and send the Lark notification. So, if you find the docker container is down, it proven that the model that you want is available.

If you do not want to prevent exit the app, you could modify the `docker-compose.yml` file and add `restart: always`. But this may cause send lark notification duplicately.

## LICENSE

MIT

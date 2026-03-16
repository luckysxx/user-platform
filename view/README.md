# view

This template should help get you started developing with Vue 3 in Vite.

## Recommended IDE Setup

[VS Code](https://code.visualstudio.com/) + [Vue (Official)](https://marketplace.visualstudio.com/items?itemName=Vue.volar) (and disable Vetur).

## Recommended Browser Setup

- Chromium-based browsers (Chrome, Edge, Brave, etc.):
  - [Vue.js devtools](https://chromewebstore.google.com/detail/vuejs-devtools/nhdogjmejiglipccpnnnanhbledajbpd)
  - [Turn on Custom Object Formatter in Chrome DevTools](http://bit.ly/object-formatters)
- Firefox:
  - [Vue.js devtools](https://addons.mozilla.org/en-US/firefox/addon/vue-js-devtools/)
  - [Turn on Custom Object Formatter in Firefox DevTools](https://fxdx.dev/firefox-devtools-custom-object-formatters/)

## Type Support for `.vue` Imports in TS

TypeScript cannot handle type information for `.vue` imports by default, so we replace the `tsc` CLI with `vue-tsc` for type checking. In editors, we need [Volar](https://marketplace.visualstudio.com/items?itemName=Vue.volar) to make the TypeScript language service aware of `.vue` types.

## Customize configuration

See [Vite Configuration Reference](https://vite.dev/config/).

## Project Setup

```sh
pnpm install
```

### Compile and Hot-Reload for Development

```sh
pnpm dev
```

### Type-Check, Compile and Minify for Production

```sh
pnpm build
```

### Lint with [ESLint](https://eslint.org/)

```sh
pnpm lint
```

## SSO 注册页接入

当前注册页路由：

- `/register`（兼容 `/`）
- `/login`

接入应用可以跳转到：

```text
https://your-sso-domain/register?client_id=app_a&redirect_uri=https%3A%2F%2Fapp-a.com%2Fauth%2Fcallback&state=random_string
```

参数说明：

- `client_id`: 接入应用标识
- `redirect_uri`: 注册成功后的回跳地址
- `state`: 接入方自定义随机串，用于请求关联和安全校验

注册成功后，SSO 前端会回跳到 `redirect_uri`，并附带参数：

- `result=registered`
- `user_id`
- `username`
- `client_id`（如果传入）
- `state`（如果传入）

登录接入示例：

```text
https://your-sso-domain/login?client_id=app_a&redirect_uri=https%3A%2F%2Fapp-a.com%2Fauth%2Fcallback&state=random_string
```

登录成功后，SSO 前端会回跳到 `redirect_uri`：

- Query 参数：`result=logged_in`、`user_id`、`username`、`client_id`、`app_code`、`state`
- Hash 参数：`access_token`、`refresh_token`、`token_type`

说明：

- 登录接口要求 `app_code`，前端会优先读取 `app_code`，未提供时回退到 `client_id`。

## 后端 API 对接

注册请求会调用：

- `POST /api/v1/users/register`

本地开发默认通过 Vite 代理转发 `/api` 到 `http://localhost:8081`。

如果你需要直连其他环境，可在 `.env` 中设置：

```dotenv
VITE_API_BASE_URL=https://your-api-domain
```

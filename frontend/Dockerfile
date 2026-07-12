# ---- build ----
FROM node:22-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

# ---- run (Nitro node server) ----
FROM node:22-alpine
WORKDIR /app
COPY --from=build /app/.output ./.output
ENV HOST=0.0.0.0 PORT=3000
# NUXT_PUBLIC_API_BASE is read at runtime (set it in compose) so no rebuild is
# needed to point the app at a different API URL.
EXPOSE 3000
CMD ["node", ".output/server/index.mjs"]

FROM node:24.18.0-alpine3.23 AS build
ENV PNPM_HOME=/pnpm
ENV PATH=$PNPM_HOME:$PATH
RUN corepack enable && corepack prepare pnpm@11.12.0 --activate
WORKDIR /src
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/web/package.json apps/web/package.json
RUN pnpm install --frozen-lockfile
COPY apps/web apps/web
ARG NINEFOLD_REPLAY_SIGNING_KEY_ID
ARG NINEFOLD_REPLAY_PUBLIC_KEY
RUN test -n "${NINEFOLD_REPLAY_SIGNING_KEY_ID}" \
 && test -n "${NINEFOLD_REPLAY_PUBLIC_KEY}"
ENV VITE_NINEFOLD_REPLAY_SIGNING_KEY_ID=${NINEFOLD_REPLAY_SIGNING_KEY_ID}
ENV VITE_NINEFOLD_REPLAY_PUBLIC_KEY=${NINEFOLD_REPLAY_PUBLIC_KEY}
ENV VITE_NINEFOLD_PRODUCTION_BUILD=true
RUN pnpm --filter @ninefold/web build \
 && pnpm --filter @ninefold/web deploy --prod /out \
 && cp -R apps/web/build /out/build

FROM node:24.18.0-alpine3.23
ARG VERSION=unknown
LABEL org.opencontainers.image.title="Ninefold Sudoku web" \
      org.opencontainers.image.version="${VERSION}"
WORKDIR /app
ENV NODE_ENV=production
COPY --from=build --chown=node:node /out ./
RUN apk upgrade --no-cache \
 && rm -rf /usr/local/lib/node_modules/npm /usr/local/lib/node_modules/corepack /opt/yarn-v1.22.22 \
 && rm -f /usr/local/bin/npm /usr/local/bin/npx /usr/local/bin/corepack \
    /usr/local/bin/yarn /usr/local/bin/yarnpkg /usr/local/bin/pnpm /usr/local/bin/pnpx
USER node
EXPOSE 3000
CMD ["node", "build"]

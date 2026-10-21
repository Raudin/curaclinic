# syntax=docker/dockerfile:1

FROM node:24-alpine AS build

WORKDIR /app
RUN npm install -g pnpm@12.3.4

COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY . .

# Vite inlines these at build time. The API defaults to the same origin: nginx
# below proxies /api to the backend container.
ARG VITE_API_BASE_URL=/api
ARG VITE_CLINIC_TIMEZONE=Africa/Nairobi
ARG VITE_SESSION_IDLE_MINUTES=30
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL \
    VITE_CLINIC_TIMEZONE=$VITE_CLINIC_TIMEZONE \
    VITE_SESSION_IDLE_MINUTES=$VITE_SESSION_IDLE_MINUTES

RUN pnpm build

FROM nginx:1.29-alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -q -O /dev/null http://127.0.0.1/healthz || exit 1

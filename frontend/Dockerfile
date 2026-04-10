# ─── Stage 1: Build ───────────────────────────────────────────────────────────
FROM node:20-alpine AS builder

WORKDIR /app

# VITE_API_URL baked into the bundle at build time
ARG VITE_API_URL=http://localhost:8080
ENV VITE_API_URL=$VITE_API_URL

# Install deps first (cached unless package.json changes)
COPY package.json package-lock.json ./
RUN npm ci

# Build
COPY . .
RUN npm run build

# ─── Stage 2: Serve ───────────────────────────────────────────────────────────
FROM nginx:1.25-alpine

# SPA routing config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy built assets
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]

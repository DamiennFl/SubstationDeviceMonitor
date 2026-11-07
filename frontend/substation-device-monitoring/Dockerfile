###
# BUILD STAGE
###

# Use latest node version
FROM node:26-alpine3.23 AS builder

# Define work directory
WORKDIR /frontend

# Copy package and package-lock for deps and Vite build scripts.
COPY package.json package-lock.json ./

# npm ci instead of npm install. npm ci installs exactly what is defined in package-lock, unlike npm install.
RUN npm ci

# Copy directory code
COPY . .

# build frontend using Vite scripts in package.json
RUN npm run build

###
# RUN STAGE
### 

# Nginx server
FROM nginx:1.31.8-alpine

# Copy built dist files to nginx html so it can server
COPY --from=builder /frontend/dist/ /usr/share/nginx/html

# Add Nginx config file
COPY default.conf.template /etc/nginx/templates/
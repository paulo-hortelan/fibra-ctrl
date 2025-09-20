# Build stage
FROM node:20-alpine as build

WORKDIR /app

# Copy package.json and package-lock.json
COPY package*.json ./
RUN npm install

# Copy the source code
COPY . .

# Build the app
RUN npm run build

# Development stage
FROM node:20-alpine as development

WORKDIR /app

# Copy package.json and package-lock.json
COPY package*.json ./
# Install dependencies and ensure binaries are available in PATH
RUN npm install && npm install -g vite typescript

# Copy the source code
COPY . .

EXPOSE 3000

# Use npm run dev directly instead of npm start
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]

# Production stage
FROM nginx:alpine as production

# Copy the build output to replace the default nginx contents
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]

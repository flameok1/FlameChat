# Use the official Golang image as the base image
FROM golang:1.20

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY ./goserve/go.mod ./goserve/go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY ./goserve/ .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

# Use the official Node.js image as the base image for building the frontend
FROM node:20 AS build

# Set the Current Working Directory inside the container
WORKDIR /vueapp

# Copy package.json and package-lock.json
COPY vueweb/package*.json ./

# Install dependencies
RUN npm install

# Copy the Vue.js app source code
COPY vueweb/ .

# Build the Vue.js app
RUN npm run build

# Use the official Nginx image to serve the frontend
FROM nginx:alpine

# Copy the built files from the build stage to the Nginx html directory
COPY --from=build /vueapp/dist /usr/share/nginx/html
# Copy the Go app executable to the Nginx container
COPY --from=0 /app/main ./

# Copy the custom nginx configuration file to the Nginx configuration directory
COPY ./nginx.conf /etc/nginx/conf.d/default.conf

# Expose port 80 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["sh","-c","./main & nginx -g 'daemon off;'"]
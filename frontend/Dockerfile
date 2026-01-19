# Build stage
FROM oven/bun:1.3.6 AS builder

WORKDIR /app

# Copy package files and install dependencies
COPY package.json bun.lock ./
RUN bun install

# Copy the rest of the application code
COPY . .

# Build the SvelteKit application
RUN bun run build

# Final stage
FROM oven/bun:1.3.6

WORKDIR /app

# Copy the build output and package files
COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json ./package.json

# Set environment variables
ENV PORT=3000
ENV NODE_ENV=production

EXPOSE 3000

# Run the SvelteKit server with Bun
CMD ["bun", "./build/index.js"]

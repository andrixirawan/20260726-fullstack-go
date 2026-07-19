FROM node:20-alpine

WORKDIR /app

# Copy package manifests first for better layer caching
COPY package.json package-lock.json* yarn.lock* ./

# Install dependencies
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi

# Copy source code
COPY . .

# Build the Next.js application
RUN npm run build

EXPOSE 8080

CMD ["npm", "run", "start"]

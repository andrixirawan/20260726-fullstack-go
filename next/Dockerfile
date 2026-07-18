FROM node:20-alpine

WORKDIR /app

# Copy package manifests first for better layer caching
COPY package.json ./

# Install dependencies
RUN npm install

# Copy source code
COPY . .

EXPOSE 3000

CMD ["npm", "run", "dev", "--", "--hostname", "0.0.0.0"]

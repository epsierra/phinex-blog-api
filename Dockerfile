# Production stage only — since we're building dist outside Docker
FROM node:22.14.0 AS prod

WORKDIR /app

COPY package*.json ./
COPY tsconfig.json ./

RUN npm install --production

# Install nanoid version 5 specifically
RUN npm install nanoid@5

# Copy only required artifacts
COPY dist ./dist
COPY views ./views
# COPY logs ./logs
COPY README.md ./README.md

EXPOSE 5000

CMD ["npm", "start"]

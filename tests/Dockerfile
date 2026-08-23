FROM mcr.microsoft.com/playwright:v1.49.1-jammy
WORKDIR /tests
ENV PYTHONIOENCODING=utf-8
COPY package.json ./
RUN npm install --registry=https://registry.npmmirror.com
COPY . .
CMD ["npx", "playwright", "test"]

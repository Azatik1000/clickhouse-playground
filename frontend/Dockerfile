# pull official base image
FROM node:13.12.0-alpine

# set working directory
WORKDIR /install
COPY package.json .
COPY package-lock.json .
RUN npm install --no-package-lock
RUN npm install react-scripts@3.4.1 -g

ENV NODE_PATH /install/node_modules
ENV PATH /install/node_modules/.bin:$PATH

# add app
COPY . .

# start app
CMD ["npm", "start"]

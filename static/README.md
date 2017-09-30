# zanecloud

> Zane Cloud Management System

## Build Setup

``` bash
#install dependencies
npm install
```

## Develop with hot reload 
``` bash
# with localhost:8080
npm run dev

# with live apiServer 8080
npm run dev -- -s 192.168.56.2

# with live apiServer and other port
npm run dev -- -s 192.168.56.2 -p 8090

# mock apiServer 
npm run dev -- -m

```
## build for production with minification
``` bash
npm run build
```

For detailed explanation on how things work, consult the [docs for vue-loader](http://vuejs.github.io/vue-loader).

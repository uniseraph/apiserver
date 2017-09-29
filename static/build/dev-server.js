const path = require('path')
const opn = require('opn')
const express = require('express');
const webpack = require('webpack');
const connectHistoryApiFallback = require('connect-history-api-fallback')();
const proxyMiddleware = require('http-proxy-middleware')
const webpackDevMiddleware = require('webpack-dev-middleware');
const webpackHotMiddleware = require('webpack-hot-middleware');
const app = express();
const config = require('../webpack.config.js');

process.env.NODE_ENV = 'development';
//取apiServer地址
const prog = require('commander');
prog
  .version('0.1.0')
  .option('-s, --server [server]', 'optional apiServer address', 'localhost')
  .option('-p, --port [port]', 'optional apiServer port', '8080')
  .option('-m, --mock', 'optional mock apiServer')
  .parse(process.argv);

const proxyOption = {target: 'http://' + prog.server + ':' + prog.port};


//调试时，输出url路径是根路径。名称中不加hash，减少编译时间
config.output.publicPath = '/';
config.output.filename = '[name].js';
//web客户端接收服务器的推送
Object.keys(config.entry).forEach(function (name) {
  config.entry[name] = ['./build/dev-client'].concat(config.entry[name])
})
//服务端启用推送
config.plugins = (config.plugins || []).concat([
  new webpack.HotModuleReplacementPlugin(),
  new webpack.NoEmitOnErrorsPlugin(),
])

const compiler = webpack(config);

let devMiddleware = webpackDevMiddleware(compiler, {
  publicPath: '/'
});
var hotMiddleware = webpackHotMiddleware(compiler, {
  log: false,
  heartbeat: 2000
})

//index.html变化时，推送
compiler.plugin('compilation', function (compilation) {
  compilation.plugin('html-webpack-plugin-after-emit', function (data, cb) {
    hotMiddleware.publish({action: 'reload'})
    cb()
  })
})
//public 静态目录
app.use('/public', express.static(path.join(__dirname, '../public')))

app.use(connectHistoryApiFallback)
app.use(devMiddleware);
app.use(hotMiddleware);
if (prog.mock) {
  console.log('> mock apiServer');
  const fs = require('fs');
  app.all('/api/:path([\\s\\S]*)', function (req, res, next) {
    console.log('> api request', req.params.pth);
    const filename = path.join(__dirname, '../public/mock', req.params.path);
    res.send(fs.readFileSync(filename));
  })
} else {
//后台接口转发
  app.use(proxyMiddleware('/api', proxyOption))
}

console.log('> Starting dev server...')
const port = 3000;
const uri = 'http://localhost:' + port;
var _resolve
var readyPromise = new Promise(resolve => {
  _resolve = resolve
})
devMiddleware.waitUntilValid(() => {
  console.log('> Listening at ' + uri + '\n')
  //自动打开调试页面
  opn(uri)
  _resolve()
})

var server = app.listen(port)

module.exports = {
  ready: readyPromise,
  close: () => {
    server.close()
  }
}
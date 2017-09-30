/* eslint-disable */
require('eventsource-polyfill')
var hotClient = require('webpack-hot-middleware/client?noInfo=true&reload=true')

hotClient.subscribe(function (event) {
  //reload 事件只在index.html变化时产生
  if (event.action === 'reload') {
    window.location.reload()
  }
})

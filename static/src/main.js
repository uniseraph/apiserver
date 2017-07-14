import Vue from 'vue'
import Vuetify from 'vuetify'
import axios from 'axios'
import store from './vuex/store'
import router from './router'
import common from './common'
import constants from './constants'
import App from './App.vue'

Vue.use(Vuetify)
Vue.use(common)
Vue.use(constants)

Vue.filter('formatDate', function(value) {
	if (!value || value.length == 0) {
		return '';
	}

    return new Date(value * 1000).toLocaleDateString();
});

Vue.filter('formatDateTime', function(value) {
	if (!value || value.length == 0) {
		return '';
	}
	
    return new Date(value * 1000).toLocaleString();
});

Vue.filter('dividedBy1024', function(value) {
	if (!value || value.length == 0) {
		return '';
	}

	if (typeof value !== 'number') {
		value = parseInt(value.toString());
	}
	
    return Math.floor(value / 1024);
});

new Vue({
  el: '#app',
  store,
  router,
  render: h => h(App)
})

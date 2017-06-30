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
	if (!value) {
		return '';
	}

    return new Date(value).toLocaleDateString();
});

Vue.filter('formatDateTime', function(value) {
	if (!value) {
		return '';
	}
	
    return new Date(value).toLocaleString();
});

new Vue({
  el: '#app',
  store,
  router,
  render: h => h(App)
})
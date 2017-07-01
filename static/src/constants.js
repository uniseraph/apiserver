export default {
  install(Vue, options) {
    Vue.prototype.constants = {
    	ROLE_NORMAL_USER: 0x01,
    	ROLE_SYS_ADMIN: 0x02,
    	ROLE_APP_ADMIN: 0x04
    }
  }
}
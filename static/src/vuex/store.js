import Vue from 'vue'
import Vuex from 'vuex'
import UIStore from './ui-store'
import UserStore from './user-store'

Vue.use(Vuex)

export default new Vuex.Store({
    modules: {
        UIStore,
        UserStore
    }
})
import store from '../vuex/store'

export function setToken(token) {
    store.dispatch('token', token)
}
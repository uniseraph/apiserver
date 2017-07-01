import store from '../vuex/store'

export function alert(msg, type='warning') {
    store.dispatch('showAlert', true)
    store.dispatch('alertType', type)
    store.dispatch('alertMsg', msg)

    setTimeout(() => {
        store.dispatch('showAlert', false); 
    }, 1500);
}
import store from '../vuex/store'

export function alert(msg, type='warning') {
    store.dispatch('alertArea', store.getters.showAlertAt)
    store.dispatch('alertType', type)
    store.dispatch('alertMsg', msg)

    setTimeout(() => {
        store.dispatch('alertArea', null); 
    }, type == 'success' ? 1500 : 2500);
}

export function showAlertAt(area='global') {
	store.dispatch('showAlertAt', area); 
}
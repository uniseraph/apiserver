import store from '../vuex/store'

export function alert(msg, type='warning') {
    store.dispatch('alertArea', store.getters.showAlertAt)
    store.dispatch('alertType', type)
    store.dispatch('alertMsg', msg)

    if (type == 'success') {
	    setTimeout(() => {
	        store.dispatch('alertArea', null); 
	    }, 1500);
	}
}

export function showAlertAt(area='global') {
	store.dispatch('showAlertAt', area); 
}
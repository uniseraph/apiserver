const state = {
    showAlert: false,
    alertType: 'warning',
    alertMsg: ''
}

const getters = {
    showAlert: state => state.showAlert,
    alertType: state => state.alertType,
    alertMsg: state => state.alertMsg,
}

const SHOW_ALERT = 'SHOW_ALERT'
const ALERT_TYPE = 'ALERT_TYPE'
const ALERT_MSG = 'ALERT_MSG'

const actions = {
    showAlert({ commit }, status) {
        commit(SHOW_ALERT, status)
    },
    alertType({ commit }, str) {
        commit(ALERT_TYPE, str)
    },
    alertMsg({ commit }, str) {
        commit(ALERT_MSG, str)
    }
}

const mutations = {
	[SHOW_ALERT](state, status) {
        state.showAlert = status
    },
	[ALERT_TYPE](state, str) {
        state.alertType = str
    },
    [ALERT_MSG](state, str) {
        state.alertMsg = str
    }
}

export default {
    state,
    getters,
    actions,
    mutations
}
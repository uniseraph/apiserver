const state = {
    showAlertAt: 'global',
    alertArea: null,
    alertType: null,
    alertMsg: null
}

const getters = {
    showAlertAt: state => state.showAlertAt,
    alertArea: state => state.alertArea,
    alertType: state => state.alertType,
    alertMsg: state => state.alertMsg,
}

const SHOW_ALERT_AT = 'SHOW_ALERT_AT'
const ALERT_AREA = 'ALERT_AREA'
const ALERT_TYPE = 'ALERT_TYPE'
const ALERT_MSG = 'ALERT_MSG'

const actions = {
    showAlertAt({ commit }, status) {
        commit(SHOW_ALERT_AT, status)
    },
    alertArea({ commit }, str) {
        commit(ALERT_AREA, str)
    },
    alertType({ commit }, str) {
        commit(ALERT_TYPE, str)
    },
    alertMsg({ commit }, str) {
        commit(ALERT_MSG, str)
    }
}

const mutations = {
	[SHOW_ALERT_AT](state, str) {
        state.showAlertAt = str
    },
    [ALERT_AREA](state, str) {
        state.alertArea = str
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
const state = {
    token: null
}

const getters = {
    token: state => state.token
}

const TOKEN = 'TOKEN'

const actions = {
    token({ commit }, token) {
        commit(TOKEN, token)
    }
}

const mutations = {
	[TOKEN](state, token) {
        state.token = token
    }
}

export default {
    state,
    getters,
    actions,
    mutations
}
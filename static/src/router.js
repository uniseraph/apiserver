import Vue from 'vue'
import VueRouter from 'vue-router'

import Pools from './pages/Pools.vue'
import PoolDetail from './pages/PoolDetail.vue'
import Envs from './pages/Envs.vue'
import EnvDetail from './pages/EnvDetail.vue'
import Users from './pages/Users.vue'
import CreateUser from './pages/CreateUser.vue'
import UserDetail from './pages/UserDetail.vue'
import ResetPassword from './pages/ResetPassword.vue'
import Teams from './pages/Teams.vue'
import TeamDetail from './pages/TeamDetail.vue'

Vue.use(VueRouter)

export default new VueRouter({
	routes: [
		{ path: '/pools', component: Pools },
		{ path: '/pool/:id/detail', component: PoolDetail },
		{ path: '/envs', component: Envs },
		{ path: '/env/:id/detail', component: EnvDetail },
		{ path: '/users', component: Users },
		{ path: '/users/create', component: CreateUser },
		{ path: '/user/:id/detail', component: UserDetail },
		{ path: '/user/:id/password', component: ResetPassword },
		{ path: '/teams', component: Teams },
		{ path: '/team/:id/detail', component: TeamDetail },
		{ path: '*', redirect: '/pools' }
	]
})
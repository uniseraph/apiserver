import Vue from 'vue'
import VueRouter from 'vue-router'

import Pools from './pages/Pools.vue'
import PoolDetail from './pages/PoolDetail.vue'
import EnvTrees from './pages/EnvTrees.vue'
import EnvTree from './pages/EnvTree.vue'
import EnvDetail from './pages/EnvDetail.vue'
import Users from './pages/Users.vue'
import CreateUser from './pages/CreateUser.vue'
import UserDetail from './pages/UserDetail.vue'
import ResetPassword from './pages/ResetPassword.vue'
import Teams from './pages/Teams.vue'
import TeamDetail from './pages/TeamDetail.vue'
import Templates from './pages/Templates.vue'
import CreateTemplate from './pages/CreateTemplate.vue'
import TemplateDetail from './pages/TemplateDetail.vue'

Vue.use(VueRouter)

export default new VueRouter({
	routes: [
		{ path: '/pools', component: Pools },
		{ path: '/pool/:id', component: PoolDetail },
		{ path: '/env/trees', component: EnvTrees },
		{ path: '/env/tree/:id/:name', component: EnvTree },
		{ path: '/env/value/:id', component: EnvDetail },
		{ path: '/users', component: Users },
		{ path: '/users/create', component: CreateUser },
		{ path: '/user/:id', component: UserDetail },
		{ path: '/user/password/:id', component: ResetPassword },
		{ path: '/teams', component: Teams },
		{ path: '/team/:id', component: TeamDetail },
		{ path: '/templates', component: Templates },
		{ path: '/templates/create', component: CreateTemplate },
		{ path: '/template/:id', component: TemplateDetail },
		{ path: '*', redirect: '/pools' }
	]
})
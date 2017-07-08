import Vue from 'vue'
import VueRouter from 'vue-router'

import Pools from './pages/Pools.vue'
import PoolDetail from './pages/PoolDetail.vue'
import EnvTrees from './pages/EnvTrees.vue'
import EnvTree from './pages/EnvTree.vue'
import EnvDetail from './pages/EnvDetail.vue'
import Users from './pages/Users.vue'
import UserDetail from './pages/UserDetail.vue'
import ResetPassword from './pages/ResetPassword.vue'
import Teams from './pages/Teams.vue'
import TeamDetail from './pages/TeamDetail.vue'
import Templates from './pages/Templates.vue'
import TemplateDetail from './pages/TemplateDetail.vue'
import Applications from './pages/Applications.vue'
import ApplicationDetail from './pages/ApplicationDetail.vue'
import CreateApplication from './pages/CreateApplication.vue'
import UpgradeApplication from './pages/UpgradeApplication.vue'
import RollbackApplication from './pages/RollbackApplication.vue'
import Containers from './pages/Containers.vue'

Vue.use(VueRouter)

export default new VueRouter({
	routes: [
		{ path: '/pools/:id', component: PoolDetail },
		{ path: '/pools', component: Pools },
		{ path: '/env/trees/values/:id', component: EnvDetail },
		{ path: '/env/trees/:id/:name', component: EnvTree },
		{ path: '/env/trees', component: EnvTrees },
		{ path: '/users/password/:id', component: ResetPassword },
		{ path: '/users/create', component: UserDetail },
		{ path: '/users/:id', component: UserDetail },
		{ path: '/users', component: Users },
		{ path: '/teams/:id', component: TeamDetail },
		{ path: '/teams', component: Teams },
		{ path: '/templates/create', component: TemplateDetail },
		{ path: '/templates/copy/:id/:title', component: TemplateDetail },
		{ path: '/templates/:id', component: TemplateDetail },
		{ path: '/templates', component: Templates },
		{ path: '/applications/containers/:applicationId/:serviceName/:serviceTitle', component: Containers },
		{ path: '/applications/create/:poolId', component: CreateApplication },
		{ path: '/applications/:id/upgrade', component: UpgradeApplication },
		{ path: '/applications/:id/rollback', component: RollbackApplication },
		{ path: '/applications/:id', component: ApplicationDetail },
		{ path: '/applications', component: Applications },
		{ path: '*', redirect: '/pools' }
	]
})
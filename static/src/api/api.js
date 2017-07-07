import axios from 'axios'
import qs from 'qs'

import * as ui from '../util/ui'
import store from '../vuex/store'

// axios默认配置
axios.defaults.timeout = 5000;
axios.defaults.baseURL = 'http://localhost:8080/api';

// 仅测试用
/*
axios.defaults.baseURL = 'http://localhost:8080/public/mock';
axios.interceptors.request.use((config) => {
    if(config.method === 'post'){
        config.method = 'get';
        config.url = config.url + '?' + qs.stringify(config.data);
    }

    return config;
});
//*/

export function fetch(url, params) {
    return new Promise((resolve, reject) => {
        axios.post(url, params)
            .then(response => {
                resolve(response.data);
            }, error => {
                let res = error.response;
                if (res && res.status != 403) {
                    ui.alert(res.data);
                }

                reject(error);
            })
            .catch(error => {
                ui.alert('系统错误: ' + error);
                reject(error)
            })
    })
}

export default {

    Me() {
        return fetch('/users/current');
    },

    Login(params) {
        return fetch('/users/' + encodeURIComponent(params.Name) + '/login?Pass=' + encodeURIComponent(params.Password), params);
    },

    Pools(params) {
        return fetch('/pools/ps', params);
    },

    Pool(id) {
        return fetch('/pools/' + id + '/inspect');
    },

    CreatePool(params) {
        return fetch('/pools/register', params); 
    },

    RemovePool(id) {
        return fetch('/pools/' + id + '/remove'); 
    },

    UpdatePool(params) {
        return fetch('/pools/' + params.Id + '/update', params); 
    },

    AddTeamToPool(params) {
        return fetch('/pools/' + params.Id + '/add-team', params);
    },

    RemoveTeamFromPool(params) {
        return fetch('/pools/' + params.Id + '/remove-team', params);
    },

    AddUserToPool(params) {
        return fetch('/pools/' + params.Id + '/add-user', params);
    },

    RemoveUserFromPool(params) {
        return fetch('/pools/' + params.Id + '/remove-user', params);
    },

    EnvTrees(params) {
        return fetch('/envs/trees/list', params);
    },

    CreateEnvTree(params) {
        return fetch('/envs/trees/create', params);  
    },

    UpdateEnvTree(params) {
        return fetch('/envs/trees/' + params.Id + '/update', params);  
    },

    RemoveEnvTree(id) {
        return fetch('/envs/trees/' + id + '/remove');  
    },

    EnvDirs(params) {
        return fetch('/envs/dirs/list', params);
    },

    CreateEnvDir(params) {
        return fetch('/envs/dirs/create', params);
    },

    UpdateEnvDir(params) {
        return fetch('/envs/dirs/' + params.Id + '/update', params);
    },

    RemoveEnvDir(id) {
        return fetch('/envs/dirs/' + id + '/remove');
    },

    EnvValues(params) {
        return fetch('/envs/values/list', params);
    },

    EnvValue(id) {
        return fetch('/envs/values/' + id + '/detail');
    },

    CreateEnvValue(params) {
        return fetch('/envs/values/create', params);
    },

    UpdateEnvValue(params) {
        return fetch('/envs/values/' + params.Id + '/update', params);
    },

    RemoveEnvValue(id) {
        return fetch('/envs/values/' + id + '/remove');
    },

    UpdateEnvValues(params) {
        return fetch('/envs/values/update', params);
    },

    Teams(params) {
        return fetch('/teams/list', params);
    },

    Team(id) {
        return fetch('/teams/' + id + '/inspect');
    },

    CreateTeam(params) {
        return fetch('/teams/create', params);
    },

    RemoveTeam(id) {
        return fetch('/teams/' + id + '/remove');
    },

    UpdateTeam(params) {
        return fetch('/teams/' + params.Id + '/update', params);
    },

    AppointLeader(params) {
        return fetch('/teams/' + params.TeamId + '/appoint?UserId=' + params.UserId, params);
    },

    JoinTeam(params) {
        return fetch('/users/' + params.UserId + '/join?TeamId=' + params.TeamId, params);
    },

    QuitTeam(params) {
        return fetch('/users/' + params.UserId + '/quit?TeamId=' + params.TeamId, params);
    },

    Users(params) {
        return fetch('/users/list', params);
    },

    User(id) {
        return fetch('/users/' + id + "/detail");
    },

    CreateUser(params) {
        return fetch('/users/create', params);
    },
    
    RemoveUser(id) {
        return fetch('/users/' + id + '/remove');
    },

    UpdateUser(params) {
        return fetch('/users/' + params.Id + '/update', params);
    },

    ResetPassword(params) {
        return fetch('/users/' + params.Id + '/resetpass?Pass=' + encodeURIComponent(params.Pass), params);
    },

    Templates(params) {
        return fetch('/templates/list', params);
    },

    Template(id) {
        return fetch('/templates/' + id + '/detail');
    },

    CreateTemplate(params) {
        return fetch('/templates/create', params);
    },

    CopyTemplate(id, title) {
        let params = { Title: title };
        return fetch('/templates/' + id + '/copy', params);
    },

    UpdateTemplate(params) {
        return fetch('/templates/' + params.Id + '/update', params);
    },

    RemoveTemplate(id) {
        return fetch('/templates/' + id + '/remove');
    },

    Applications(params) {
        return fetch('/applications/list', params);
    },

    CreateApplication(params) {
        return fetch('/applications/create', params);
    },

    StartApplication(id) {
        return fetch('/applications/' + id + '/start');
    },

    StopApplication(id) {
        return fetch('/applications/' + id + '/stop');
    },

    RemoveApplication(id) {
        return fetch('/applications/' + id + '/remove');
    },

    Containers(params) {
        return fetch('/applications/' + params.Id + '/containers', params);
    },

    RestartContainer(id) {
        return fetch('/applications/containers/' + id + '/restart');
    },

    ScaleService(params) {
        return fetch('/applications/' + params.Id + '/scale', params);
    },

    DeploymentHistory(params) {
        return fetch('/applications/' + params.Id + '/history');
    },

    UpgradeApplication(params) {
        return fetch('/applications/' + params.Id + '/upgrade');
    },

    RollbackApplication(params) {
        return fetch('/applications/' + params.Id + '/rollback');
    },

    AddTeamToApplication(params) {
        return fetch('/applications/' + params.Id + '/add-team');
    },

    RemoveUserToApplication(params) {
        return fetch('/applications/' + params.Id + '/remove-team');
    },

    AddUserToApplication(params) {
        return fetch('/applications/' + params.Id + '/add-user');
    },

    RemoveUserToApplication(params) {
        return fetch('/applications/' + params.Id + '/remove-user');
    }

}
import axios from 'axios'
import qs from 'qs'

import * as ui from '../util/ui'

// axios默认配置
axios.defaults.timeout = 5000;
axios.defaults.baseURL = 'http://localhost:8080/public/mock';

export function fetch(url, params) {
    return new Promise((resolve, reject) => {
        axios.get(url, params)
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
        return fetch('/user/' + encodeURIComponent(params.Name) + '/login', params);
    },

    Pools(params) {
        return fetch('/pools/list', params);
    },

    Pool(id) {
        return fetch('/pools/' + id + '/detail');
    },

    CreatePool(params) {
        return fetch('/pools/create', params); 
    },

    RemovePool(params) {
        return fetch('/pools/' + params.Id + '/remove', params); 
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

    EnvDirs(params) {
        return fetch('/envs/dirs/list', params);
    },

    CreateEnvDir(params) {
        return fetch('/envs/dirs/create', params);
    },

    UpdateEnvDir(params) {
        return fetch('/envs/dirs/' + params.Id + '/update', params);
    },

    RemoveEnvDir(params) {
        return fetch('/envs/dirs/' + params.Id + '/remove', params);
    },

    EnvValues(params) {
        return fetch('/envs/values/list', params);
    },

    EnvValue(id) {
        return fetch('/envs/values/' + id + '/detail', params);
    },

    CreateEnvValue(params) {
        return fetch('/envs/values/create', params);
    },

    UpdateEnvValue(params) {
        return fetch('/envs/values/' + params.Id + '/update', params);
    },

    RemoveEnvValue(params) {
        return fetch('/envs/values/' + params.Id + '/remove', params);
    },

    UpdateEnvValues(params) {
        return fetch('/envs/values/update', params);
    },

    Teams(params) {
        return fetch('/teams/list', params);
    },

    Team(id) {
        return fetch('/teams/' + id + '/detail');
    },

    CreateTeam(params) {
        return fetch('/teams/create', params);
    },

    RemoveTeam(params) {
        return fetch('/teams/' + params.Id + '/remove', params);
    },

    UpdateTeam(params) {
        return fetch('/teams/' + params.Id + '/update', params);
    },

    AppointLeader(params) {
        return fetch('/teams/' + params.Id + '/appoint', params);
    },

    JoinTeam(params) {
        return fetch('/users/' + params.Id + '/join', params);
    },

    QuitTeam(params) {
        return fetch('/users/' + params.Id + '/quit', params);
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
    
    RemoveUser(params) {
        return fetch('/users/' + params.Id + '/remove', params);
    },

    UpdateUser(params) {
        return fetch('/users/' + params.Id + '/update', params);
    },

    ResetPassword(params) {
        return fetch('/users/' + params.Id + '/resetpass?Pass=' + encodeURIComponent(params.Pass), params);
    }

}